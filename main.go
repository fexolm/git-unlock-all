package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
)

type Owner struct {
	Name string `json:"name"`
}

type LockInfo struct {
	Id       string `json:"id"`
	Path     string `json:"path"`
	LockedAt string `json:"locked_at"`
	Owner    Owner  `json:"owner"`
}

type LocksInfo struct {
	Ours   []LockInfo `json:"ours"`
	Theirs []LockInfo `json:"theirs"`
}

func pushCmd() error {
	cmd := exec.Command("git", "push")

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return cmd.Run()
}

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()

	if errb.Len() > 0 {
		return "", errors.New(errb.String())
	}

	if err != nil {
		return "", err
	}

	return outb.String(), nil
}

func runLFSLocksCmd() (string, error) {
	return runCmd("git", "lfs", "locks", "--verify", "--json")
}

func runLFSUnlockCmd(path string, force bool) error {
	if force {
		_, err := runCmd("git", "lfs", "unlock", "--force", path)
		return err
	}
	_, err := runCmd("git", "lfs", "unlock", path)
	return err
}

func getLocks() (LocksInfo, error) {
	out, err := runLFSLocksCmd()
	if err != nil {
		return LocksInfo{}, err
	}

	res := LocksInfo{}

	err = json.Unmarshal([]byte(out), &res)

	if err != nil {
		return LocksInfo{}, err
	}

	return res, nil
}

func unlockAll(locks LocksInfo, force bool) error {
	for _, lock := range locks.Ours {
		err := runLFSUnlockCmd(lock.Path, force)
		if err != nil {
			log.Println(err)
		}
	}
	if force {
		for _, lock := range locks.Theirs {
			err := runLFSUnlockCmd(lock.Path, force)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}

func main() {
	force := flag.Bool("force", false, "force unlock")
	push := flag.Bool("push", false, "push before unlock")
	flag.Parse()

	if *push {
		err := pushCmd()
		if err != nil {
			log.Fatal(err)
		}
	}

	locks, err := getLocks()

	if err != nil {
		log.Fatal(err)
	}

	err = unlockAll(locks, *force)

	if err != nil {
		log.Fatal(err)
	}
}
