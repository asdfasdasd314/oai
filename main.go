package main

import (
    "time"
)

func main() {
    duration := time.Duration(24 * time.Hour)
    automaticGitSync(duration)
}

