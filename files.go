package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileEntry ...
type FileEntry struct {
	Type     string               `json:"type"`
	FileName string               `json:"filename, omitempty"`
	Version  string               `json:"version,omitempty"`
	URL      string               `json:"url,omitempty"`
	Children map[string]FileEntry `json:"children,omitempty"`
}

// Install installs the file entry to the given path.
func (fe FileEntry) Install(path, name string) {
	if fe.Type == "directory" {
		if len(fe.FileName) != 0 {
			if fe.FileName != "//" {
				path = filepath.Join(path, fe.FileName)
				fmt.Printf("Creating directory %s\n", fe.FileName)
			}
		} else if len(name) != 0 {
			path = filepath.Join(path, name)
			fmt.Printf("Creating directory %s\n", name)
		}
		os.MkdirAll(path, 0755)
		for key, value := range fe.Children {
			value.Install(path, key)
		}
	} else if fe.Type == "file" {
		if len(fe.FileName) != 0 {
			path = filepath.Join(path, fe.FileName)
		} else {
			split := strings.Split(fe.URL, "/")
			path = filepath.Join(path, split[len(split)-1])
		}
		fmt.Printf("Downloading %[1]s v%[2]s\n", name, fe.Version)
		Download(fe.URL, path)
	}
}
