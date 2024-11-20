package main

import (
	"bufio"
    "fmt"
    "strings"
    "os"
)

// We can do a DFS across the files/folders, looking through each item at each level and if it's a file
// look through the contents for checkboxes that have been filled, otherwise rerun this code in the folder at a lower level
func cleanCompletedTasks(folderPath string) int {
	items, err := os.ReadDir(folderPath)
	if err != nil {
		panic(err)
	}

    totalTasksCleared := 0

	for _, item := range items {
		itemPath := folderPath + "/" + item.Name()

		if item.IsDir() {
            totalTasksCleared += cleanCompletedTasks(itemPath)
		} else if item.Name()[len(item.Name())-3:len(item.Name())] == ".md" {
	        tasksCleared := cleanFile(itemPath)
            totalTasksCleared += tasksCleared
		}
	}

    return totalTasksCleared
}

func cleanFile(filePath string) int {
	file, err := os.Open(filePath)
    if err != nil {
		panic(err)
	}
    
	defer file.Close()
	
    scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	
    var lines []string
    removedLines := 0

	for scanner.Scan() {
        if len(scanner.Text()) < 5 || !strings.Contains(scanner.Text(), "- [x]") {
            lines = append(lines, scanner.Text())
        } else {
            removedLines++;
        }
	}
    
    if err := scanner.Err(); err != nil {
        panic(err)
    }

    outputFile, err := os.Create(filePath)
    
    if err != nil {
        panic(err)
    }
    defer outputFile.Close()

    // Now write stuff to the file
    writer := bufio.NewWriter(outputFile)
	for _, line := range lines  {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			panic(err)
		}
    }
    
    if removedLines > 0 {
        fmt.Printf("Cleared %d tasks in %s\n", removedLines, filePath)
    }

    writer.Flush()

    return removedLines
}
