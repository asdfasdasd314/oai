package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
)

var int64Comparator = func(a, b interface{}) int {
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

// This holds all of the state necessary to run the app, including the time to retry or verify certain things in a separate goroutine and a queue of the times set up to sync
type AppState struct {
	// I'm not 100% sure these have to be pointers, but I think it is in the actual type, so for now it's like this
	RetryGithubConnectionInterval time.Duration
	VerifyAccurateTimingInterval  time.Duration
	SyncTimes                     *treemap.Map

	// We also need the channels to validate if something can happen
	// I would love to also type the buffer size, but I don't think Go even keeps track of that stuff anyway, so TBD!!!
	CanAccessQueue chan bool
	CanExit        chan bool
}

func InitAppState(retryGithubConnectionInterval time.Duration, verifyAccurateTimingInterval time.Duration) *AppState {
	queue := treemap.NewWith(int64Comparator)
	canAccessQueue := make(chan bool, 1) // We don't immediately send on this because there is nothing in the queue at the start
	canExit := make(chan bool, 1)        // By default we can exit so send on it immediately
	canExit <- true
	return &AppState{RetryGithubConnectionInterval: retryGithubConnectionInterval, VerifyAccurateTimingInterval: verifyAccurateTimingInterval, SyncTimes: queue, CanAccessQueue: canAccessQueue, CanExit: canExit}
}

func printHelpCommands() {
	fmt.Println("Meta Commands")
	fmt.Println("   help: Lists commands that do stuff")
	fmt.Println("   exit: Exits the program safely without potentially being in the middle of a syncing command")
	fmt.Println()

	fmt.Println("Syncing")
	fmt.Println("   sync: Syncs with GitHub")
	fmt.Println("   skipsync: Skips the next sync that would happen")
	fmt.Println()

	fmt.Println("Querying Data")
	fmt.Println("   nextsynctime: Gets the next time GitHub will automatically sync")
	fmt.Println("   timeuntilsync: Calculates the time until the next sync in days, hours, minutes, and seconds")
	fmt.Println("   listsynctimes: Lists all the recurrent times GitHub is synced")
	fmt.Println()

	fmt.Println("Mutating Internal Data")
	fmt.Println("   setsynctime: Sets a recurrent time at which the client will sync with GitHub given some recurring basis of days")
	fmt.Println("   erasetime: Removes a time for which GitHub was supposed to sync")
	fmt.Println()

	fmt.Println("Note Mutating Functions (**BE CAREFUL AS THESE INTERACT WITH YOU'RE ACTUAL NOTES**)")
	fmt.Println("   cleancompletedtasks: Clears completed tasks throughout the entire vault using a recursive function")
	fmt.Println()
}

func skipNextSync(syncTimes *treemap.Map) {
    itr := syncTimes.Iterator()
    ok := itr.First()
    if !ok {
        fmt.Println("No times added, so can't skip any")
        return
    }

    syncTime, _ := getClosestSyncTime(syncTimes)
    if (*syncTime).SkipOccurence {
        fmt.Println("Already skipping that time")
        return
    }

    fmt.Println("Skipping sync time at " + time.Unix(syncTime.GetSyncTimestamp(), 0).Format(time.UnixDate))
    (*syncTime).SkipOccurence = true
    syncTimes.Put(syncTime.GetCurrentDayTimestamp(), syncTime)
    fmt.Println("New sync at " + time.Unix(syncTime.GetSyncTimestamp(), 0).Format(time.UnixDate))
}

// A helper function for the `getNextSyncTime` and `getTimeUntilNextsync` functions
// Does an O(n) search through the `syncTimes` and returns the smallest timestamp at which a sync should occur
// Returns a timestamp and a boolean to represent if a successful retrieval occured (theoretically this would only not work if if there was nothing in the tree at the start, so basically that's the meaning of the boolean)
// The sync timestamp is measured in seocnds
func getClosestSyncTime(syncTimes *treemap.Map) (*SyncTime, bool) {
	itr := syncTimes.Iterator()
	ok := itr.First() // Move the iterator to the first element

	if !ok {
		return nil, false
	}

	// So what we have to do is an O(n) searc for the minimum
	// This is the price for log(n) insertion and arbitrary retrieval and I 100% think that was the correct decision, because automatic syncing doesn't really grow in complexity
    
	var closestSyncTimestamp int64 = 9_223_372_036_854_775_807 // Maximum value for a 64 bit signed integer
    var closestSync *SyncTime; 

	syncTime := itr.Value().(*SyncTime)
	temp := syncTime.GetSyncTimestamp()

	if temp < closestSyncTimestamp {
		closestSyncTimestamp = temp
        closestSync = syncTime
	}

	for itr.Next() {
		syncTime = itr.Value().(*SyncTime)
		temp = syncTime.GetSyncTimestamp()
		if temp < closestSyncTimestamp {
			closestSyncTimestamp = temp
            closestSync = syncTime
		}
	}

	return closestSync, true
}

// Returns the next time a sync will occur formatted as a string
func getNextSyncTime(syncTimes *treemap.Map) string {
	closestSyncTime, ok := getClosestSyncTime(syncTimes)
	if !ok {
		return "No sync times added yet"
	}

	return time.Unix(closestSyncTime.GetSyncTimestamp(), 0).Format(time.UnixDate)
}

// Returns the time until the next sync formatted as a string, or a string to signify that there were no times in the SyncTimes tree yet
func getTimeUntilNextSync(syncTimes *treemap.Map) string {
	closestSyncTime, ok := getClosestSyncTime(syncTimes)
	if !ok {
		return "No sync times added yet"
	}

	return calculateTimeUntilSync(closestSyncTime.GetSyncTimestamp())
}

func printSyncTimes(syncTimes *treemap.Map) {
	timestamps := syncTimes.Keys()
	values := syncTimes.Values()

	if len(timestamps) == 0 {
		fmt.Println("No sync times added yet")
		return
	}

	for index := range timestamps {
		syncTime := values[index].(*SyncTime)

		formattedTime := formatTime(time.Unix((*syncTime).GetSyncTimestamp(), 0))
		fmt.Print(formattedTime)
		fmt.Print(" | Days Between Syncs: " + strconv.FormatInt(int64((*syncTime).DaysBetweenSync), 10))
		fmt.Print(" | Days Since Last Sync: " + strconv.FormatInt(int64((*syncTime).DaysSinceLastSync), 10))

		fmt.Print(" | ")
		if (*syncTime).SkipOccurence {
			fmt.Println("Skipping next occurence")
		} else {
			fmt.Println("Not skipping next occurence")
		}
	}
}

func getSyncTimeFromUser() *SyncTime {
	_, dailyTime := GetTimeFromUser()

	// Get the days between syncs
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

	return NewSyncTime(*dailyTime, dayInterval)
}

// This adds to the queue of times
// Also this asks for the app state because it has to do a lot more business logic than just receiving something from the queue
func setSyncTime(newTime *SyncTime, appState *AppState) {
	(*appState).SyncTimes.Put(newTime.GetCurrentDayTimestamp(), newTime) // Add pointers to these sync times
	// At the start of the program there is nothing in the channel, so here we have to determine if we can tell the other goroutine that it can now do it's Syncing
	if len((*appState).CanAccessQueue) == 0 {
		(*appState).CanAccessQueue <- true
	}
}

// This asks for a `UniqueDailyTime` because (as of now) there can be only one sync time per unique daily time, and this is way more convenient for the user
// Also this asks for the app state because it has to do a lot more business logic than just receiving something from the queue
func eraseSyncTime(timestampToRemove int64, appState *AppState) {
	// First check if there are even any times
	if (*appState).SyncTimes.Size() == 0 {
		fmt.Println("No times have been set to sync, this should have been caught earlier")
	}

	// The automatic syncing is in another goroutine, and so we need to check that it's safe
	// Recieve on channel so we wait until we know it's safe to erase the time
	<-(*appState).CanAccessQueue

	_, found := (*appState).SyncTimes.Get(timestampToRemove)

	if !found {
		fmt.Println("That time is not in use, cannot remove it")
	} else {
		// Remove that thang
		(*appState).SyncTimes.Remove(timestampToRemove)
		fmt.Println("Successfully removed time")
	}

	// Send back on the channel so it can be accessed
	(*appState).CanAccessQueue <- true
}

// Runs all app logic to not clutter main file
func RunApp(retryGithubConnectionInterval time.Duration, verifyAccurateTimingInterval time.Duration, inDebugMode bool) {
	// First initialize appstate
	appState := InitAppState(retryGithubConnectionInterval, verifyAccurateTimingInterval)

	// Load environment variables //

	// Set automatic syncing to happen in a separate goroutine, but can still communicate with this main goroutine via some channels I have setup
	go AutomaticSync(appState, inDebugMode)

	// Run the blocking code that the user interacts with
    printHelpCommands()
	for {
		var input string
		fmt.Scanln(&input)

		// Basically all of these cases should call some separate function
		switch input {
		case "help":
			printHelpCommands()
		case "exit":
			<-(*appState).CanExit // First check we can exit
			return
		case "sync":
			fmt.Println("Syncing with Github...")
			if !inDebugMode {
				runGitSyncCommands(retryGithubConnectionInterval)
			} else {
				fmt.Println("In debug so not actually syncing with Github")
			}
			fmt.Println("Synced with Github!")
		case "skipsync":
			skipNextSync((*appState).SyncTimes)
		case "nextsynctime":
			fmt.Println(getNextSyncTime((*appState).SyncTimes))
		case "timeuntilsync":
			fmt.Println(getTimeUntilNextSync((*appState).SyncTimes))
		case "listsynctimes":
			printSyncTimes((*appState).SyncTimes)
		case "setsynctime":
			syncTime := getSyncTimeFromUser()
			setSyncTime(syncTime, appState)
		case "erasetime":
			// First check if there are even any times
			if (*appState).SyncTimes.Size() == 0 {
				fmt.Println("No times have been set to sync")
				continue
			}

			timestamp, _ := GetTimeFromUser()
			eraseSyncTime(timestamp, appState)
		case "clearcompletedtasks":
			if !inDebugMode {
				ClearCompletedTasks(".")
			} else {
				fmt.Println("Can't run this command in debug mode :(")
			}
		}
	}
}
