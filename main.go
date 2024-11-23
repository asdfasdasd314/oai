package main

import (
	"time"
)

func main() {
	sync := time.Duration(24 * time.Hour)
	waitTime := time.Duration(5 * time.Second)
	retryTime := time.Duration(15 * time.Minute)
	automaticGitSync(sync, waitTime, retryTime)
}
