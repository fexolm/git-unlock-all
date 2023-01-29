package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
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
	Ours []LockInfo `json:"ours"`
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

func runLFSUnlockCmd(path string) error {
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

func unlockAll(locks LocksInfo) error {
	for _, lock := range locks.Ours {
		err := runLFSUnlockCmd(lock.Path)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func main() {
	locks, err := getLocks()

	if err != nil {
		log.Fatal(err)
	}

	err = unlockAll(locks)

	if err != nil {
		log.Fatal(err)
	}
}
