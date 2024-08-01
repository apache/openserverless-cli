package tools

import (
	"fmt"
	"os"
)

func Empty() (int, error) {
	if len(os.Args) < 2 {
		fmt.Println("Empty creates an empty file - returns error if it already exists\nUsage: filename")
		return 0, nil
	}
	filename := os.Args[1]

	// Check if the file exists
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return 1, fmt.Errorf("file already exists")
	}

	// Create an empty file
	file, err := os.Create(filename)
	if err != nil {
		return 1, err
	}
	defer file.Close()
	return 0, nil

}
