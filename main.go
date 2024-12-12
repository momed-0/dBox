package main

import (
	"os"
	"os/exec"
	"syscall"
	"fmt"
	"time"
	"path"
	"log"
	"strconv"
)

const cGroupPath = "/sys/fs/cgroup/dBox.slice/"

func containerInit() {	
	
	//rerun the same process with new namespaces and a extra child argument
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	//setup the cgroup controllers
	cGroupInit()

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
				HostID:      1000, 
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1000,
				Size:        1,
			},
		},
		GidMappingsEnableSetgroups: false,
	}

	log.Printf("Trying to start the process in new namespace...")
	// Start the new process with a seperate uts namespace
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error starting container: %v", err)
	}
}

func child() { 
	// //adjust the PATH for binaries inside the container
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	
	// Set the hostname within the new namespace
	hostname := fmt.Sprintf("container-%s", time.Now().Format("20060102-150405"))

	log.Printf("Setting container hostname to: %s ...", hostname)
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		log.Fatalf("Error setting hostname: %v", err)
	}
	
	// Setup OverlayFS for the container
	if err := setupOverlayFS(hostname); err != nil {
		log.Fatalf("Error setting up overlayfs: %v", err)
	}
	defer teardownOverlayFS(hostname)

	rootDir := fmt.Sprintf("/tmp/%s/root", hostname)

	// change root directory
	log.Printf("Changing root of container to image....")
	if err := syscall.Chroot(rootDir) ; err != nil {
		log.Fatalf("Error changing the root directory: %v", err) 
	}
	//move to root directory
	if err := syscall.Chdir("/") ; err != nil {
		log.Fatalf("Error changing pwd to root '/': %v",err)
	}

	//mount /proc
	log.Printf("Mounting /proc file system..")
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		log.Fatalf("failed to mount /proc: %v", err)
	}

	if len(os.Args) > 2 {
		log.Printf("Executing the command: %s ", os.Args[2])
		cmd := exec.Command(os.Args[2],os.Args[3:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Execute the command
		if err := cmd.Start(); err != nil {
			log.Fatalf("Error running container: %v", err)
		}
		
		if err := cmd.Wait(); err != nil {
			log.Fatalf("Error running the conatiner: %v", err)
		}
	}
	//unmount proc after running
	log.Printf("starting to unmount /proc")
	if err := syscall.Unmount("proc", 0); err != nil {
		log.Fatalf("Failed to unmount /proc: %v", err)
	}
}


func main() {
	if len(os.Args) < 3 {
		log.Fatalf("Error! Usage ./main run [command] [options]")
	}

	switch os.Args[1] {
	case "run":
		containerInit()
	case "child":
		child()
	default:
		log.Fatalf("%s : command not supported.Correct usage ./main run [command] [options]",os.Args[1])
	}
}

func setupOverlayFS(containerName string) error {
	// Directories for OverlayFS layers
	lowerDir := "./root_alpine"                           // Original Alpine filesystem (read-only)
	upperDir := fmt.Sprintf("/tmp/%s/upper", containerName) // Writable layer
	workDir := fmt.Sprintf("/tmp/%s/work", containerName)   // Work directory
	rootDir := fmt.Sprintf("/tmp/%s/root", containerName) // Visible merged directory

	// Create directories
	log.Printf("Creating OverlayFS....")
	for _, dir := range []string{upperDir, workDir, rootDir} {
		log.Printf("Creating %q", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	// Mount OverlayFS
	log.Printf("Trying to mount overlayFS...")
	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)
	if err := syscall.Mount("overlay", rootDir, "overlay", 0, options); err != nil {
		return fmt.Errorf("failed to mount overlayfs: %v", err)
	}

	return nil
}

func teardownOverlayFS(containerName string) {
	log.Printf("Tearing down OverlayFS...")
	//write code to tear down OverlayFS
}

func cGroupInit() error{
	controller := path.Join(cGroupPath,"containers")
	log.Printf("Creating cgroup controller: %q", controller)
	
	if err := os.MkdirAll(controller, 0755) ;err != nil {
		log.Fatalf("Failed to create cgroup controller %q: %v", controller, err)
	}
	// allow cpu and memory controllers.
	writeToFile(path.Join(cGroupPath, "cgroup.subtree_control"), "+cpu +memory")

	log.Printf("Setting up cgroup rules...")

	writeToFile(path.Join(controller, "cpu.max"), "10000 100000")
	writeToFile(path.Join(controller, "memory.max"), "512M")
	writeToFile(path.Join(controller, "memory.swap.max"), "0")
	writeToFile(path.Join(controller, "cgroup.procs"), strconv.Itoa(os.Getpid()))
	return nil
}

func writeToFile(filename, message string) {
	err := os.WriteFile(filename, []byte(message), 0644)
	if err != nil {
		log.Fatalf("failed to write to file: %v", err)
	}
}