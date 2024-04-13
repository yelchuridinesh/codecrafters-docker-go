//go:build linux
// +build linux

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	image := os.Args[2]
	imageDir := fmt.Sprintf("./images/%s", image)

	if _, err := os.Stat(imageDir); err != nil {
		if os.IsNotExist(err) {
			imageDir, err = ImagePull(image, "./images")
			if err != nil {
				logError(err, "Error while pulling the image")
				os.Exit(255)
			}
		}
	}

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// change root filesystem for the child process using chroot
	// this is necessary to make the child process believe it is running in a different root filesystem
	// not needed during final task since we are downloading our image
	IsolatedProcess()

	cmd := exec.Command(command, args...)

	//It will create a process Isolation by creating a new Namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
		Chroot:     imageDir, //changed the
		// this CLONE_NEWPID Unshare the PID namespace, so that the calling
		// process has a new PID namespace for its children which is
		// not shared with any previously existing process.  The
		// calling process is not moved into the new namespace.  The
		// first child created by the calling process will have the
		// process ID 1 and will assume the role of init(1) in the
		// new namespace
	}

	// bind the standard input, output and error to the parent process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	// exit with the same exit code as the child process
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
	}
}

func IsolatedProcess() {
	rootFsPath, err := os.MkdirTemp("", "temp_")
	if err != nil {
		logError(err, "Failed to create temporary directory")
	}
	err = os.Chmod(rootFsPath, 0755)
	if err != nil {
		logError(err, "Failed to change permissions of temporary directory")
	}

	defer os.Remove(rootFsPath)

	binPath := "/usr/local/bin"

	err = os.MkdirAll(path.Join(rootFsPath, binPath), 0755)
	if err != nil {
		logError(err, "Failed to create bin directory")
	}

	//Link command is used to link the existing path with the new path in this case /tmp/temp_*/usr/local/bin/docker-explorer
	os.Link("/usr/local/bin/docker-explorer", path.Join(rootFsPath, "/usr/local/bin/docker-explorer"))
	if err != nil {
		logError(err, "Failed to copy binaries to root file system")
	}

	err = syscall.Chroot(rootFsPath)
	if err != nil {
		logError(err, "Failed to change root filesystem")
	}
}

func logError(err error, errorMessage string) {
	log.Fatalf("%s: %v", errorMessage, err)
	os.Exit(1)
}

// func main() {

// 	command := os.Args[3]           //it takes the third argument as a command to execute /usr/local/bin/docker-explorer/
// 	args := os.Args[4:len(os.Args)] // from argument 4 until the end of arguments it take as  args

// 	//chroot is isolated a process from  a main root process.
// 	//we can create temp directory and copy the binary of the process
// 	//we need to execute to that chroot directory and change current root to the temporary root

// 	chrootPath := path.Join(os.TempDir(), fmt.Sprintf("%d", os.Getpid())) //it gives a random temp directory by attaching a pid
// 	chrootCommand := path.Join(chrootPath, command)                       // for that temp directory it adds the command so /temp/pid//usr/local/bin/docker-explorer
// 	err := os.MkdirAll(chrootPath, 0755)                                  // creates the directory with the details /temp/<pid_number>/
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "temp dir err: %v", err)
// 		os.Exit(1)
// 	}
// 	commandFile, err := os.ReadFile(command) //it reads file in our case binary details of /usr/local/bin/docker-explorer

// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "reading command err: %v", err)
// 		os.Exit(1)
// 	}
// 	// fmt.Println(path.Dir(chrootCommand))
// 	err = os.MkdirAll(path.Dir(chrootCommand), 0755) //it creates /user/local/bin inside /temp/<pid_number>/
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "temp two command err: %v", err)
// 		os.Exit(1)
// 	}

// 	result, err := os.Create(chrootCommand)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "command file create err: %v", err)
// 		os.Exit(1)
// 	}
// 	os.Chmod(chrootCommand, 0777)
// 	_, err = result.Write(commandFile) //write the details of binary into the file we created
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "command write err: %v", err)
// 		os.Exit(1)
// 	}
// 	result.Close()

// 	// if os.Args[0] == "exit" {
// 	// 	exCode, _ := strconv.Atoi(os.Args[1])
// 	// 	os.Exit(exCode)
// 	// }

// 	cmd := exec.Command(command, args...)
// 	//this cmd has /usr/local/bin/docker-explorer/ echo hey

// 	cmd.Stderr = os.Stderr
// 	cmd.Stdout = os.Stdout
// 	// cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: chrootPath, Cloneflags: syscall.CLONE_NEWPID} // this is for changing the Chroot path from / to temp directory
// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		Chroot:     chrootPath,
// 		Cloneflags: syscall.CLONE_NEWPID,
// 		// this CLONE_NEWPID Unshare the PID namespace, so that the calling
// 		// process has a new PID namespace for its children which is
// 		// not shared with any previously existing process.  The
// 		// calling process is not moved into the new namespace.  The
// 		// first child created by the calling process will have the
// 		// process ID 1 and will assume the role of init(1) in the
// 		// new namespace
// 	}
// 	err = cmd.Run()
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "err: %v", err)
// 		//os.Exit(1)
// 		os.Exit(cmd.ProcessState.ExitCode()) //--> this is alternate way of existing the process with the same exit code as argument
// 	}
// }
