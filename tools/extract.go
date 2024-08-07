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
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xi2/xz"
)

func ExtractFileFromCompressedTar(filename string, target string) error {
	stream, err := os.Open(filename)
	if err != nil {
		return err
	}
	var uncompressedStream io.Reader = nil
	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		trace("extracting gzip")
		uncompressedStream, err = gzip.NewReader(stream)
		if err != nil {
			return err
		}
	}
	if strings.HasSuffix(filename, ".tar.xz") {
		trace("extracting xz")
		uncompressedStream, err = xz.NewReader(stream, 0)
		if err != nil {
			return err
		}
	}
	if strings.HasSuffix(filename, ".tar.bz2") {
		trace("extracting bzip2")
		uncompressedStream = bzip2.NewReader(stream)
	}
	if strings.HasSuffix(filename, ".tar") {
		trace("extracting tar")
		uncompressedStream = stream
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			trace("skipping dir", header.Name)
			continue
		case tar.TypeReg:
			file := filepath.Base(header.Name)
			if file == target {
				trace("creating file", target)
				flags := os.O_RDWR | os.O_CREATE | os.O_TRUNC
				outFile, err := os.OpenFile(target, flags, 0755)
				if err != nil {
					return err
				}
				_, err = io.Copy(outFile, tarReader)
				outFile.Close()
				return nil
			}
			trace("discarding file", header.Name)
			_, err = io.Copy(io.Discard, tarReader)
			if err == nil {
				continue
			}
			return err
		default:
			return fmt.Errorf("uknown type: %d in %s", header.Typeflag, header.Name)
		}
	}
	return fmt.Errorf("file not found")
}

func ExtractFileFromZip(filename string, target string) error {
	archive, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	trace("opened file", filename)

	defer archive.Close()

	for _, f := range archive.File {

		if f.FileInfo().IsDir() {
			trace("skipping dir", f.Name)
			continue
		}

		file := filepath.Base(f.Name)

		if file != target {
			trace("skipping file", f.Name)
			continue
		}

		trace("writing file", f.Name)
		flags := os.O_RDWR | os.O_CREATE | os.O_TRUNC
		dstFile, err := os.OpenFile(target, flags, 0755)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}
		defer fileInArchive.Close()

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("file not found")
}

func Extract() (int, error) {
	if len(os.Args) < 3 {
		fmt.Println("Extract one single file from a .zip .tar, .tgz, .tar.gz, tar.bz2, tar.gz")
		fmt.Println("Usage: file.(zip|tgz|tar[.gz|.bz2|.xz]) target")
		return 0, nil
	}
	_, err := os.Stat(os.Args[1])
	if os.IsNotExist(err) {
		return 1, err
	}
	if strings.HasSuffix(os.Args[1], ".zip") {
		err = ExtractFileFromZip(os.Args[1], os.Args[2])
	} else {
		err = ExtractFileFromCompressedTar(os.Args[1], os.Args[2])
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}
