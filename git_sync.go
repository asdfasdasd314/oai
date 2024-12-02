package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gen2brain/beeep"
)

// Represents all the necessary sync info for a time to sync on
type SyncTime struct {
    DailyTime UniqueDailyTime
	DaysBetweenSync   int
	DaysSinceLastSync int
    SkipOccurence bool
}

func NewSyncTime(recurrentSync UniqueDailyTime, daysBetweenSync int) *SyncTime {
    return &SyncTime{DailyTime: recurrentSync, DaysBetweenSync: daysBetweenSync, DaysSinceLastSync: 0, SkipOccurence: false}
}

func (si *SyncTime) GetSyncTimestamp() int64 {
    now := time.Now()
    day := now.Day()
    month := now.Month()
    year := now.Year()
    location := now.Location()
    
    // This is the time where we build based on the days between sync and whether not it's going to be skipped
    unixTimestamp := time.Date(year, month, day, si.DailyTime.Hours, si.DailyTime.Minutes, si.DailyTime.Seconds, 0, location).Unix()

    if (unixTimestamp < time.Now().Unix()) {
        unixTimestamp = AddDayToTime(unixTimestamp)
    }

    // Adjust timestamp for the difference in days
	differenceInDays := si.DaysBetweenSync - si.DaysSinceLastSync - 1 // Because otherwise if the two were 10 seconds apart, this would still say they were a day apart. They are a day and some change apart
    unixTimestamp += int64(differenceInDays*24*60*60)
    
    // Adjust for if the time will be skipped or not
    if si.SkipOccurence {
        unixTimestamp += int64(si.DaysBetweenSync*24*60*60)
    }

    return unixTimestamp
}

// TODO: This code is too messy
func AutomaticSync(appState *AppState, inDebugMode bool) {
    // There could theoretically be an issue in that the user may exit while the sync is happening
    // We don't want that to happen, so here we can use channels to pass messages about the completion status of each goroutine
    for {

        // Do nothing until the condition is met to break out of the loop (i.e., there is something in the queue and we have passed the time of that first thing in the queue)
        for {
            // We have to wait for something to be in the queue, but also we need to wait until either this loop will run again, or the end of this save-cycle to allow the main goroutine to go again
            // This means that we have to not send on the channel and make it wait until two points later in this loop
            <-(*appState).CanAccessQueue

            // The above blocks until something is placed in the queue, and then we can access the queue

            // Here we need to check if we have hit the first time in the queue
            queueItr := (*appState).SyncTimes.Iterator()
            notEmpty := queueItr.First() // Moves to the first element

            if notEmpty {
                firstElement := queueItr.Key()
                firstTimestamp := firstElement.(int64) // This does panic if the type isn't what it is expected to be, but this is just a big script, so I think panicking here is completely fine

                // We've met the condition
                if time.Now().Unix() >= firstTimestamp {
                    value := queueItr.Value()
                    syncTime := value.(*SyncTime)

                    (*syncTime).DaysSinceLastSync++
                    (*syncTime).DaysSinceLastSync %= (*syncTime).DaysBetweenSync
                    
                    // we know we've hit it
                    if (*syncTime).DaysSinceLastSync == 0 {
                        if (*syncTime).SkipOccurence == true {
                            (*syncTime).SkipOccurence = false
                            (*appState).SyncTimes.Remove(firstTimestamp)
                            newTimestamp := AddDayToTime(firstTimestamp)
                            (*appState).SyncTimes.Put(newTimestamp, syncTime)
                        } else {
                            break
                        }
                    } else {
                        // Otherwise we need to bump up the timestamp
                        (*appState).SyncTimes.Remove(firstTimestamp)
                        newTimestamp := AddDayToTime(firstTimestamp)
                        (*appState).SyncTimes.Put(newTimestamp, syncTime)
                    }
                }
            }

            // In this situation we can send back on the channel because we don't care if the user erases this recurrence if we've already determined we're going to wait
            (*appState).CanAccessQueue <- true

            // Otherwise we sleep
            time.Sleep((*appState).VerifyAccurateTimingInterval)
        }

        // Notice that we never sent back on the `emptyQueue` channel, so the main goroutine should be waiting, and we can run all of this code safely

        // Receive from the channel so the main goroutine must stop
        <-(*appState).CanExit

        // If we've gotten to this point we need to guarantee that we can run the git commands
        fmt.Println("Syncing data automatically...")
        if !inDebugMode {
            runGitSyncCommands((*appState).RetryGithubConnectionInterval)
        } else {
            fmt.Println("In debug so not actually syncing with Github")
        }
        notifySuccess()
        fmt.Println("Successfully synced data! | " + formatTime(time.Now()))

        // Send to the channel so now the main goroutine can exit if it wants to
        (*appState).CanExit <- true

        // Now we have to adjust the queue
        // We know we can adjust it because the main goroutine should be blocked
        for {
            queueItr := (*appState).SyncTimes.Iterator()
            queueItr.First() // Move the iterator to the first element

            firstKey := queueItr.Key()
            firstTimestamp := firstKey.(int64)

            value := queueItr.Value()
            syncTime := value.(*SyncTime)
            
            // TODO: this is wrong because if something happens a second later and the for loop in the previous portion doesn't catch it, then it will here but even if it's supposed three days from now, it will move it to the end and not adjust the data appropriately
            if time.Now().Unix() >= firstTimestamp {
                (*appState).SyncTimes.Remove(firstTimestamp)
                newTimestamp := AddDayToTime(firstTimestamp)
                (*appState).SyncTimes.Put(newTimestamp, syncTime)
            } else {
                break
            }
        }

        // At the end we send back on the goroutine
        (*appState).CanAccessQueue <- true
    }
}

func runGitSyncCommands(retryGitSyncInterval time.Duration) {
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

	githubIsAccessible := isGithubAccessible()

	for !githubIsAccessible {
		fmt.Println("Couldn't access GitHub... trying again in 2 minutes")
		time.Sleep(retryGitSyncInterval)
		githubIsAccessible = isGithubAccessible()
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

func formatTime(timeObj time.Time) string {
	return string(timeObj.Format(time.UnixDate))
}

func calculateTimeUntilSync(nextSyncTimestamp int64) string {
	differenceInTime := nextSyncTimestamp - time.Now().Unix()

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

// Validates that github can be reached
func isGithubAccessible() bool {
	_, err := http.Get("https://www.github.com")
	return err == nil
}
