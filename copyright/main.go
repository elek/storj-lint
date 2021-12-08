// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package copyright

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeebo/errs"
)

var checkFiles = map[string]bool{
	".go":  true,
	".ts":  true,
	".js":  true,
	".vue": true,
}

var ignoreFolder = map[string]bool{
	".git":         true,
	"node_modules": true,
	"coverage":     true,
	"dist":         true,
}

func CheckCopyright() error {
	var failed int

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if info.IsDir() && ignoreFolder[info.Name()] {
			return filepath.SkipDir
		}
		if !checkFiles[filepath.Ext(path)] {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			failed++
			fmt.Printf("failed to read %v: %v\n", path, err)
			return nil
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		var header [256]byte
		n, err := file.Read(header[:])
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("failed to read %v: %v\n", path, err)
			return nil
		}

		if bytes.Contains(header[:n], []byte(`AUTOGENERATED`)) ||
			bytes.Contains(header[:n], []byte(`Code generated`)) ||
			bytes.Contains(header[:n], []byte(`Autogenerated`)) {
			return nil
		}

		if !bytes.Contains(header[:n], []byte(`Copyright `)) {
			failed++
			fmt.Printf("missing copyright: %v\n", path)
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	if failed > 0 {
		return errs.New("One or more file has wrong copyright")
	}
	return nil
}