package main

import (
	"log"
	"os/exec"
	"runtime"
)

func main() {
	cmd := "go"
	buildArgs := []string{"build", "-o"}

	goos := runtime.GOOS
	if goos == "windows" {
		buildArgs = append(buildArgs, "./tmp/main.exe")
	} else {
		buildArgs = append(buildArgs, "./tmp/main")
	}

	buildArgs = append(buildArgs, "./cmd/app")

	buildCmd := exec.Command(cmd, buildArgs...)
	buildCmd.Stdout = log.Writer()
	buildCmd.Stderr = log.Writer()
	err := buildCmd.Run()
	if err != nil {
		panic(err)
	}
}
