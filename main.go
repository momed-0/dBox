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

	// //connect the standard input output error
	cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	//create a new UTS , PID, user namespace , mount namespace using clone syscall
	//basically spans a new process inside a new set of namespaces
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID  | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		Credential: &syscall.Credential{
			Uid: 0,
			Gid: 0,
		},
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
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
	// //adjust the PATH for binaries inside the container
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	
	// Set the hostname within the new namespace
	hostname := fmt.Sprintf("container-%s", time.Now().Format("20060102-150405"))

	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		fmt.Println("Error setting hostname: ", err)
		os.Exit(1)
	}
	

	// Setup OverlayFS for the container
	if err := setupOverlayFS(hostname); err != nil {
		fmt.Println("Error setting up overlayfs: ", err)
		os.Exit(1)
	}
	defer teardownOverlayFS(hostname)

	mergedDir := fmt.Sprintf("/tmp/%s/merged", hostname)

	// change root directory
	if err := syscall.Chroot(mergedDir) ; err != nil {
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

func setupOverlayFS(containerName string) error {
	// Directories for OverlayFS layers
	lowerDir := "./root_alpine"                           // Original Alpine filesystem (read-only)
	upperDir := fmt.Sprintf("/tmp/%s/upper", containerName) // Writable layer
	workDir := fmt.Sprintf("/tmp/%s/work", containerName)   // Work directory
	mergedDir := fmt.Sprintf("/tmp/%s/merged", containerName) // Visible merged directory

	// Create directories
	for _, dir := range []string{upperDir, workDir, mergedDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	// Mount OverlayFS
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	if err := syscall.Mount("overlay", mergedDir, "overlay", 0, options); err != nil {
		return fmt.Errorf("failed to mount overlayfs: %v", err)
	}

	return nil
}

func teardownOverlayFS(containerName string) error {
	mergedDir := fmt.Sprintf("/tmp/%s/merged", containerName)
	containerRoot := fmt.Sprintf("/tmp/%s", containerName)    // Root directory of the container
	// Unmount merged directory
	if err := syscall.Unmount(mergedDir, 0); err != nil {
		return fmt.Errorf("failed to unmount overlayfs: %v", err)
	}
	if err := os.RemoveAll(containerRoot); err != nil {
		fmt.Errorf("Failed to remove container's tmp directory: %v", err)
	}

	return nil
}

