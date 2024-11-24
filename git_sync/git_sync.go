package git_sync

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gen2brain/beeep"
    "github.com/emirpasic/gods/maps/treemap"
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
}

func NewRecurrentTime(hours int, minutes int, seconds int) (*RecurrentTime) {
    return &RecurrentTime{Hours: hours, Minutes: minutes, Seconds: seconds}
}

func AutomaticGitSync(checkTimeAccurateInterval time.Duration, retryGitSyncInterval time.Duration) {
	// By default either one can exit
	canExit := make(chan bool, 1)
	canExit <- true

    canAccessQueue := make(chan bool, 1)

    // We use a treeset because it's ordered and has log(n) insertion with unique elements
    // The keys are unix timestamps so they can be sorted and accessed in order

    int64Comparator := func(a, b interface{}) int {
		ka := a.(int64)
		kb := b.(int64)
		switch {
		case ka < kb:
			return -1
		case ka > kb:
			return 1
		default:
			return 0
		}
	}

    queue := treemap.NewWith(int64Comparator)

	go func() {
		// There could theoretically be an issue in that the user may exit while the sync is happening
		// We don't want that to happen, so here we can use channels to pass messages about the completion status of each goroutine
		for {

            // Do nothing until the condition is met to break out of the loop (i.e., there is something in the queue and we have passed the time of that first thing in the queue)
            for {
                // We have to wait for something to be in the queue, but also we need to wait until either this loop will run again, or the end of this save-cycle to allow the main goroutine to go again
                // This means that we have to not send on the channel and make it wait until two points later in this loop
                <-canAccessQueue

                // The above blocks until something is placed in the queue, and then we can access the queue
                
                // Here we need to check if we have hit the first time in the queue
                queueItr := queue.Iterator()
                queueItr.First() // Moves to the first element

                firstElement := queueItr.Key()
                firstTimestamp := firstElement.(int64) // This does panic if the type isn't what it is expected to be, but this is just a big script, so I think panicking here is completely fine

                // We've met the condition
                if time.Now().Unix() >= firstTimestamp {
                    break
                }

                // In this situation we can send back on the channel because we don't care if the user erases this recurrence if we've already determined we're going to wait
                canAccessQueue <- true

                // Otherwise we sleep
                time.Sleep(checkTimeAccurateInterval)
            }

            // Notice that we never sent back on the `emptyQueue` channel, so the main goroutine should be waiting, and we can run all of this code safely
            
			// Receive from the channel so the main goroutine must stop
			<-canExit

			// If we've gotten to this point we need to guarantee that we can run the git commands
			fmt.Println("Syncing data automatically...")
			// runGitSyncCommands(retryGitSyncInterval)
			notifySuccess()
			fmt.Println("Successfully synced data! | " + formatTime(time.Now()))

			// Send to the channel so now the main goroutine can exit if it wants to
			canExit <- true

            // Now we have to adjust the queue
            // We know we can adjust it because the main goroutine should be blocked
            for {
                queueItr := queue.Iterator()
                queueItr.First() // Move the iterator to the first element

                firstKey := queueItr.Key()
                firstTimestamp := firstKey.(int64)

                value := queueItr.Value()
                dayInterval := value.(int)

                if time.Now().Unix() >= firstTimestamp {
                    queue.Remove(firstTimestamp)
                    newTimestamp := addDayToTime(firstTimestamp)
                    queue.Put(newTimestamp, dayInterval)
                } else {
                    break
                }
            }

            // At the end we send back on the `emptyQueue` goroutine
            canAccessQueue <- true
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
            queueItr := queue.Iterator()
            ok := queueItr.First() // Move the iterator to the first element

            if ok {
                firstElement := queueItr.Key()
                firstTimestamp := firstElement.(int64)

                fmt.Println(formatTime(time.Unix(firstTimestamp, 0)))
            } else {
                fmt.Println("No sync times added yet")
            }

		case "time-until-sync":
            queueItr := queue.Iterator()
            ok := queueItr.First() // Move the iterator to the first element

            if ok {
                firstElement := queueItr.Key()
                firstTimestamp := firstElement.(int64)

                fmt.Println(getTimeUntilSync(firstTimestamp))
            } else {
                fmt.Println("No sync times added yet")
            }

		case "list-current-recurrent-times":
            timestamps := queue.Keys()

            if len(timestamps) == 0 {
                fmt.Println("No sync times added yet")
            } else {
                for _, value := range timestamps {
                    timestamp := value.(int64)
                    fmt.Println(formatTime(time.Unix(timestamp, 0)))
                }
            }

		// Mutable operations //
		case "set-sync-time":
			// Read the hours, minutes, and seconds from the user
            var hours int
            for {
                fmt.Print("Enter the number of hours (0-23): ")
                _, err := fmt.Scanln(&hours)
                if err != nil {
                    fmt.Println("Enter an actual integer")
                } else if hours < 0 || hours > 23 {
                    fmt.Println("Enter a number between 0 and 23")
                } else {
                    break
                }
            }
            var minutes int
            for {
                fmt.Print("Enter the number of minutes (0-59): ")
                _, err := fmt.Scanln(&minutes)
                if err != nil {
                    fmt.Println("Enter an actual integer")
                } else if minutes < 0 || minutes > 59 {
                    fmt.Println("Enter a number between 0 and 59")
                } else {
                    break
                }
            }
           var seconds int
            for {
                fmt.Print("Enter the number of seconds (0-59): ")
                _, err := fmt.Scanln(&seconds)
                if err != nil {
                    fmt.Println("Enter an actual integer")
                } else if seconds < 0 || seconds > 59 {
                    fmt.Println("Enter a number between 0 and 59")
                } else {
                    break
                }
            }
            
            // Using these hours, minutes, and seconds, we need to calculate what would be the timestamp
            now := time.Now()
            day := now.Day()
            month := now.Month()
            year := now.Year()
            currLocation := now.Location()
            recurrentDateObj := time.Date(year, month, day, hours, minutes, seconds, 0, currLocation)
            unixTimestamp := recurrentDateObj.Unix()
            if now.Unix() > unixTimestamp {
                unixTimestamp = addDayToTime(unixTimestamp)
            }

            // Get the recurrence interval
            var dailyInterval int
            for {
                fmt.Print("Enter the days between syncs (recurrence interval): ")
                _, err := fmt.Scanln(&dailyInterval)
                if err != nil {
                    fmt.Println("Enter an actual integer")
                } else if dailyInterval < 0 {
                    fmt.Println("Enter a number greater than 0")
                } else {
                    break
                }
            }

            queue.Put(unixTimestamp, dailyInterval)

            // At the start of the program there is nothing in the channel, so here we have to determine if we can tell the other goroutine that it can now do it's Syncing
            if len(canAccessQueue) == 0 {
                canAccessQueue <- true
            }
        
        case "erase-sync-time":
            panic("TODO")

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
