package main

import (
	"bytes"
	"errors"
	"log"
	"net"
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
				return result
			}
		} else {
			// Stderr?
			log.Fatalf("cmd.Wait: %+v", err)
			result.Error = err
			return result
		}
	}
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
	snapshotName := strings.TrimSpace(
		strings.Replace(lines[1],
			"snapshot directory:",
			"",
			1))
	return snapshotName, nil
}

func ClearSnapshot(snapshotName string) error {
	result := SystemCall("nodetool", "clearsnapshot", snapshotName)
	if result.StatusCode != 0 {
		return errors.New("Clearing snapshot failed")
	}

	return nil
}

// GetNodeTokens ...
func GetNodeTokens() []string {
	addresses, err := GetInterfaceAddresses()
	if err != nil {
		return nil
	}

	tokens := []string{}
	result := SystemCall("nodetool", "ring")
	lines := strings.Split(string(result.Output), "\n")
	for _, v := range lines {
		for _, address := range addresses {
			if !strings.HasPrefix(v, address) {
				continue
			}

			// Assume the token is always first if reversed.
			values := strings.Split(strings.TrimSpace(v), " ")
			tokens = append(tokens, values[len(values)-1])
		}
	}

	return tokens
}

// GetInterfaceAddresses ...
// We're gettiing the interfaces...
func GetInterfaceAddresses() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("GetInterfacesAddresses: Failed to get interfaces: " +
			err.Error())
		return nil, err
	}

	ips := []string{}
	for _, v := range ifaces {
		if v.Flags&net.FlagUp == 0 || v.Flags&net.FlagLoopback != 0 {
			continue
		}

		addresses, err := v.Addrs()
		if err != nil {
			log.Println("GetInterfaceAddresses: Failed to Address for interface. ")
			continue
		}

		for _, address := range addresses {
			var ip net.IP
			switch addr := address.(type) {
			case *net.IPNet:
				ip = addr.IP
			case *net.IPAddr:
				ip = addr.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

// prettyPrint ...
func prettyPrint(in ...interface{}) {
	pretty.Println(in)
}

// SliceContainsString ...
func SliceContainsString(str string, slice []string) bool {
	for _, value := range slice {
		if value == str {
			return true
		}
	}

	return false
}

// SplitAndTrim ...
func SplitAndTrim(str string, sep string) []string {
	result := []string{}
	for _, element := range strings.Split(str, sep) {
		element = strings.TrimSpace(element)
		if len(element) == 0 {
			continue
		}

		result = append(result, element)
	}

	return result
}
