package main

import (
	"os"
	"os/exec"
	"syscall"
	"strings"
	"fmt"
	"time"
)


func runCommand() {	

	// Create a hostname based on the current date
	currentDate := time.Now().Format("20060102-150405") // Format: YYYYMMDD-HHMMSS
	hostname := fmt.Sprintf("container-%s", currentDate)


	// Construct the command string: "hostname <date-based-hostname> && exec <command> <args>"
	cmdString := fmt.Sprintf("hostname %s && exec %s", hostname, strings.Join(os.Args[2:], " "))

	cmd := exec.Command("/bin/sh", "-c", cmdString)


	// //connect the standard input out error
	cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	// Create a new UTS namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS,
	}

	// Start the new process with a seperate uts namespace
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}

	// Wait for the process to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Give me an argument !")
	}

	switch os.Args[1] {
	case "run":
		runCommand()
	default:
		fmt.Println("Command not supported")
	}
}
