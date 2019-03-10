// goPacked - A simple text-based Minecraft modpack manager.
// Copyright (C) 2019 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"io"
	"maunium.net/go/gopacked/lib/gopacked"
	"maunium.net/go/gopacked/log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileEntry contains the data of a file or directory.
type FileEntry struct {
	Type     string               `json:"type"`
	FileName string               `json:"filename,omitempty"`
	Version  string               `json:"version,omitempty"`
	Side     string               `json:"side,omitempty"`
	URL      string               `json:"url,omitempty"`
	Children map[string]FileEntry `json:"children,omitempty"`
}

// The possible file types
const (
	TypeFile       string = "file"
	TypeDirectory  string = "directory"
	TypeZIPArchive string = "zip-archive"
)

// Install installs the file entry to the given path.
func (fe FileEntry) Install(path, name, side string) {
	if !fe.checkSide(side) {
		return
	}
	if fe.Type == TypeDirectory {
		log.Infof("Creating directory %s", name)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Warnf("Failed to create %[1]s: %[2]s", name, err)
		}
		for key, value := range fe.Children {
			value.Install(value.path(path, key), key, side)
		}
	} else if fe.Type == TypeFile {
		log.Infof("Downloading %[1]s v%[2]s", name, fe.Version)
		err := downloadFile(fe.URL, path)
		if err != nil {
			log.Errorf("Failed to download %[1]s: %[2]s", name, err)
		}
	} else if fe.Type == TypeZIPArchive {
		log.Infof("Downloading and unzipping %[1]s v%[2]s", name, fe.Version)
		err := os.MkdirAll(path, 0755)
		if err != nil {
			log.Warnf("Failed to create directory for %[1]s: %[2]s", name, err)
			return
		}
		archivePath := filepath.Join(path, "temp-archive.zip")
		err = downloadFile(fe.URL, archivePath)
		if err != nil {
			log.Errorf("Failed to download %[1]s: %[2]s", name, err)
			return
		}
		err = unzip(archivePath, path)
		if err != nil {
			log.Errorf("Failed to unzip %[1]s: %[2]s", name, err)
		}
		err = os.Remove(archivePath)
		if err != nil {
			log.Warnf("Failed to remove temp archive file: %[1]s", err)
		}
	}
}

// Remove removes the given FileEntry from the given path.
func (fe FileEntry) Remove(path, name, side string) {
	if !fe.checkSide(side) {
		return
	}
	if fe.Type == TypeDirectory || fe.Type == TypeZIPArchive {
		log.Infof("Removing %[1]s...", path)
		err := os.RemoveAll(path)
		if err != nil {
			log.Errorf("Failed to remove %[1]s: %[2]s", path, err)
		}
	} else if fe.Type == TypeFile {
		log.Infof("Removing %[1]s v%[2]s...", name, fe.Version)
		err := os.Remove(path)
		if err != nil {
			log.Errorf("Failed to remove %[1]s (%[3]s): %[2]s", name, err, path)
		}
	}
}

// Update updates this FileEntry to the given new version.
func (fe FileEntry) Update(new FileEntry, path, newpath, name, side string) {
	if !fe.checkSide(side) {
		return
	}
	if fe.Type == TypeDirectory {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Infof("Creating directory %s", name)
			err = os.MkdirAll(path, 0755)
			if err != nil {
				log.Warnf("Failed to create directory %s", name)
			}
		}
		// Loop through the old file list. This loop updates outdated files and removes files that are no longer
		// in the updated modpack definition.
		for key, value := range fe.Children {
			newVal, ok := new.Children[key]
			if ok {
				// File already exists, call Update
				value.Update(newVal, value.path(path, key), newVal.path(path, key), key, side)
			} else {
				// File no longer exists, call Remove
				value.Remove(value.path(path, key), key, side)
			}
		}

		// Loop through the new file list. This loop installs new files that did not exist before.
		for key, value := range new.Children {
			_, ok := fe.Children[key]
			if !ok {
				// File didn't exist before, call Install
				value.Install(value.path(path, key), key, side)
			}
		}
	} else if fe.Type == TypeFile || fe.Type == TypeZIPArchive {
		// Compare the versions of the new and old file.
		compare, err := gopacked.ParseAndCompare(new.Version, fe.Version)
		if err != nil {
			log.Errorf("Failed to parse version entry of %s", name)
		}

		// If the version number of the new file is different from the current one, upgrade (or downgrade) it.
		if compare == 1 {
			log.Infof("Updating %[1]s from v%[2]s to v%[3]s", name, fe.Version, new.Version)
		} else if compare == -1 {
			log.Infof("Downgrading %[1]s from v%[2]s to v%[3]s", name, fe.Version, new.Version)
		} else {
			return
		}

		if fe.Type == TypeFile {
			err = os.Remove(path)
			if err != nil {
				log.Warnf("Failed to remove file at %[1]s: %[2]s", path, err)
			}
			err = downloadFile(new.URL, newpath)
			if err != nil {
				log.Errorf("Failed to download %[1]s: %[2]s", name, err)
			}
		} else if fe.Type == TypeZIPArchive {
			err = os.RemoveAll(path)
			if err != nil {
				log.Warnf("Failed to remove directory at %[1]s: %[2]s", path, err)
			}
			err := os.MkdirAll(newpath, 0755)
			if err != nil {
				log.Warnf("Failed to create directory for %[1]s: %[2]s", name, err)
				return
			}
			archivePath := filepath.Join(newpath, "temp-archive.zip")
			err = downloadFile(fe.URL, archivePath)
			if err != nil {
				log.Errorf("Failed to download %[1]s: %[2]s", name, err)
				return
			}
			err = unzip(archivePath, newpath)
			if err != nil {
				log.Errorf("Failed to unzip %[1]s: %[2]s", name, err)
			}
			err = os.Remove(archivePath)
			if err != nil {
				log.Warnf("Failed to remove temp archive file: %[1]s", err)
			}
		}
	}
}

func (fe FileEntry) path(path, name string) string {
	if fe.Type == TypeDirectory || fe.Type == TypeZIPArchive {
		if len(fe.FileName) != 0 {
			if fe.FileName != "//" {
				path = filepath.Join(path, fe.FileName)
			}
		} else if len(name) != 0 {
			path = filepath.Join(path, name)
		}
	} else if fe.Type == TypeFile {
		if len(fe.FileName) != 0 {
			path = filepath.Join(path, fe.FileName)
		} else {
			split := strings.Split(fe.URL, "/")
			path = filepath.Join(path, split[len(split)-1])
		}
	}
	return path
}

func (fe FileEntry) checkSide(side string) bool {
	return len(fe.Side) == 0 || side == fe.Side || side == "both"
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
