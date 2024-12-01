package main

import (
    "time"
)

func main() {
    retryInterval := time.Duration(2 * time.Minute)
    verifyInterval := time.Duration(30 * time.Second)
    RunApp(retryInterval, verifyInterval)
}

