package main

import (
	"fmt"
)

// We need to parse cli args from the user and this has to exist in a single function because it might need to be its own goroutine
// I want this to be a function that returns the user input so time to structure it

// We have an enum of commands
type Command int

const (
	Help Command = iota
	Exit
	Sync
	SkipSync
	NextSyncTime
	TimeUntilSync
	ListSyncTimes
	SetSyncTime
	EraseTime
	ClearCompletedTasks
	InvalidCommand
)

type UniqueDailyTime struct {
    Hours int
    Minutes int
    Seconds int
}

type RecurrentTime struct {
    DailyTime UniqueDailyTime
    RecurrentInterval int // The number of days between the sync for the given unique time
}

// These are a field on a struct that represents the actual input
type UserInput struct {
	Action Command
	Data   struct{} // this could change, but I want this to be dynamic like this but also static
}

func NewUniqueDailyTime(hours int, minutes int, seconds int) *UniqueDailyTime {
    return &UniqueDailyTime{Hours: hours, Minutes: minutes, Seconds: seconds}
}

func NewRecurrentTime(dailyTime UniqueDailyTime, recurrentInterval int) *RecurrentTime {
    return &RecurrentTime{DailyTime: dailyTime, RecurrentInterval: recurrentInterval}
}

func NewUserInput(command Command, data struct{}) *UserInput {
	return &UserInput{Action: command, Data: data}
}

// I completely understand I'm going to do the switch twice, but I think this is good practice for understanding enums and how to pass stateful data using enums
// Also it was a slightly poor design choice that I'm not going to bother to tweak
func GetUserInput() *UserInput {
	var input string
	fmt.Scanln(&input)

	var commandEntered Command
	var data struct{}

	switch input {
	case "help":
		commandEntered = Help
	case "exit":
		commandEntered = Exit
	case "sync":
		commandEntered = Sync
	case "skip-sync":
		commandEntered = SkipSync
	case "next-sync-time":
		commandEntered = NextSyncTime
	case "time-until-sync":
		commandEntered = TimeUntilSync
	case "list-sync-times":
		commandEntered = ListSyncTimes
	case "set-sync-time":
		// This one requires we prompt the user for the recurrent time they wish to set
		commandEntered = SetSyncTime

        GetTimeFromUser()
	case "erase-time":
		// This one requires we prompt the user for the unique daily time they wish to set
		commandEntered = EraseTime
	case "clear-completed-tasks":
		commandEntered = ClearCompletedTasks
	default:
		commandEntered = InvalidCommand
	}

	return NewUserInput(commandEntered, data)
}


