// goPacked - A simple text-based Minecraft modpack manager.
// Copyright (C) 2016 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GoPack is the base struct for a goPacked modpack.
type GoPack struct {
	Name        string                 `json:"name"`
	SimpleName  string                 `json:"simplename"`
	UpdateURL   string                 `json:"update-url"`
	Author      string                 `json:"author"`
	Version     string                 `json:"version"`
	ProfileArgs map[string]interface{} `json:"profile-settings"`
	MCLVersion  FileEntry              `json:"mcl-version"`
	Files       FileEntry              `json:"files"`
}

// InstallProfile installs the profile data into launcher_profiles.json
func (gp GoPack) InstallProfile(path, mcPath string) error {
	profileData, err := ioutil.ReadFile(filepath.Join(mcPath, "launcher_profiles.json"))
	if err != nil {
		return fmt.Errorf("Failed to read file: %s", err)
	}

	profiles, err := gabs.ParseJSON(profileData)
	if err != nil {
		return fmt.Errorf("Failed to parse JSON: %s", err)
	}

	Infof("Adding %s to launcher_profiles.json", gp.Name)

	_, err = profiles.Set(gp.Name, "profiles", gp.Name, "name")
	if err != nil {
		return fmt.Errorf("Failed to edit JSON: %s", err)
	}
	_, err = profiles.Set(path, "profiles", gp.Name, "gameDir")
	if err != nil {
		return fmt.Errorf("Failed to edit JSON: %s", err)
	}
	_, err = profiles.Set(gp.SimpleName, "profiles", gp.Name, "lastVersionId")
	if err != nil {
		return fmt.Errorf("Failed to edit JSON: %s", err)
	}

	for key, value := range gp.ProfileArgs {
		_, err = profiles.Set(value, "profiles", gp.Name, key)
		if err != nil {
			return fmt.Errorf("Failed to edit JSON: %s", err)
		}
	}
	err = ioutil.WriteFile(filepath.Join(mcPath, "launcher_profiles.json"), []byte(profiles.StringIndent("", "  ")), 0644)
	if err != nil {
		return fmt.Errorf("Failed to save file: %s", err)
	}
	return nil
}

// UninstallProfile uninstalls the profile data from launcher_profiles.json
func (gp GoPack) UninstallProfile(path, mcPath string) error {
	profileData, err := ioutil.ReadFile(filepath.Join(mcPath, "launcher_profiles.json"))
	if err != nil {
		return fmt.Errorf("Failed to read file: %s", err)
	}

	profiles, err := gabs.ParseJSON(profileData)
	if err != nil {
		return fmt.Errorf("Failed to parse json: %s", err)
	}

	Infof("Removing %s from launcher_profiles.json", gp.Name)
	err = profiles.Delete("profiles", gp.Name)
	if err != nil {
		return fmt.Errorf("Failed to edit JSON: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(mcPath, "launcher_profiles.json"), []byte(profiles.StringIndent("", "  ")), 0644)
	if err != nil {
		return fmt.Errorf("Failed to save file: %s", err)
	}
	return nil
}

// Install installs the GoPack to the given path and minecraft directory.
func (gp GoPack) Install(path, mcPath, side string) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		Warnf("Failed to get absolute version of %s: %s", path, err)
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		Warnf("Failed to get absolute version of %s: %s", mcPath, err)
	}

	Infof("Installing %[1]s v%[2]s by %[3]s to %[4]s (%[5]s-side)", gp.Name, gp.Version, gp.Author, path, side)

	if side == "client" {
		err = gp.InstallProfile(path, mcPath)
		if err != nil {
			Errorf("Profile install failed: %s", err)
		}

		gp.MCLVersion.Install(filepath.Join(mcPath, "versions", gp.SimpleName), "", side)
	}
	gp.Files.Install(path, "", side)

	Infof("Saving goPack definition to %s", filepath.Join(path, "gopacked.json"))
	err = gp.Save(filepath.Join(path, "gopacked.json"))
	if err != nil {
		Errorf("goPack definition save failed: %s", err)
	}
}

// Update this GoPack.
func (gp GoPack) Update(new GoPack, path, mcPath, side string) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		Warnf("Failed to get absolute version of %s: %s", path, err)
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		Warnf("Failed to get absolute version of %s: %s", mcPath, err)
	}

	Infof("Updating %[1]s by %[3]s to v%[2]s (%[4]s-side)", gp.Name, gp.Version, gp.Author, side)

	if side == "client" {
		err = gp.InstallProfile(path, mcPath)
		if err != nil {
			Errorf("Profile install failed: %s", err)
		}

		gp.MCLVersion.Update(new.MCLVersion, filepath.Join(mcPath, "versions", gp.SimpleName), filepath.Join(mcPath, "versions", new.SimpleName), "", side)
	}
	gp.Files.Update(new.Files, path, path, "", side)

	Infof("Saving goPack definition to %s", filepath.Join(path, "gopacked.json"))
	err = new.Save(filepath.Join(path, "gopacked.json"))
	if err != nil {
		Errorf("goPack definition save failed: %s", err)
	}
}

// Uninstall this GoPack.
func (gp GoPack) Uninstall(path, mcPath, side string) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		Warnf("Failed to get absolute version of %s: %s", path, err)
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		Warnf("Failed to get absolute version of %s: %s", mcPath, err)
	}

	Infof("Uninstalling %[1]s v%[2]s by %[3]s from %[4]s (%[5]s-side)", gp.Name, gp.Version, gp.Author, path, side)

	if side == "client" {
		err = gp.UninstallProfile(path, mcPath)
		if err != nil {
			Errorf("Profile uninstall failed: %s", err)
		}

		gp.MCLVersion.Remove(filepath.Join(mcPath, "versions", gp.SimpleName), "", side)
	}
	gp.Files.Remove(path, "", side)
	os.RemoveAll(path)
}

// Save saves the gopack definion to the given path.
func (gp GoPack) Save(path string) error {
	json, err := json.Marshal(gp)
	if err != nil {
		return fmt.Errorf("Failed to marshal: %s", err)
	}
	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write: %s", err)
	}
	return nil
}
