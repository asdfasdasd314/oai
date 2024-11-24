package git_sync

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

type CreateRecurrentTimeError int

const (
	InputFormat CreateRecurrentTimeError = iota // this starts at 0
	ParseHours
	ParseMinutes
	ParseSeconds
	InvalidHours
	InvalidMinutes
	InvalidSeconds
)

func (e CreateRecurrentTimeError) Error() string {
	switch e {
	case InputFormat:
		return "Invalid recurrent time format"
	case ParseHours:
		return "Could not parse hours string"
	case ParseMinutes:
		return "Could not parse minutes string"
	case ParseSeconds:
		return "Could not parse seconds string"
	case InvalidHours:
		return "Hours was not between 0 and 23"
	case InvalidMinutes:
		return "Minutes was not between 0 and 59"
	case InvalidSeconds:
		return "Seconds was not between 0 and 59"
	default:
		return "This error wasn't handled properly in CreateRecurrentTimeError"
	}
}

type RecurrentTime struct {
	Hours       int
	Minutes     int
	Seconds     int
	DayInterval int
}

func NewRecurrentTime(timeAsStr string) (*RecurrentTime, error) {
	splitTime := strings.Split(timeAsStr, ":")
	if len(splitTime) != 3 {
		return nil, InputFormat
	}

	hours, err := strconv.ParseInt(splitTime[0], 10, 32)
	if err != nil {
		return nil, ParseHours
	}
	minutes, err := strconv.ParseInt(splitTime[1], 10, 32)
	if err != nil {
		return nil, ParseMinutes
	}
	seconds, err := strconv.ParseInt(splitTime[2], 10, 32)
	if err != nil {
		return nil, ParseSeconds
	}

	if hours > 23 || hours < 0 {
		return nil, InvalidHours
	}
	if minutes > 59 || minutes < 0 {
		return nil, InvalidMinutes
	}
	if seconds > 59 || seconds < 0 {
		return nil, InvalidSeconds
	}

	return &RecurrentTime{Hours: int(hours), Minutes: int(minutes), Seconds: int(seconds)}, nil
}

