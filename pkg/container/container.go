package container

import (
	"fmt"
	"os"
	"os/exec"
	"log"
	"syscall"
	"time"
	"dBox/pkg/filesystem"
)


func ContainerInit(imageData []string) {
	//remove the image name from the command and execute it again
	argArray := append(os.Args[:2], os.Args[3:]...)
	cmd := exec.Command(argArray[0], argArray[1:]...)
	
	cGroupInit()
	imageName := "IMAGE_NAME=" + imageData[0]
	imageTag := "IMAGE_TAG=" + imageData[1]
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PROCESS=CHILD")
	cmd.Env = append(cmd.Env ,imageName)
	cmd.Env = append(cmd.Env, imageTag)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		Credential: &syscall.Credential{Uid: 0, Gid: 0},
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
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error starting container: %v", err)
	}
}

func Child(imageName,tag string) {
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	hostname := fmt.Sprintf("container-%s", time.Now().Format("20060102-150405"))

	log.Printf("Setting container hostname to: %s ...", hostname)
	if err := syscall.Sethostname([]byte(hostname)); err != nil {
		log.Fatalf("Error setting hostname: %v", err)
	}

	if err := filesystem.SetupOverlayFS(hostname,imageName,tag); err != nil {
		log.Fatalf("Error setting up overlayfs: %v", err)
	}
	defer filesystem.TeardownOverlayFS(hostname)

	rootDir := fmt.Sprintf("/tmp/%s/root", hostname)
	log.Printf("Changing root of container to image....")
	if err := syscall.Chroot(rootDir); err != nil {
		log.Fatalf("Error changing the root directory: %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		log.Fatalf("Error changing pwd to root '/': %v", err)
	}

	log.Printf("Mounting /proc file system..")
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		log.Print("failed to mount /proc: %v", err)
	}

	if len(os.Args) > 1 {
		log.Printf("Executing the command: %s", os.Args[2])
		cmd := exec.Command(os.Args[2], os.Args[3:]...) // os.Args[3] is the command, os.Args[4:] are its arguments
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Fatalf("Error running container: %v", err)
		}

		if err := cmd.Wait(); err != nil {
			log.Fatalf("Error running the container: %v", err)
		}
	}

	log.Printf("Unmounting /proc file system..")
	if err := syscall.Unmount("proc", 0); err != nil {
		log.Printf("Failed to unmount /proc: %v", err)
	}
}
