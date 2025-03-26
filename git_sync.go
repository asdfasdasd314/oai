package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gen2brain/beeep"
)

type UniqueDailyTime struct {
    Hours int
    Minutes int
    Seconds int
}

// Represents all the necessary sync info for a time to sync on
type SyncTime struct {
    CurrTimestamp int64
	DaysBetweenSync   int
}

func NewSyncTime(udt UniqueDailyTime, daysBetweenSync int) *SyncTime {
    return &SyncTime{CurrTimestamp: udt.GetNextOccurence(), DaysBetweenSync: daysBetweenSync}
}

// This function is necessary for placing the timestamp in the current day's range of timestamps
// It places it somewhere in the next 24 hours
func (udt *UniqueDailyTime) GetNextOccurence() int64 {
	now := time.Now()
	day := now.Day()
	month := now.Month()
	year := now.Year()
	location := now.Location()

	// This is the time where we build based on the days between sync and whether not it's going to be skipped
	unixTimestamp := time.Date(year, month, day, udt.Hours, udt.Minutes, udt.Seconds, 0, location).Unix()

	if unixTimestamp < time.Now().Unix() {
		unixTimestamp = unixTimestamp + 24 * 60 * 60;
	}

	return unixTimestamp
}

// This will return the next timestamp to be synced 
func (si *SyncTime) GetNextSync(previousTimestamp int64) int64 {
    return previousTimestamp + int64(si.DaysBetweenSync) * 24 * 60 * 60
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

func calculateTimeUntilSync(syncTimestamp int64) string {
	differenceInTime := syncTimestamp - time.Now().Unix()

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
