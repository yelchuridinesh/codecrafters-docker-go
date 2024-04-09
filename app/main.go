package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	//fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	//
	command := os.Args[3] //it takes the third argument as a command to execute /usr/local/bin/docker-explorer/
	// fmt.Println(command)
	// fmt.Println(len(os.Args)) 
	args := os.Args[4:len(os.Args)]  // from argument 4 until the end of arguments it take as  args
	fmt.Println(args)
	fmt.Println(args[0], args[1])
	// fmt.Println(len(os.Args))
	
	cmd := exec.Command(command, args...)
	fmt.Println(cmd)
	//this cmd has /usr/local/bin/docker-explorer/ echo hey
	
	cmd.Stderr=os.Stderr
	cmd.Stdout=os.Stdout
	err:=cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v", err)
		//os.Exit(1)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
