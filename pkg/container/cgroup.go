package container

import (
	"log"
	"os"
	"path"
	"strconv"
	"dBox/pkg/utils"
)

const cGroupPath = "/sys/fs/cgroup/dBox.slice/"

// cGroupInit initializes the cgroup for resource management in containers
func cGroupInit() error {
	// Define the controller directory for the container
	controller := path.Join(cGroupPath, "containers")
	log.Printf("Creating cgroup controller: %q", controller)

	// Create the cgroup directory if it doesn't exist
	if err := os.MkdirAll(controller, 0755); err != nil {
		log.Fatalf("Failed to create cgroup controller %q: %v", controller, err)
		return err
	}

	// Enable CPU and memory controllers
	utils.WriteToFile(path.Join(cGroupPath, "cgroup.subtree_control"), "+cpu +memory")

	// Set the cgroup rules (e.g., CPU and memory limits)
	log.Printf("Setting up cgroup rules...")
	utils.WriteToFile(path.Join(controller, "cpu.max"), "10000 100000")
	utils.WriteToFile(path.Join(controller, "memory.max"), "512M")
	utils.WriteToFile(path.Join(controller, "memory.swap.max"), "0")

	// Add the current process (container) to the cgroup
	utils.WriteToFile(path.Join(controller, "cgroup.procs"), strconv.Itoa(os.Getpid()))

	return nil
}
