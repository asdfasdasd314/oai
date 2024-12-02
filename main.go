package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// Then we have command line args passed into the program
	inDebugMode := false
	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if arg == "debug" {
				fmt.Println("Running in debug mode")
				inDebugMode = true
			}
		}
	}
	retryInterval := time.Duration(2 * time.Minute)
	verifyInterval := time.Duration(1 * time.Second)
	RunApp(retryInterval, verifyInterval, inDebugMode)
}
