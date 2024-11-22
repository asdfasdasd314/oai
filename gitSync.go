package main

import (
    "time"
    "fmt"
    "strconv"
    "os/exec"
    "github.com/gen2brain/beeep"
)

func automaticGitSync(syncInterval time.Duration, waitCheckInterval time.Duration) {
    // This one automatically syncs on the specified interval
    canExit := make(chan bool, 1)
    // By default either one can exit
    canExit <- true

    nextSyncTime := time.Now().Add(syncInterval)

    go func() {
        for {
            // There could theoretically be an issue in that the user may exit while the sync is happening
            // We don't want that to happen, so here we can use channels to pass messages about the completion status of each goroutine
            for time.Now().UnixMicro() < nextSyncTime.UnixMicro() {
                time.Sleep(waitCheckInterval)
            }

            // Receive from the channel so the main goroutine must stop
            <-canExit
            
            // If we've gotten to this point we need to guarantee that we can run the git commands
            fmt.Println("Syncing data automatically...")
            //runGitSyncCommands()
            notifySuccess()
            fmt.Println("Successfully synced data! | " + formatTime(time.Now()))

            // Send to the channel so now the main goroutine can exit if it wants to
            canExit <- true

            nextSyncTime = time.Now().Add(syncInterval)
        }
    }()

    // This syncs on user input
    for {
        var input string
        fmt.Scanln(&input)
        switch input {
            case "sync":
                fmt.Println("Syncing data...")
                //runGitSyncCommands()
                fmt.Println("Successfully synced data! | " + formatTime(time.Now()))
            case "next-sync-time":
                fmt.Println(formatTime(nextSyncTime))
            case "time-until-sync":
                fmt.Println(getTimeUntilSync(nextSyncTime))
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

    commitMessage := "Committed changes up to " + formatTime(time.Now())
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

func formatTime(time time.Time) string {
    return string(time.Format("Mon Jan 2 2006 15:03:02 MST"))
}

func getTimeUntilSync(syncTime time.Time) string {
    differenceInTime := syncTime.Unix() - time.Now().Unix()
    
    days := differenceInTime / (60 * 60 * 24)
    differenceInTime -= (days * 60 * 60 * 24)
    hours := differenceInTime / (60 * 60)
    differenceInTime -= (hours * 60 * 60)
    minutes := differenceInTime / 60
    differenceInTime -= (minutes * 60)
    seconds := differenceInTime
    
    output := ""
    if days > 0 {
        output += strconv.FormatInt(days, 10)
        output += " days"
    }
    if hours > 0 {
        if len(output) > 0 {
            output += " "
        }
        output += strconv.FormatInt(hours, 10)
        output += " hours"
    }
    if minutes > 0 {
        if len(output) > 0 {
            output += " "
        }
        output += strconv.FormatInt(minutes, 10)
        output += " minutes"
    }
    if seconds > 0 {
        if len(output) > 0 {
            output += " "
        }
        output += strconv.FormatInt(seconds, 10)
        output += " seconds"
    }

    return output
}
