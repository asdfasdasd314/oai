package main

import (
	"time"

	"github.com/emirpasic/gods/maps/treemap"

	"context"
	"fmt"
	"sync"
)

// App struct
type App struct {
	ctx       context.Context
	syncTimes *treemap.Map
	syncMutex *sync.Mutex // Add mutex for thread-safe operations

	RetryConnectionInterval time.Duration
	RestInterval            time.Duration

	InDebugMode bool
}

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

// NewApp creates a new App application struct
func NewApp(retryConnectionInterval time.Duration, restInterval time.Duration, inDebugMode bool) *App {
	mutex := sync.Mutex{}
	queue := treemap.NewWith(int64Comparator)
	return &App{
		syncTimes:               queue,
		syncMutex:               &mutex,
		RetryConnectionInterval: retryConnectionInterval,
		RestInterval:            restInterval,
		InDebugMode:             inDebugMode,
	}
}

func (app *App) ExecuteSyncs() {
	// There could theoretically be an issue in that the user may exit while the sync is happening
	// We don't want that to happen, so here we can use a mutex to lock the queue
	for {

		// Do nothing until the condition is met to break out of the loop (i.e., there is something in the queue and we have passed the time of that first thing in the queue)
		for {
			// Here we need to check if we have hit the first time in the queue
			queueItr := app.syncTimes.Iterator()
			notEmpty := queueItr.First() // Moves to the first element

			if notEmpty {
				app.syncMutex.Lock()
				firstElement := queueItr.Key()
				firstTimestamp := firstElement.(int64) // This does panic if the type isn't what it is expected to be, but this is just a big script, so I think panicking here is completely fine

				if time.Now().Unix() >= firstTimestamp {
					break
				}
			}

			// In this situation we can send back on the channel because we don't care if the user erases this recurrence if we've already determined we're going to wait

			// Otherwise we sleep
			time.Sleep(app.RestInterval)
		}

		// Receive from the channel so the main goroutine must stop
		app.syncMutex.Lock()
		defer app.syncMutex.Unlock()

		// If we've gotten to this point we need to guarantee that we can run the git commands
		if !app.InDebugMode {
			runGitSyncCommands(app.RetryConnectionInterval)
		}
		notifySuccess()
		fmt.Println("Successfully synced data! | " + formatTime(time.Now()))

		// Now we have to adjust the queue as we didn't actually update anything above
		for {
			queueItr := app.syncTimes.Iterator()
			queueItr.First() // Move the iterator to the first element

			firstKey := queueItr.Key()
			firstTimestamp := firstKey.(int64)

			value := queueItr.Value()
			syncTime := value.(*SyncTime)

			if time.Now().Unix() >= firstTimestamp {
				app.syncTimes.Remove(firstTimestamp)
				newTimestamp := syncTime.GetNextSync(firstTimestamp)
				app.syncTimes.Put(newTimestamp, syncTime)
			} else {
				break
			}
		}
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	go a.ExecuteSyncs() // Leave this running in a separate goroutine
}

func SkipNextSync(app *App) {
	app.syncMutex.Lock()
	defer app.syncMutex.Unlock()

	itr := app.syncTimes.Iterator()
	ok := itr.First()
	if !ok {
		fmt.Println("No times added, so can't skip any")
		return
	}

	syncTime, _ := app.GetClosestSyncTime()
	nextTime := syncTime.GetNextSync(syncTime.CurrTimestamp)

	app.syncTimes.Remove(syncTime.CurrTimestamp)
	fmt.Println("Skipping sync time at " + time.Unix(syncTime.CurrTimestamp, 0).Format(time.UnixDate))
	app.syncTimes.Put(nextTime, syncTime)
	fmt.Println("New sync at " + time.Unix(nextTime, 0).Format(time.UnixDate))
}

func (app *App) GetClosestSyncTime() (*SyncTime, bool) {
	app.syncMutex.Lock()
	defer app.syncMutex.Unlock()

	itr := app.syncTimes.Iterator()
	ok := itr.First() // Move the iterator to the first element

	if !ok {
		return nil, false
	}

	syncTime := itr.Value().(*SyncTime)
	return syncTime, true
}

// This adds to the queue of times
func (app *App) SetSyncTime(newTime *SyncTime) {
	app.syncMutex.Lock()
	defer app.syncMutex.Unlock()

	(*app).syncTimes.Put(newTime.CurrTimestamp, newTime) // Add pointers to these sync times
}

// This asks for a `UniqueDailyTime` because (as of now) there can be only one sync time per unique daily time, and this is way more convenient for the user
// Also this asks for the app state because it has to do a lot more business logic than just receiving something from the queue
func (app *App) EraseSyncTime(timestamp int64) {
	app.syncMutex.Lock()
	defer app.syncMutex.Unlock()

	// First check if there are even any times
	if app.syncTimes.Size() == 0 {
		fmt.Println("No times have been set to sync, this should have been caught earlier")
	}

	// The automatic syncing is in another goroutine, and so we need to check that it's safe
	// Recieve on channel so we wait until we know it's safe to erase the time

	_, found := app.syncTimes.Get(timestamp)

	if !found {
		fmt.Println("That time is not in use, cannot remove it")
	} else {
		// Remove that thang
		app.syncTimes.Remove(timestamp)
		fmt.Println("Successfully removed time")
	}
}
