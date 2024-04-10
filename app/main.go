package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	//fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	//
	command := os.Args[3]           //it takes the third argument as a command to execute /usr/local/bin/docker-explorer/
	args := os.Args[4:len(os.Args)] // from argument 4 until the end of arguments it take as  args

	//chroot is isolated a process from  a main root process.
	//we can create temp directory and copy the binary of the process
	//we need to execute to that chroot directory and change current root to the temporary root

	chrootPath := path.Join(os.TempDir(), fmt.Sprintf("%d", os.Getpid())) //it gives a random temp directory by attaching a pid
	chrootCommand := path.Join(chrootPath, command)                       // for that temp directory it adds the command so /temp/pid//usr/local/bin/docker-explorer
	err := os.MkdirAll(chrootPath, 0755)                                  // creates the directory with the details /temp/<pid_number>/
	if err != nil {
		fmt.Fprintf(os.Stderr, "temp dir err: %v", err)
		os.Exit(1)
	}
	commandFile, err := os.ReadFile(command) //it reads file in our case binary details of /usr/local/bin/docker-explorer

	if err != nil {
		fmt.Fprintf(os.Stderr, "reading command err: %v", err)
		os.Exit(1)
	}
	// fmt.Println(path.Dir(chrootCommand))
	err = os.MkdirAll(path.Dir(chrootCommand), 0755) //it creates /user/local/bin inside /temp/<pid_number>/
	if err != nil {
		fmt.Fprintf(os.Stderr, "temp two command err: %v", err)
		os.Exit(1)
	}

	result, err := os.Create(chrootCommand)
	if err != nil {
		fmt.Fprintf(os.Stderr, "command file create err: %v", err)
		os.Exit(1)
	}
	os.Chmod(chrootCommand, 0777)
	_, err = result.Write(commandFile) //write the details of binary into the file we created
	if err != nil {
		fmt.Fprintf(os.Stderr, "command write err: %v", err)
		os.Exit(1)
	}
	result.Close()

	if os.Args[0] == "exit" {
		exCode, _ := strconv.Atoi(os.Args[1])
		os.Exit(exCode)
	}

	cmd := exec.Command(command, args...)
	//this cmd has /usr/local/bin/docker-explorer/ echo hey

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: chrootPath} // this is for changing the Chroot path from / to temp directory
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v", err)
		//os.Exit(1)
		//os.Exit(cmd.ProcessState.ExitCode()) --> this is alternate way of existing the process with the same exit code as argument
	}
}
