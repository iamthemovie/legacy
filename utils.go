package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kr/pretty"
)

// SystemCallResult ...
type SystemCallResult struct {
	Output     []byte
	StatusCode int
	Error      error
}

func SystemCall(name string, args ...string) SystemCallResult {
	command := exec.Command(name, args...)
	result := SystemCallResult{}
	var output bytes.Buffer
	command.Stdout = &output
	if err := command.Start(); err != nil {
		log.Fatalf("cmd.Start: %+v", err)
	}

	if err := command.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// If we're okay
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				result.StatusCode = status.ExitStatus()
				result.Output = output.Bytes()
				fmt.Printf("Exit Status: %d", result.StatusCode)
				return result
			}
		} else {
			// Stderr?
			log.Fatalf("cmd.Wait: %+v", err)
			result.Error = err
			return result
		}
	}

	log.Printf("%+v", string(output.Bytes()))
	result.Output = output.Bytes()
	result.StatusCode = 0
	return result
}

func CreateNewSnapshot(snapshotTag string) (string, error) {
	if len(snapshotTag) == 0 {
		snapshotTag = strconv.Itoa(int(time.Now().Unix()))
	}

	tag := "--tag legacy-" + snapshotTag
	result := SystemCall("nodetool", "snapshot", tag)
	if result.StatusCode != 0 {
		return "", errors.New("Snapshotting failed")
	}

	lines := strings.Split(strings.ToLower(string(result.Output)), "\n")
	snapshotName := strings.TrimSpace(strings.Replace(lines[1], "snapshot directory:", "", 1))
	return snapshotName, nil
}

func prettyPrint(in ...interface{}) {
	pretty.Println(in)
}
