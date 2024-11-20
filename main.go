package main

import (
    "fmt"
)

func main() {
    tasksCleared := cleanCompletedTasks(".")
    if tasksCleared > 0 {
        fmt.Printf("%d tasks were cleared\n", tasksCleared)
    } else {
        fmt.Println("No tasks needed to be cleared")
    }
}
