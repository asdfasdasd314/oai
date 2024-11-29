package main

import (
    "fmt"
    "strconv"
    "strings"
    "time"
)

// Gets an int from the user in the specified bounds (inclusive lower bound and exclusive upper bound)
func getAmountOfTimeFromUser(name string, lowerBound int, upperBound int) int {
    var value int
    name = strings.ToLower(name)
    for {
        fmt.Printf("Enter the number of %s (%d-%d): ", name, lowerBound, upperBound - 1); 
    }
    return value
}

func GetTimeFromUser() (int64, *UniqueDailyTime) {
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
		unixTimestamp = AddDayToTime(unixTimestamp)
	}

    return unixTimestamp, NewUniqueDailyTime(actualSeconds, actualMinutes, actualHours)
}

// Takes in the number of seconds since epoch and returns the corresponding time a day later
func AddDayToTime(unixTimestamp int64) int64 {
	return unixTimestamp + 24*60*60
}