func AutomaticGitSync(syncInterval time.Duration, checkTimeAccurateInterval time.Duration, retryGitSyncInterval time.Duration) {
	// This one automatically syncs on the specified interval
	canExit := make(chan bool, 1)
	// By default either one can exit
	canExit <- true

	// Initialize the queue (because this is actually just a slice, dequeing is O(n)
	var recurrentTimesQueue []int64 = []int64{time.Now().Add(syncInterval).Unix()}

	go func() {
		// There could theoretically be an issue in that the user may exit while the sync is happening
		// We don't want that to happen, so here we can use channels to pass messages about the completion status of each goroutine
		for {
			// Here we need to check if we have hit the first time in the queue
			for time.Now().Unix() <= recurrentTimesQueue[0] {
				time.Sleep(checkTimeAccurateInterval)
			}

			// Receive from the channel so the main goroutine must stop
			<-canExit

			// If we've gotten to this point we need to guarantee that we can run the git commands
			fmt.Println("Syncing data automatically...")
			// runGitSyncCommands(retryGitSyncInterval)
			notifySuccess()
			fmt.Println("Successfully synced data! | " + formatTime(time.Now()))

			// Send to the channel so now the main goroutine can exit if it wants to
			canExit <- true

			// Remember before where we checked if we hit the FIRST time, well there's a chance we've actually hit multiple
			// This would only happen if the computer were to sleep through a few of these, but anyways we have to clear all of the
			// times up until the current time

			for time.Now().Unix() >= recurrentTimesQueue[0] {
				// Pushes back the things one day until the first point in the queue is going to happen after the moment at which this is evaluated
				newTimestamp := addDayToTime(recurrentTimesQueue[0])
				recurrentTimesQueue = recurrentTimesQueue[1:]
				insertIntoQueue(newTimestamp, &recurrentTimesQueue)
			}
		}
	}()

	// This syncs on user input
	for {
		var input string
		fmt.Scanln(&input)

		switch input {
		// Immutable operatons //
		case "sync":
			fmt.Println("Syncing data...")
			// runGitSyncCommands(retryGitSyncInterval)
			fmt.Println("Successfully synced data! | " + formatTime(time.Now()))
		case "next-sync-time":
			fmt.Println(formatTime(time.Unix(recurrentTimesQueue[0], 0)))
		case "time-until-sync":
			fmt.Println(getTimeUntilSync(recurrentTimesQueue[0]))
		case "list-current-recurrent-times":
			for _, timestamp := range recurrentTimesQueue {
				fmt.Println(formatTime(time.Unix(timestamp, 0)))
			}

		// Mutable operations //
		case "set-sync-time":
			// Read the hours, minutes, and seconds from the user
			fmt.Print("Enter the number of hours, minutes and seconds (HH:MM:SS) in military time: ")
			var recurringTimeString string
			fmt.Scanln(&recurringTimeString)

			var errorCreatingTime bool

			recurrentTime, err := NewRecurrentTime(recurringTimeString)
			if err != nil {
				errorCreatingTime = true
			}

			for errorCreatingTime {
				fmt.Println(err.Error())
				fmt.Print("Enter the number of hours, minutes and seconds (HH:MM:SS) in military time: ")
				fmt.Scanln(&recurringTimeString)

				recurrentTime, err = NewRecurrentTime(recurringTimeString)
				if err != nil {
					errorCreatingTime = true
				} else {
					errorCreatingTime = false
				}
			}

			now := time.Now()

			// Now we have a valid recurrent time object, we have to create the timestamp
			currYear, currMonth, currDay := now.Date()
			currLocation := now.Location()

			dateObject := time.Date(currYear, currMonth, currDay, recurrentTime.Hours, recurrentTime.Minutes, recurrentTime.Seconds, 0, currLocation)
			unixTime := dateObject.Unix()

			// Next we add to the queue

			/* First make sure it's not in the queue, which is an O(n) operation, and we can guarantee that there will be no duplicates because we limited the recurrent times to those between the
			   current time today and the corresponding time tomorrow */

			found := false
			for _, timestamp := range recurrentTimesQueue {
				// If we already have it stored then just don't add it
				if timestamp == unixTime {
					fmt.Println("That timestamp is already being used")
					found = true
					break
				}
			}

			if !found {
				insertIntoQueue(unixTime, &recurrentTimesQueue)
			}

		// Exit //
		case "exit":
			// Wait to receive from the channel
			// This waits until there's something in the channel, which always before or after the automatic git sync commands are ran
			<-canExit
			return
		}
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

// Takes the hours, minutes, and seconds of a military time that happens every day and maps it to the next closest occurence of that time
// For example, if it were 6 PM on a certain day and 15 hours, 30 minutes, 30 seconds was given, then 3:30:30 PM on the next day would be returned
func mapRecurrentTimeToTimestamp(hours int, minutes int, seconds int) int64 {
	now := time.Now()
	currHours := now.Hour()
	currMinutes := now.Minute()
	currSeconds := now.Second()

	currTotalSeconds := currHours*60*60 + currMinutes*60 + currSeconds
	recurrentTimeTotalSeconds := hours*60*60 + minutes*60 + seconds

	currYear, currMonth, currDay := now.Date()
	currLocation := now.Location()

	recurrentTime := time.Date(currYear, currMonth, currDay, hours, minutes, seconds, 0, currLocation)
	unixTime := recurrentTime.Unix()

	if currTotalSeconds > recurrentTimeTotalSeconds { // Put the recurrent timestamp on the next day
		return unixTime
	} else { // Put the recurrent timestamp on the current day
		return addDayToTime(unixTime)
	}
}

// Inserts a new timestamp into the existing timestamps
func findInsertionIndex(newTimestamp int64, existingTimestamps []int64) int {
	if len(existingTimestamps) == 0 || newTimestamp < existingTimestamps[0] {
		return 0
	}

	if newTimestamp > existingTimestamps[len(existingTimestamps)-1] {
		return len(existingTimestamps)
	}

	// Do the binary search
	left := 0
	right := len(existingTimestamps) - 1
	for left < right {
		middle := (right + left) / 2
		if right-left == 1 {
			if newTimestamp > existingTimestamps[left] {
				return right
			} else {
				return left
			}
		}
		if newTimestamp < existingTimestamps[middle] {
			left = middle
		} else {
			right = middle
		}
	}

	return right
}

func formatTime(timeObj time.Time) string {
	return string(timeObj.Format(time.UnixDate))
}

func getTimeUntilSync(nextSyncTimestamp int64) string {
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

// Takes in the number of seconds since epoch and returns the corresponding time a day later
func addDayToTime(unixTimestamp int64) int64 {
	return unixTimestamp + 24*60*60
}

func insertIntoQueue(newTimestamp int64, timestampsQueue *[]int64) {
	// Do a linear search and insert at the correct index (this is really really slow, but like I don't care)
	// I might be the guy that does a linear search on a sorted array
	for index, timestamp := range *timestampsQueue {
		if timestamp > newTimestamp {
			*timestampsQueue = append((*timestampsQueue)[:index], append([]int64{newTimestamp}, (*timestampsQueue)[index:]...)...)
			return
		}
	}

	// If we get here then we have the greatest possible timestamp, so just add to the end
	*timestampsQueue = append((*timestampsQueue), newTimestamp)
}
