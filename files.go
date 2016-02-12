package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileEntry contains the data of a file or directory.
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
		fmt.Printf("Creating directory %s\n", name)
		os.MkdirAll(path, 0755)
		for key, value := range fe.Children {
			value.Install(value.path(path, key), key)
		}
	} else if fe.Type == "file" {
		fmt.Printf("Downloading %[1]s v%[2]s\n", name, fe.Version)
		downloadFile(fe.URL, path)
	}
}

// Remove removes the given FileEntry from the given path.
func (fe FileEntry) Remove(path, name string) {
	if fe.Type == "directory" {
		fmt.Printf("Removing %[1]s...\n", name)
		os.RemoveAll(path)
	} else if fe.Type == "file" {
		fmt.Printf("Removing %[1]s v%[2]s...\n", name, fe.Version)
		os.Remove(path)
	}
}

// Update updates this FileEntry to the given new version.
func (fe FileEntry) Update(new FileEntry, path, newpath, name string) {
	if fe.Type == "directory" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Creating directory %s\n", name)
			os.MkdirAll(path, 0755)
		}
		for key, value := range fe.Children {
			newVal, ok := new.Children[key]
			if ok {
				value.Update(newVal, value.path(path, key), newVal.path(path, key), key)
			} else {
				value.Remove(value.path(path, key), key)
			}
		}
	} else if fe.Type == "file" {
		compare, err := ParseAndCompare(new.Version, fe.Version)
		if err != nil {
			fmt.Printf("Failed to parse version entry of %s\n", name)
		}

		if compare == 1 {
			fmt.Printf("Updating %[1]s from v%[2]s to v%[3]s\n", name, fe.Version, new.Version)
			os.Remove(path)
			downloadFile(new.URL, newpath)
		}
	}
}

func (fe FileEntry) path(path, name string) string {
	if fe.Type == "directory" {
		if len(fe.FileName) != 0 {
			if fe.FileName != "//" {
				path = filepath.Join(path, fe.FileName)
			}
		} else if len(name) != 0 {
			path = filepath.Join(path, name)
		}
	} else if fe.Type == "file" {
		if len(fe.FileName) != 0 {
			path = filepath.Join(path, fe.FileName)
		} else {
			split := strings.Split(fe.URL, "/")
			path = filepath.Join(path, split[len(split)-1])
		}
	}
	return path
}

func downloadFile(url, saveTo string) error {
	out, err := os.Create(saveTo)
	defer out.Close()
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
