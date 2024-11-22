package main

import (
    "time"
    "fmt"
    "os/exec"
    "github.com/gen2brain/beeep"
)

func automaticGitSync(syncInterval time.Duration) {
    // This one automatically syncs on the specified interval
    canExit := make(chan bool, 1)
    // By default either one can exit
    canExit <- true

    go func() {
        for {
            // There could theoretically be an issue in that the user may exit while the sync is happening
            // We don't want that to happen, so here we can use channels to pass messages about the completion status of each goroutine
            time.Sleep(syncInterval)

            // Receive from the channel so the main goroutine must stop
            <-canExit
            
            // If we've gotten to this point we need to guarantee that we can run the git commands
            fmt.Println("Syncing data automatically...")
            runGitSyncCommands()
            notifySuccess()
            fmt.Println("Successfully synced data! | " + getFormattedTime())

            // Send to the channel so now the main goroutine can exit if it wants to
            canExit <- true
        }
    }()

    // This syncs on user input
    for {
        var input string
        fmt.Scanln(&input)
        switch input {
            case "sync":
                fmt.Println("Syncing data...")
                runGitSyncCommands()
                fmt.Println("Successfully synced data! | " + getFormattedTime())
            case "exit":
                // Wait to receive from the channel
                // This waits until there's something in the channel, which always before or after the automatic git sync commands are ran
                <-canExit
                return
        }
    }
}

func runGitSyncCommands() {
    cmd := exec.Command("git", "add", ".")
    err := cmd.Run()

    if err != nil {
        panic(err)
    }

    commitMessage := "Committed changes up to " + getFormattedTime()
    cmd = exec.Command("git", "commit", "-m", commitMessage)
    err = cmd.Run()
    if err != nil {
        panic(err)
    }

    cmd = exec.Command("git", "push", "-u", "origin")
    err = cmd.Run()
    if err != nil {
        panic(err)
    }
}

func notifySuccess() {
    err := beeep.Notify("Succesfully Synced with GitHub", "Success", "")
    if err != nil {
        panic(err)
    }
}

func getFormattedTime() string {
    now := time.Now()
    return string(now.Format("Mon Jan 2 2006 15:03:02 MST"))
}
