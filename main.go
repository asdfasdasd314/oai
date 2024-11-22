package main

import (
    "time"
)

func main() {
    duration := time.Duration(24 * time.Hour)
    waitTime := time.Duration(5 * time.Second)
    automaticGitSync(duration, waitTime)
}
