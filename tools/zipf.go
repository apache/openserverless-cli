// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package tools

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func zipfTool(args []string) error {
	flag := flag.NewFlagSet("zipf", flag.ExitOnError)
	flag.Usage = func() {
		fmt.Println(`nuv -zipf

Zip an action folder. 
If the folder contains a main file, the output file is <folder>.<ext>.zip where the extension is the extension of the main file.

You can specify the output file with the -o option to override the default name.

You can pass a command with the -x option. 
The command is a shell command that is executed on the zipped folder and its output is saved as the output file.

N.B.: The zip file is saved in the same parent folder of the input folder.

Usage:
  nuv -zipf <folder> [-o <zipfile>] [-x <command>]

Options:`)
		flag.PrintDefaults()
	}

	out := flag.String("o", "", "override output file name (default is <folder>[.<ext>].zip)")
	cmd := flag.String("x", "", "command to execute on the zipped folder")
	help := flag.Bool("h", false, "print this help")

	err := flag.Parse(args)
	if err != nil {
		return err
	}

	if *help {
		flag.Usage()
		return nil
	}

	if flag.NArg() != 1 {
		flag.Usage()
		return fmt.Errorf("invalid number of arguments")
	}

	dir := flag.Arg(0)
	if *out == "" {
		*out, err = generateOutputFileName(dir)
		if err != nil {
			return err
		}
	}

	*out = filepath.Join(filepath.Dir(dir), *out)

	buf, err := Zip(dir)
	if err != nil {
		return err
	}

	if *cmd == "" {
		err = os.WriteFile(*out, buf, 0644)
		if err != nil {
			return err
		}
		log.Println(*out)
		return nil
	}

	stdout, err := runCommandWithStdin(*cmd, buf)
	if err != nil {
		return err
	}

	err = os.WriteFile(*out, stdout, 0644)
	if err != nil {
		return err
	}
	log.Println(*out)
	return nil
}

func runCommandWithStdin(cmd string, stdin []byte) ([]byte, error) {
	c := exec.Command(cmd)
	c.Stdin = bytes.NewReader(stdin)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		return nil, err
	}
	// return stdout
	return c.Output()
}

// GenerateOutputFileName generates the name of the zipped folder
// If the output file  is not specified, the output file is `<folder>.<ext>` where the extension can be:
//
// - `.js.zip` if there is a `main.js` in the `<folder>`
// -  `.py.zip` if there is a `__main__.py` in the `<folder>
// - `.go.zip` if there is a `main.go` in the `<folder>`
// - `php.zip` if there is a `main.pho` in the `<folder>`
// - `.java.zip` if there is a `Main.java` in the folder
//
// in general, look for a file with this regexp: `^.*[mM]ain.**\.(.*)$` and use the extension of the main file
// if no main found use just `.zip`
func generateOutputFileName(folder string) (string, error) {
	// List of possible main file extensions and their corresponding output extensions
	mainFileExtensions := map[string]string{
		".js":   ".js.zip",
		".py":   ".py.zip",
		".go":   ".go.zip",
		".php":  ".php.zip",
		".java": ".java.zip",
	}

	// Compile the regular expression pattern
	mainFilePattern := regexp.MustCompile(`^.*[mM]ain.*\.(.*)$`)

	// Check if any file in the folder matches the regexp for main files
	dirEntries, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}

	for _, dirEntry := range dirEntries {
		fullPath := filepath.Join(folder, dirEntry.Name())

		if !dirEntry.IsDir() && isRegularFile(fullPath) && mainFilePattern.MatchString(dirEntry.Name()) {
			ext := filepath.Ext(dirEntry.Name())
			outputExt, ok := mainFileExtensions[ext]
			if ok {
				// remove the extension from the main file name
				return filepath.Base(folder) + outputExt, nil
			}
		}
	}

	// If no main file was found, use ".zip" as the default extension
	return filepath.Base(folder) + ".zip", nil
}

func isRegularFile(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// Zip a directory
func Zip(dir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zwr := zip.NewWriter(buf)
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(filePath string, info os.FileInfo, err error) error {

		// trim the relevant part of the path
		relPath := strings.TrimPrefix(filePath, dir)
		if relPath == "" {
			return nil
		}
		relPath = relPath[1:]
		if err != nil {
			return err
		}

		// create a proper entry
		isLink := (info.Mode() & os.ModeSymlink) == os.ModeSymlink
		header := &zip.FileHeader{
			Name:   relPath,
			Method: zip.Deflate,
		}
		if isLink {
			header.SetMode(0755 | os.ModeSymlink)
			w, err := zwr.CreateHeader(header)
			if err != nil {
				return err
			}
			ln, err := os.Readlink(filePath)
			if err != nil {
				return err
			}
			_, err = w.Write([]byte(ln))
			if err != nil {
				return err
			}
		} else if info.IsDir() {
			header.Name = relPath + "/"
			header.SetMode(0755)
			_, err := zwr.CreateHeader(header)
			if err != nil {
				return err
			}
		} else if info.Mode().IsRegular() {
			header.SetMode(0755)
			w, err := zwr.CreateHeader(header)
			if err != nil {
				return err
			}
			fsFile, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer fsFile.Close()
			_, err = io.Copy(w, fsFile)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = zwr.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
