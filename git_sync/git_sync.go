package git_sync

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/gen2brain/beeep"
)

type SyncInfo struct {
	DaysBetweenSync   int
	DaysSinceLastSync int
    SkipOccurence bool
}

func NewSyncInfo(daysBetweenSync int, daysSinceLastSync int) *SyncInfo {
	return &SyncInfo{DaysBetweenSync: daysBetweenSync, DaysSinceLastSync: daysSinceLastSync, SkipOccurence: false}
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

	// This treemap stores a map of each unix timestamp to the number of days between syncs and the number of syncs that have occured
	// These entries are int64s to SyncInfo objects
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
				notEmpty := queueItr.First() // Moves to the first element

				if notEmpty {
					firstElement := queueItr.Key()
					firstTimestamp := firstElement.(int64) // This does panic if the type isn't what it is expected to be, but this is just a big script, so I think panicking here is completely fine

					// We've met the condition
					if time.Now().Unix() >= firstTimestamp {
						queueItr := queue.Iterator()
	
                        queueItr.First() // Move the iterator to the first element
						value := queueItr.Value()
						syncInfo := value.(*SyncInfo)

						(*syncInfo).DaysSinceLastSync = ((*syncInfo).DaysSinceLastSync + 1) % (*syncInfo).DaysBetweenSync
                        
						// we know we've hit it
						if (*syncInfo).DaysSinceLastSync == 0 {
                            if (*syncInfo).SkipOccurence == true {
                                (*syncInfo).SkipOccurence = false
                            } else {
							    break
                            }
						} else {
                            // Otherwise we need to bump up the timestamp
                            queue.Remove(firstTimestamp)
                            newTimestamp := addDayToTime(firstTimestamp)
                            queue.Put(newTimestamp, syncInfo)
						}
					}
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
			runGitSyncCommands(retryGitSyncInterval)
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
				syncInfo := value.(*SyncInfo)

				if time.Now().Unix() >= firstTimestamp {
					queue.Remove(firstTimestamp)
					newTimestamp := addDayToTime(firstTimestamp)
					queue.Put(newTimestamp, syncInfo)
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
		case "help":
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

		case "exit":
			// Wait to receive from the channel
			// This waits until there's something in the channel, which always before or after the automatic git sync commands are ran
			<-canExit
			return

		// Sync //
		case "sync":
			fmt.Println("Syncing data...")
			runGitSyncCommands(retryGitSyncInterval)
			fmt.Println("Successfully synced data! | " + formatTime(time.Now()))

		case "skip-sync":
            if queue.Size() == 0 {
                fmt.Println("No times currently queued up to sync");
            } else {
                // We skip the next time
                queueItr := queue.Iterator()
                ok := queueItr.First() // Move to the first element
                if !ok {
                    fmt.Println("Failed to get the first time in the queue")
                } else {
                    value := queueItr.Value()
                    syncInfo := value.(*SyncInfo)
                    
                    // Go until we find a time we aren't skipping
                    foundTimeToSkip := true
                    for (*syncInfo).SkipOccurence == true {
                        ok = queueItr.Next()
                        // If this branch is entered then we must have hit the last time
                        if !ok {
                            fmt.Println("Every time is being skipped")
                            foundTimeToSkip = false
                        } else {
                            value = queueItr.Value()
                            syncInfo = value.(*SyncInfo)
                        }
                    }

                    if foundTimeToSkip {
                        (*syncInfo).SkipOccurence = true
                        key := queueItr.Key()
                        timestamp := key.(int64)

                        formattedSyncTime := formatTime(time.Unix(timestamp, 0))

                        fmt.Println("Will skip the sync at " + formattedSyncTime)
                    }
                }
            }

		// Immutable operatons //
		case "next-sync-time":
			queueItr := queue.Iterator()
			ok := queueItr.First() // Move the iterator to the first element

			if ok {
				// So what we have to do is an O(n) searc for the minimum
				// This is the price for log(n) insertion and arbitrary retrieval and I 100% think that was the correct decision, because automatic syncing doesn't really grow in complexity
				var closestSyncTime int64 = 9_223_372_036_854_775_807 // Maximum value for a 64 bit signed integer

				temp := calculateSyncTimestamp(queueItr.Key().(int64), queueItr.Value().(*SyncInfo))
				if temp < closestSyncTime {
					closestSyncTime = temp
				}

				for queueItr.Next() {
					temp = calculateSyncTimestamp(queueItr.Key().(int64), queueItr.Value().(*SyncInfo))
					if temp < closestSyncTime {
						closestSyncTime = temp
					}
				}

				fmt.Println(formatTime(time.Unix(closestSyncTime, 0)))
			} else {
				fmt.Println("No sync times added yet")
			}

		case "time-until-sync":
			queueItr := queue.Iterator()
			ok := queueItr.First() // Move the iterator to the first element

			if ok {
				// So what we have to do is an O(n) searc for the minimum
				// This is the price for log(n) insertion and arbitrary retrieval and I 100% think that was the correct decision, because automatic syncing doesn't really grow in complexity
				var closestSyncTime int64 = 9_223_372_036_854_775_807 // Maximum value for a 64 bit signed integer

				temp := calculateSyncTimestamp(queueItr.Key().(int64), queueItr.Value().(*SyncInfo))
				if temp < closestSyncTime {
					closestSyncTime = temp
				}

				for queueItr.Next() {
					temp = calculateSyncTimestamp(queueItr.Key().(int64), queueItr.Value().(*SyncInfo))
					if temp < closestSyncTime {
						closestSyncTime = temp
					}
				}

				fmt.Println(getTimeUntilSync(closestSyncTime))
			} else {
				fmt.Println("No sync times added yet")
			}

		case "list-recurrent-times":
			timestamps := queue.Keys()
			syncInfos := queue.Values()

			if len(timestamps) == 0 {
				fmt.Println("No sync times added yet")
			} else {
				for index := range timestamps {
					timestamp := timestamps[index].(int64)
					syncInfo := syncInfos[index].(*SyncInfo)

					formattedTime := formatTime(time.Unix(calculateSyncTimestamp(timestamp, syncInfo), 0))
					fmt.Println(formattedTime + " | Days Between Syncs: " + strconv.FormatInt(int64((*syncInfo).DaysBetweenSync), 10) + " | Days Since Last Sync: " + strconv.FormatInt(int64((*syncInfo).DaysSinceLastSync), 10))
				}
			}

		// Mutable operations //
		case "set-sync-time":
			unixTimestamp := getUnixTimeFromUser()

			// Get the recurrence interval
			var dayInterval int
			for {
				fmt.Print("Enter the days between syncs (recurrence interval): ")
				_, err := fmt.Scanln(&dayInterval)
				if err != nil {
					fmt.Println("Enter an actual integer")
				} else if dayInterval <= 0 {
					fmt.Println("Enter a number greater than 0")
				} else {
					break
				}
			}

			syncInfo := NewSyncInfo(dayInterval, 0)
			queue.Put(unixTimestamp, syncInfo)

			// At the start of the program there is nothing in the channel, so here we have to determine if we can tell the other goroutine that it can now do it's Syncing
			if len(canAccessQueue) == 0 {
				canAccessQueue <- true
			}

		case "erase-time":
			// The automatic syncing is in another goroutine, and so we need to check that it's safe
			// Recieve on channel so we wait until we know it's safe to erase the time
			<-canAccessQueue

			// Get the time the user wants to remove
			timestamp := getUnixTimeFromUser()

			_, found := queue.Get(timestamp)

			if !found {
				fmt.Println("That time is not in use, cannot remomve it")
			} else {
				// Remove that thang
				queue.Remove(timestamp)
				fmt.Println("Successfully removed time")
			}

			// At the end we can send back on the channel so the separate go routine can do it's thing
			canAccessQueue <- true
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

func calculateSyncTimestamp(timestamp int64, syncInfo *SyncInfo) int64 {
	differenceInDays := (*syncInfo).DaysBetweenSync - (*syncInfo).DaysSinceLastSync - 1 // Because otherwise if the two were 10 seconds apart, this would still say they were a day apart. They are a day and some change apart
	return int64(differenceInDays*24*60*60) + timestamp
}

func getUnixTimeFromUser() int64 {
	// Read the hours, minutes, and seconds from the user
	var actualHours int
	for {
		fmt.Print("Enter the number of hours (0-23): ")
		var inputStr string
		fmt.Scanln(&inputStr)
		hours, err := strconv.ParseInt(inputStr, 10, 0)
		if err != nil {
			fmt.Println("Enter an actual integer")
		} else if hours < 0 || hours > 23 {
			fmt.Println("Enter a number between 0 and 23")
		} else {
			actualHours = int(hours)
			break
		}
	}
	var actualMinutes int
	for {
		fmt.Print("Enter the number of minutes (0-59): ")
		var inputStr string
		fmt.Scanln(&inputStr)
		minutes, err := strconv.ParseInt(inputStr, 10, 0)
		if err != nil {
			fmt.Println("Enter an actual integer")
		} else if minutes < 0 || minutes > 59 {
			fmt.Println("Enter a number between 0 and 59")
		} else {
			actualMinutes = int(minutes)
			break
		}
	}
	var actualSeconds int
	for {
		fmt.Print("Enter the number of seconds (0-59): ")
		var inputStr string
		fmt.Scanln(&inputStr)
		seconds, err := strconv.ParseInt(inputStr, 10, 0)
		if err != nil {
			fmt.Println("Enter an actual integer")
		} else if seconds < 0 || seconds > 59 {
			fmt.Println("Enter a number between 0 and 59")
		} else {
			actualSeconds = int(seconds)
			break
		}
	}

	// Using these hours, minutes, and seconds, we need to calculate what would be the timestamp
	now := time.Now()
	day := now.Day()
	month := now.Month()
	year := now.Year()
	currLocation := now.Location()
	recurrentDateObj := time.Date(year, month, day, actualHours, actualMinutes, actualSeconds, 0, currLocation)
	unixTimestamp := recurrentDateObj.Unix()
	if now.Unix() > unixTimestamp {
		unixTimestamp = addDayToTime(unixTimestamp)
	}

	return unixTimestamp
}
