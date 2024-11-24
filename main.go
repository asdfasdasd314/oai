package main

import (
	"obsidianautomation/git_sync"
	"time"
)

func main() {
	sync := time.Duration(24 * time.Hour)
	waitTime := time.Duration(5 * time.Second)
	retryTime := time.Duration(15 * time.Minute)
	git_sync.AutomaticGitSync(sync, waitTime, retryTime)
}
