package main

import (
	"time"
)

func main() {
	waitTime := time.Duration(5 * time.Second)
	retryTime := time.Duration(15 * time.Minute)
    AutomaticGitSync(waitTime, retryTime)
}
