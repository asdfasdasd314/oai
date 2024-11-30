package main

import (
    "fmt"
)

func printHelpCommands() {
    fmt.Println("Meta Commands")
    fmt.Println("   help: Lists commands that do stuff")
    fmt.Println("   exit: Exits the program safely without potentially being in the middle of a syncing command")
    fmt.Println()

    fmt.Println("Syncing")
    fmt.Println("   sync: Syncs with GitHub")
    fmt.Println("   skip-sync: Skips the next sync that would happen")
    fmt.Println()

    fmt.Println("Querying Data")
    fmt.Println("   next-sync-time: Gets the next time GitHub will automatically sync")
    fmt.Println("   time-until-sync: Calculates the time until the next sync in days, hours, minutes, and seconds")
    fmt.Println("   list-recurrent-times: Lists all the recurrent times GitHub is synced")
    fmt.Println()

    fmt.Println("Mutating Internal Data")
    fmt.Println("   set-sync-time: Sets a recurrent time at which the client will sync with GitHub given some recurring basis of days")
    fmt.Println("   erase-time: Removes a time for which GitHub was supposed to sync")
    fmt.Println()

    fmt.Println("Note Mutating Functions (**BE CAREFUL AS THESE INTERACT WITH YOU'RE ACTUAL NOTES**)")
    fmt.Println("   clean-completed-tasks: Clears completed tasks throughout the entire vault using a recursive function")
    fmt.Println()
}

// Runs all app logic to not clutter main file
func RunApp() {
    // Run the blocking code that the user interacts with
    for {
        var input string
        fmt.Scanln(&input)

        // Basically all of these cases should call some separate function
        switch input {
        case "help":
            printHelpCommands()
        case "exit":
            break;
        case "sync":
            runGitSyncCommands()
        case "nextsynctime":
            getNextSyncTime()
        case "timeuntilsync":
            getTimeUntilSync()
        case "listsynctimes":
            printSyncTimes()
        case "setsynctime":
            syncTime := getSyncTime()
            setSyncTime(syncTime)
        case "erasetime":
            syncTime := getSyncTime()
            eraseTime(syncTime)
        case "clearcompletedtasks":
            ClearCompletedTasks(".")
        }
    }
}
