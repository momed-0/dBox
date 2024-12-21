# dBox

`dBox` is a lightweight container runtime written in Go that showcases core containerization principles using Linux namespaces such as UTS, PID, filesystem namespaces, cgroups, and union file systems. It provides isolated environments for executing commands, simulating basic container behavior. To prevent resource misuse, cgroup controllers are implemented. Users can pull container images from Docker Hub's official library and recreate the image layers.

---

## Features
- **Custom container runtime**: Built with namespaces, cgroups, and union file systems.
- **Isolated environments**: Execute commands in a secure, containerized space.
- **Image management**: Pull and run container images from Docker Hub's official library.

---

## Requirements
- **Go** (version 1.16 or later)
- **Linux-based OS**
- **Root user permissions**

---

## Setup

### 1. Install Go
Download and install Go from the official [Go website](https://golang.org/dl/).

### 2. Clone the Repository
Clone the repository and navigate to the project directory:
```bash
git clone https://github.com/momed-0/dBox
cd dBox
```

### 3. Configure cgroup Controller
To manage resources, set up a cgroup controller in `/sys/fs/cgroup/` and configure the `cgroup.subtree_control` file:
```bash
mkdir /sys/fs/cgroup/dBox.slice
ls /sys/fs/cgroup/dBox.slice
echo "+cpu +memory -io" > /sys/fs/cgroup/<parent>/cgroup.subtree_control
```

### 4. Build the dBox CLI
Compile the `dBox` CLI program:
```bash
go build -o dBox cmd/main.go
```
This will create an executable named `dBox` in the project directory.

---

## Usage

Once compiled, `dBox` allows you to execute commands in an isolated container environment.

### Running a Command
To run a command in a containerized environment:
```bash
sudo ./dBox run <image_name> <command> [args...]
sudo ./dBox run <image_name>:<tag (optional)> <command> [args...]
```
- Root permissions are required.
- Commands inside the container are executed as a non-root user.

#### Example:
Run a shell in an isolated environment (assuming the Ubuntu image is already pulled):
```bash
sudo ./dBox run ubuntu /bin/bash
sudo ./dBox run ubuntu:20.04 /bin/bash
```

### Pulling an Image
To pull an image from the official Docker library ([Docker Official Images](https://hub.docker.com/search?badges=official)):
```bash
sudo ./dBox pull <image_name>
```
#### Example:
```bash
sudo ./dBox pull archlinux
sudo ./dBox pull archlinux:latest
sudo ./dBox pull ubuntu:20.04
```

### Listing Local Images
To list all locally stored images:
```bash
sudo ./dBox images
```

---

## TODO
- Data Volume Containers
- Data Volumes
- Automatically clean up temporary files (currently reset on reboot).
- Implement a single command to pull and run an image.
- Add networking support using the network namespace.
- Provide an option to remove locally stored images.

---

## License
This project is licensed under the **GNU General Public License v3.0**. See the LICENSE file for more details.
