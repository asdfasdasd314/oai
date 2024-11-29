package cli

// We need to parse cli args from the user and this has to exist in a single function because it might need to be its own goroutine
// I want this to be a function that returns the user input so time to structure it

// We have an enum of commands
type Command int

const (
	Help Command = iota,
	Exit,
	Sync,
	SkipSync,
	NextSyncTime,
	TimeUntilSync,
	ListSyncTimes,
	SetSyncTime,
	EraseTime,
	ClearCompletedTasks,
	InvalidCommand,
)

// These are a field on a struct that represents the actual input
type UserInput struct {
	Command Command
	Data struct{} // this could change, but I want this to be dynamic like this but also static
}

func NewUserInput(command Command, data struct{}) (*UserInput) {
	return &UserInput{Command: command, Data: data}
}

func GetUserInput() (*UserInput) {
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
		commandEntered = SetSyncTime
	case "erase-time":
		commandEntered = EraseTime
	case "clear-completed-tasks":
		commandEntered = ClearCompletedTasks
	default:
		commandEntered = InvalidCommand
	}

	return NewUserInput(commandEntered, data)
}