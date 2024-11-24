package main

import (
	"obsidianautomation/git_sync"
	"time"
)

func main() {
	waitTime := time.Duration(5 * time.Second)
	retryTime := time.Duration(15 * time.Minute)
	git_sync.AutomaticGitSync(waitTime, retryTime)
}
