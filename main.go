package main

import (
	"os"
	"os/exec"
	"syscall"
	"fmt"
	"time"
)


func containerInit() {	
	
	//rerun the same process with new namespaces and a extra child argument
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	//connect the standard input output error
	cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	//create a new UTS , PID, user namespace , mount namespace using clone syscall
	//basically spans a new process inside a new set of namespaces
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID  | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
            {
                ContainerID: 0, //user outside container marked as root inside
                HostID:      os.Getuid(),
                Size:        1,
            },
        },
		GidMappings: []syscall.SysProcIDMap{
            {
                ContainerID: 0,
                HostID:      os.Getgid(),
                Size:        1,
            },
        },
	}

	// Start the new process with a seperate uts namespace
	if err := cmd.Run(); err != nil {
		fmt.Println("Error starting container: ", err)
		os.Exit(1)
	}
}

func child() { 
	//adjust the PATH for binaries inside the container
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")

	// Set the hostname within the new namespace
	hostname := fmt.Sprintf("container-%s", time.Now().Format("20060102-150405"))

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		fmt.Println("Error setting hostname: ", err)
		os.Exit(1)
	}

	// change root directory
	if err := syscall.Chroot("./root_alpine") ; err != nil {
		fmt.Println("Error changing the root directory: ", err) 
		os.Exit(1)
	}
	//move to root directory
	if err := syscall.Chdir("/") ; err != nil {
		fmt.Println("Error changing pwd to root '/': ")
		os.Exit(1)
	}

	//mount /proc
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		fmt.Println("failed to mount /proc: ", err)
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		cmd := exec.Command(os.Args[2],os.Args[3:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Execute the command
		if err := cmd.Start(); err != nil {
			fmt.Println("Error running container: ", err)
			os.Exit(1)
		}

		
		if err := cmd.Wait(); err != nil {
			fmt.Printf("Error running the conatiner: ", err)
			os.Exit(1)
		}
	}
	//unmount proc after running
	if err := syscall.Unmount("proc", 0); err != nil {
		fmt.Println("Failed to unmount /proc: ", err)
		os.Exit(1)
	}
}


func main() {
	if len(os.Args) < 3 {
		fmt.Println("Give me an argument !")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		containerInit()
	case "child":
		child()
	default:
		fmt.Println("Command not supported")
		os.Exit(1)
	}
}
