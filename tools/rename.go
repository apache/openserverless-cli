package tools

import (
	"fmt"
	"os"
)

func Rename() (int, error) {
	if len(os.Args) != 3 {
		fmt.Println("Rename a file\nUsage: rename <source> <destination>")
		return 0, nil
	}

	source := os.Args[1]
	destination := os.Args[2]

	err := os.Rename(source, destination)
	if err != nil {
		return 1, err
	}

	fmt.Printf("renamed %s -> %s\n", source, destination)
	return 0, nil
}
