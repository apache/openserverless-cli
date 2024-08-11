package tools

import (
	"fmt"
	"os"
	"strings"
)

func Executable() (int, error) {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Make a file executable:")
		fmt.Println(" chmod +x in Unix, rename to .exe in Windows")
		fmt.Println("Usage: executable <file>")
		return 0, nil
	}

	file := args[0]

	// Get the current file permissions
	info, err := os.Stat(file)
	if err != nil {
		return 1, err
	}

	// Add execute permissions for the owner
	if GetOS() == "windows" {
		if !strings.HasSuffix(strings.ToLower(file), ".exe") {
			fileexe := file + ".exe"
			err = os.Rename(file, fileexe)
			if err != nil {
				return 1, err
			}
			fmt.Printf("Successfully renamed %s to %s\n", file, fileexe)
			return 0, nil
		} else {
			fmt.Println("Nothing to do")
			return 0, nil
		}
	} else {
		err = os.Chmod(file, info.Mode()|0100)
		if err != nil {
			return 1, err
		}
		fmt.Printf("Successfully added execute permissions to %s\n", file)
		return 0, nil
	}
}
