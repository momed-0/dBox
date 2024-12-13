# dBox

`dBox` is a lightweight container runtime written in Go that demonstrates core containerization principles using Linux namespaces such as UTS, PID, filesystem namespaces, cgroups, and union file systems. It provides users with isolated environments to execute commands, mimicking basic container behavior. To prevent resource exploitation, additional cgroup controllers are also implemented. Users can pull container images from Docker Hub's official library and recreate the layers.

## Features
- Custom container runtime with namespaces, cgroups and union file systems.
- Execute commands inside isolated environments.
- Pull and run images from Docker Hub's official library.

  
## Requirements
- Go (version 1.16 or later)
- Linux-based operating system
- root user permission


## Setup

### 1. Install Go

Ensure you have Go installed on your machine. You can download it from the official [Go website](https://golang.org/dl/).

### 2. Clone the Repository
Clone the repository and navigate to the project directory:
```bash
git clone https://github.com/momed-0/dBox
cd dBox
```
### 3. Create cgroup controller for dBox
To manage resources, create a cgroup controller in `/sys/fs/cgroup/` and configure the `cgroup.subtree_control` file
```bash
mkdir /sys/fs/cgroup/dBox.slice

ls /sys/fs/cgroup/dBox.slice

echo "+cpu +memory -io" > /sys/fs/cgroup/<parent>/cgroup.subtree_control
```
## 4. Build the dBox CLI

To compile the `dBox` CLI program, run:

```bash
go build -o dBox cmd/main.go
```

This will produce an executable named `dBox` in the project directory.

## Usage

Once compiled, `dBox` can be used to run commands in an isolated environment.

### Running a Command in dBox

To run a command in a containerized environment:

```bash
sudo ./dBox run <image_name> <command> [args...]
```
- Root permissions are required to start the container.
- Inside the container, commands are executed as a non-root user.

For example, to run a shell inside the isolated environment:

```bash
sudo ./dBox run ubuntu /bin/bash
```
Inside the isolated shell, you can run commands like ps, ls, or any other installed commands.

To pull an image that is available in the official docker library. ( https://hub.docker.com/search?badges=official ) use the following command:

sudo ./dBox pull [image_name] 
For example, to pull the Arch Linux image:

```bash
sudo ./dBox pull archlinux
```

## License

This project is licensed under the GNU General Public License v3.0. See the LICENSE file for more details.
