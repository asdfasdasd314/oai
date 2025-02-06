package main

import (
	"bufio"
	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
	"os"
	"strconv"
	"strings"
)

// I'm in the process of converting my vault into a system based on tags and not folders
// Admittedly I am just listening to what other people say, but my use case is externalizing my thinking, and that is the exact kind of reason these people suggest I use tags and not folders

// There is an edge case where there will be multiple files with the same name, so to take care of that we're going to have a count of how many files of the same name there are, and it's up to the end user to rename these to their heart's content

type ItemEntry struct {
	Item os.DirEntry
	Path []string
}

func newItemEntry(item os.DirEntry, path []string) *ItemEntry {
	return &ItemEntry{item, path}
}

var names map[string]struct{} = make(map[string]struct{})
var symbols map[rune]string = map[rune]string{
	// We also need the space character because it can't be in tag or file names
	' ': "-",

	'!': "", '@': "_at_", '#': "_number_", '$': "_dollars_", '%': "_percent_", '^': "_exp_", '&': "_and_", '*': "_asterisk_",
	'(': "-", ')': "-", '=': "_equals_", '+': "_plus_", '[': "-", ']': "-",
	'{': "-", '}': "-", '|': "_pipe_", '\\': "_backslash_", ';': "-semicolon-", ':': "_colon_", '\'': "", '"': "",
	',': "-", '.': "_", '<': "_less-than_", '>': "_greater-than_", '?': "_question_",
}

// We do a breadth first search because we care about higher level names before lower level names
func FoldersToTags() {
	itemsToVisit := llq.New()

	items, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}

	for _, item := range items {
		itemsToVisit.Enqueue(*newItemEntry(item, []string{item.Name()}))
	}

	for !itemsToVisit.Empty() {
		front, ok := itemsToVisit.Peek()
		if !ok {
			panic("This should never happen")
		}

		entry := front.(ItemEntry)
		itemsToVisit.Dequeue()

		if entry.Item.IsDir() {
			items, err = os.ReadDir(strings.Join(entry.Path, "/"))
			for _, subItem := range items {
				copiedPath := append([]string{}, entry.Path...)
				copiedPath = append(copiedPath, subItem.Name())
				itemsToVisit.Enqueue(*newItemEntry(subItem, copiedPath))
			}
		} else {
			newName, isMarkdown := createFileName(entry.Item.Name())
			if !isMarkdown {
				continue
			}

			moveFile(entry.Path, newName+".md")
		}
	}
}

func moveFile(filePath []string, newName string) {
	file, err := os.Open(strings.Join(filePath, "/"))
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	// Add the path to a tag at the start
	var splitPath []string
	for _, section := range filePath {
		splitPath = append(splitPath, filterName(section))
	}

	var lines []string = []string{}
	if len(splitPath) != 0 {
		tag := strings.Join(splitPath, "/")
		lines = append(lines, "#"+tag[:len(tag)-3])
	}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// Delete original file
	err = os.Remove(strings.Join(filePath, "/"))
	if err != nil {
		panic(err)
	}

	// Save to file at top level
	outputFile, err := os.Create(newName)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			panic(err)
		}
	}

	writer.Flush()
}

// Returns the new name of the file and whether or not it should be modified
func createFileName(initialName string) (string, bool) {
	if initialName[len(initialName)-3:] != ".md" {
		return "", false
	}

	var nameAppend int64 = 0
	for {
		key := initialName + strconv.FormatInt(nameAppend, 10)
		_, ok := names[key]
		if !ok {
			names[key] = struct{}{}
			// We should only be working with markdown files now, so this works
			if nameAppend != 0 {
				return initialName[:len(initialName)-3] + strconv.FormatInt(nameAppend, 10), true
			} else {
				return initialName[:len(initialName)-3], true
			}
		} else {
			nameAppend++
		}
	}
}

// There are often special symbols that have to be converted to something else
func filterName(itemName string) string {
	filtered := ""
	for _, char := range itemName {
		symbol, isSymbol := symbols[char]
		if isSymbol {
			filtered += symbol
		} else {
			filtered += string(char)
		}
	}

	return strings.ToLower(filtered)
}
