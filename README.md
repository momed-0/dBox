# dBox

`dBox` is a lightweight container runtime written in Go. It demonstrates the core principles of containerization by utilizing Linux namespaces such as UTS, PID, and filesystem namespaces. It allows users to create isolated environments for executing commands, mimicking basic container behavior.

## Features
- Custom container runtime with UTS and PID namespaces.
- Ability to run commands inside isolated environments.
- Utilizes `chroot` and mounts `proc` to simulate a minimal container.
  
## Requirements
- Go (version 1.16 or later)
- Alpine Linux root filesystem
- Linux-based operating system

## Setup

### 1. Install Go

Ensure you have Go installed on your machine. You can download it from the official [Go website](https://golang.org/dl/).

### 2. Clone the Repository

```bash
git clone https://github.com/yourusername/dBox.git
cd dBox  bas
```
## 3. Download Alpine Linux Filesystem

To test the early versions of this container runtime, you will need to download the Alpine Linux filesystem:

1. Download an appropriate version of Alpine Linux from [Alpine's official site](https://www.alpinelinux.org/downloads/).

2. Extract the Alpine root filesystem in your project directory:

    ```bash
    mkdir root_alpine
    sudo tar -xvpf alpine-minirootfs-<version>-x86_64.tar.gz -C root_alpine
    ```

3. Create an empty file named `ALPINE_FS_ROOT` in the root of the Alpine filesystem:

    ```bash
    touch root_alpine/ALPINE_FS_ROOT
    ```

4. Add `root_alpine` to your `.gitignore` to prevent it from being pushed to the remote repository.


## 4. Build the dBox CLI

To compile the `dBox` CLI program, run:

```bash
go build -o dBox main.go
```

This will produce an executable named `dBox` in the project directory.

## Usage

Once compiled, `dBox` can be used to run commands in an isolated environment.

### Running a Command in dBox

```bash
./dBox run <command> [args...]
```

For example, to run a shell inside the isolated environment:
```bash
./dBox run /bin/sh
```

Inside the isolated shell, you can run commands like ps, ls, or any other installed commands in the Alpine filesystem.

## License

This project is licensed under the GNU General Public License v3.0. See the LICENSE file for more details.
