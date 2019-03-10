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
	"bytes"
	"encoding/json"
	"fmt"
	"maunium.net/go/gopacked/lib/gopacked"
	"maunium.net/go/gopacked/log"
	"net/http"
	"strings"
)

const CurseMetaURL = "https://staging_cursemeta.dries007.net/api/v3"
const CurseMetaUserAgent = "maunium.net/go/gopacked/twitchparse"

type TwitchManifest struct {
	Minecraft       TwitchMinecraftInfo `json:"minecraft"`
	ManifestType    string              `json:"manifestType"`
	ManifestVersion int                 `json:"manifestVersion"`
	Name            string              `json:"name"`
	Version         gopacked.Version    `json:"version"`
	Author          string              `json:"author"`
	Files           []*TwitchFile        `json:"files"`
	OverridesDir    string              `json:"overrides"`
}

type TwitchMinecraftInfo struct {
	Version    string `json:"version"`
	ModLoaders []struct {
		ID      string `json:"id"`
		Primary bool   `json:"primary"`
	} `json:"modLoaders"`
}

type TwitchFile struct {
	ProjectID int           `json:"projectID"`
	FileID    int           `json:"fileID"`
	Required  bool          `json:"required"`
	FileData  *CurseMetaFile `json:"fileData,omitempty"`
	ModData   *CurseMetaMod  `json:"modData,omitempty"`
}

type CurseMetaMod struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Summary string `json:"summary"`
	URL     string `json:"websiteUrl"`
	Website string `json:"externalUrl"`

	Authors []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	PrimaryAuthorName string `json:"primaryAuthorName"`
}

type CurseMetaFile struct {
	ID int `json:"id"`

	FileName     string `json:"fileName"`
	DiskFileName string `json:"fileNameOnDisk"`
	FileDate     string `json:"fileDate"`
	FileLength   int    `json:"length"`
	FileStatus   int    `json:"fileStatus"`
	URL          string `json:"downloadUrl"`

	ReleaseType     int  `json:"releaseType"`
	IsAvailable     bool `json:"isAvailable"`
	IsAlternate     bool `json:"isAlternate"`
	AlternateFileID int  `json:"alternateFileId"`

	Dependencies []struct {
		AddonID int `json:"addonId"`
		Type    int `json:"type"`
	} `json:"dependencies"`

	Modules []struct {
		FolderName  string `json:"folderName"`
		Fingerprint int64  `json:"fingerprint"`
	}

	PackageFingerprint int64 `json:"packageFingerprint"`

	GameVersion     []gopacked.Version `json:"gameVersion"`
	InstallMetadata interface{}        `json:"installMetadata"`
}

func (tm TwitchManifest) LoadFileData() {
	log.Infof("Loading file data for manifest")
	manifest, err := json.Marshal(&tm)
	req, err := http.NewRequest(http.MethodPost, CurseMetaURL+"/manifest", bytes.NewReader(manifest))
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", CurseMetaUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tm)
	if err != nil {
		panic(err)
	}
	log.Infof("File data loaded")
}

func (tm TwitchManifest) LoadModData() {
	log.Infof("Loading mod data for manifest")
	ids := make([]string, len(tm.Files))
	for i, mod := range tm.Files {
		ids[i] = fmt.Sprintf("id=%d&", mod.ProjectID)
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/direct/addon?%s", CurseMetaURL, strings.Join(ids, "&")), nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", CurseMetaUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var respData []*CurseMetaMod
	err = dec.Decode(&respData)
	if err != nil {
		panic(err)
	}

	fileMap := make(map[int][]*TwitchFile)
	for _, file := range tm.Files {
		arr, ok := fileMap[file.ProjectID]
		if !ok {
			arr = []*TwitchFile{file}
		} else {
			arr = append(arr, file)
		}
		fileMap[file.ProjectID] = arr
	}

	for _, mod := range respData {
		for _, file := range fileMap[mod.ID] {
			file.ModData = mod
		}
	}
	log.Infof("Mod data loaded")
}
