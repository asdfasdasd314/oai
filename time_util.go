package main

import (
    "fmt"
    "strconv"
    "strings"
    "time"
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

func NewUniqueDailyTime(hours int, minutes int, seconds int) *UniqueDailyTime {
    return &UniqueDailyTime{Hours: hours, Minutes: minutes, Seconds: seconds}
}

func NewRecurrentTime(dailyTime UniqueDailyTime, recurrentInterval int) *RecurrentTime {
    return &RecurrentTime{DailyTime: dailyTime, RecurrentInterval: recurrentInterval}
}

// Gets an int from the user in the specified bounds (inclusive lower bound and exclusive upper bound)
// Will always return an int, not int32 or int64, int. There is a check done to ensure it is this specific type (of course the size could change depending on each machinr)
func getAmountOfTimeFromUser(name string, lowerBound int, upperBound int) int {
    var value int
    name = strings.ToLower(name)
    for {
        fmt.Printf("Enter the number of %s (%d-%d): ", name, lowerBound, upperBound - 1); 
		var inputStr string
		fmt.Scanln(&inputStr)
		possiblyValue, err := strconv.ParseInt(inputStr, 10, 0)
		if int64(int(possiblyValue)) != possiblyValue {
			fmt.Println("Number could not fit into int size (32 bits on 32 bit machines and vice versa)")
		} else if err != nil {
			fmt.Println("Enter an actual integer")
		} else if int(possiblyValue) < lowerBound || int(possiblyValue) > upperBound { 
			fmt.Println("Enter a number between 0 and 23")
		} else {
			value= int(possiblyValue)
			break
        }	
    }
    return value
}

func GetTimeFromUser() (int64, *UniqueDailyTime) {
	// Read the hours, minutes, and seconds from the user
	hours := getAmountOfTimeFromUser("hours", 0, 24)
	minutes := getAmountOfTimeFromUser("minutes", 0, 60)
	seconds := getAmountOfTimeFromUser("seconds", 0, 60)

	// Using these hours, minutes, and seconds, we need to calculate what would be the timestamp
	now := time.Now()
	day := now.Day()
	month := now.Month()
	year := now.Year()
	currLocation := now.Location()
	recurrentDateObj := time.Date(year, month, day, hours, minutes, seconds, 0, currLocation)
	unixTimestamp := recurrentDateObj.Unix()
	if now.Unix() > unixTimestamp {
		unixTimestamp = AddDayToTime(unixTimestamp)
	}

    return unixTimestamp, NewUniqueDailyTime(seconds, minutes, hours)
}

// Takes in the number of seconds since epoch and returns the corresponding time a day later
func AddDayToTime(unixTimestamp int64) int64 {
	return unixTimestamp + 24*60*60
}
