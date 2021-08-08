package main

import (
	"os"
	"os/exec"
)

func fatal(message string) {
	println(message)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		println("Usage: webview2generator <file.idl>")
		os.Exit(1)
	}

	targetDir, err := os.Getwd()
	if err != nil {
		fatal(err.Error())
	}

	err = ParseIDL(os.Args[0], targetDir)
	if err != nil {
		fatal(err.Error())
	}
	err = os.Chdir("test")
	if err != nil {
		fatal(err.Error())
	}
	command := exec.Command("go", "fmt", "./...")
	err = command.Run()
	if err != nil {
		fatal(err.Error())
	}
}
