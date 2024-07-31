package tools

import (
	"fmt"
	"os"
)

func Remove() (int, error) {
	if len(os.Args) != 2 {
		fmt.Println("Usage: remove <filename>")
		return 0, nil
	}

	filename := os.Args[1]

	err := os.Remove(filename)
	if err != nil {
		return 1, err
	}

	fmt.Printf("removed %s\n", filename)

	return 0, nil
}
