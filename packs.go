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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"maunium.net/go/gopacked/lib/gopacked"
	"maunium.net/go/gopacked/log"
	"os"
	"os/exec"
	"path/filepath"
)

// GoPack is the base struct for a goPacked modpack.
type GoPack struct {
	Name        string                 `json:"name"`
	SimpleName  string                 `json:"simplename"`
	UpdateURL   string                 `json:"update-url"`
	Author      string                 `json:"author"`
	Version     string                 `json:"version"`
	ForgeVer    string                 `json:"forge-version,omitempty"`
	GoPackedMin string                 `json:"gopacked-version-minimum,omitempty"`
	GoPackedMax string                 `json:"gopacked-version-maximum,omitempty"`
	ProfileArgs map[string]interface{} `json:"profile-settings"`
	MCLVersion  FileEntry              `json:"mcl-version"`
	Files       FileEntry              `json:"files"`
}

func readJSON(file string) (val map[string]interface{}, err error) {
	var data []byte
	data, err = ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &val)
	return
}

func writeJSON(obj interface{}, file string) error {
	data, err := json.Marshal(&obj)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, data, 0644)
}

// InstallProfile installs the profile data into launcher_profiles.json
func (gp GoPack) InstallProfile(path, mcPath string) error {
	launcherProfiles, err := readJSON(filepath.Join(mcPath, "launcher_profiles.json"))
	if err != nil {
		return fmt.Errorf("failed to read launcher_profiles.json: %s", err)
	}

	log.Infof("Adding %s to launcher_profiles.json", gp.Name)

	profiles := launcherProfiles["profiles"].(map[string]interface{})
	profile := map[string]interface{}{
		"name":          gp.Name,
		"gameDir":       path,
		"lastVersionId": gp.SimpleName,
	}
	for key, value := range gp.ProfileArgs {
		profile[key] = value
	}
	profiles[gp.Name] = profile

	err = writeJSON(launcherProfiles, filepath.Join(mcPath, "launcher_profiles.json"))
	if err != nil {
		return fmt.Errorf("failed to save file: %s", err)
	}
	return nil
}

// UninstallProfile uninstalls the profile data from launcher_profiles.json
func (gp GoPack) UninstallProfile(path, mcPath string) error {
	launcherProfiles, err := readJSON(filepath.Join(mcPath, "launcher_profiles.json"))
	if err != nil {
		return fmt.Errorf("failed to read launcher_profiles.json: %s", err)
	}

	profiles := launcherProfiles["profiles"].(map[string]interface{})
	delete(profiles, gp.Name)

	err = writeJSON(launcherProfiles, filepath.Join(mcPath, "launcher_profiles.json"))
	if err != nil {
		return fmt.Errorf("failed to save file: %s", err)
	}
	return nil
}

// InstallForge installs the required version of forge for this gopack.
func (gp GoPack) InstallForge(path, mcPath, side string) {
	if len(gp.ForgeVer) == 0 {
		return
	}

	linec := []rune(log.Inputf("Would you like to install Forge v%s [y/N] ", gp.ForgeVer))
	if linec[0] != 'y' && linec[0] != 'Y' {
		return
	}

	log.Infof("Downloading Forge v%[1]s installer", gp.ForgeVer)

	installerURL := fmt.Sprintf("http://files.minecraftforge.net/maven/net/minecraftforge/forge/%[1]s/forge-%[1]s-installer.jar", gp.ForgeVer)
	installerPath := filepath.Join(path, "forge-installer.jar")
	err := downloadFile(installerURL, installerPath)
	if err != nil {
		log.Errorf("Failed to download Forge installer: %s", err)
		return
	}
	var cmd *exec.Cmd
	if side == CLIENT {
		cmd = exec.Command("java", "-jar", installerPath)
	} else {
		cmd = exec.Command("java", "-jar", installerPath, "--installServer")
	}

	log.Infof("Starting Forge installer...")
	oldDir, _ := os.Getwd()
	_ = os.Chdir(path)
	err = cmd.Run()
	if err != nil {
		log.Warnf("Running command resulted in error: %s", err)
	}
	if len(oldDir) != 0 {
		_ = os.Chdir(oldDir)
	}
	err = os.Remove(installerPath)
	if err != nil {
		log.Warnf("Failed to remove Forge installer: %s", err)
	}
	log.Infof("Forge installer finished")
}

// CheckVersion checks whether or not the goPacked instance is within the version requirements of this goPack.
func (gp GoPack) CheckVersion() bool {
	var continueAsk = false
	if len(gp.GoPackedMax) != 0 {
		gpVer, err := gopacked.ParseVersion(gp.GoPackedMax)
		if err != nil {
			log.Warnf("Failed to parse maximum supported goPacked version")
			continueAsk = true
		} else if gpVer.Compare(version) == 1 {
			log.Warnf("goPacked version greater than maximum supported by requested goPack")
			continueAsk = true
		}
	}
	if len(gp.GoPackedMin) != 0 {
		gpVer, err := gopacked.ParseVersion(gp.GoPackedMin)
		if err != nil {
			log.Warnf("Failed to parse minimum supported goPacked version")
			continueAsk = true
		} else if gpVer.Compare(version) == -1 {
			log.Warnf("goPacked version smaller than minimum supported by requested goPack")
			continueAsk = true
		}
	}
	if continueAsk {
		linec := []rune(log.Inputf("Would you like to continue anyway [y/N]"))
		if linec[0] != 'y' && linec[0] != 'Y' {
			return false
		}
	}
	return true
}

// Install installs the GoPack to the given path and minecraft directory.
func (gp GoPack) Install(path, mcPath, side string) {
	if !gp.CheckVersion() {
		return
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Warnf("Failed to get absolute path of %s: %s", path, err)
	}
	err = os.MkdirAll(path, 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Warnf("Failed to create directory at %s: %s", path, err)
		}
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		log.Warnf("Failed to get absolute path of %s: %s", mcPath, err)
	}

	log.Infof("Installing %[1]s v%[2]s by %[3]s to %[4]s (%[5]s-side)", gp.Name, gp.Version, gp.Author, path, side)

	if side == CLIENT {
		err = gp.InstallProfile(path, mcPath)
		if err != nil {
			log.Errorf("Profile install failed: %s", err)
		}

		gp.MCLVersion.Install(filepath.Join(mcPath, "versions", gp.SimpleName), "", side)
	}
	gp.Files.Install(path, "", side)
	gp.InstallForge(path, mcPath, side)

	log.Infof("Saving goPack definition to %s", filepath.Join(path, "gopacked.json"))
	err = gp.Save(filepath.Join(path, "gopacked.json"))
	if err != nil {
		log.Errorf("goPack definition save failed: %s", err)
	}
}

// Update this GoPack.
func (gp GoPack) Update(new GoPack, path, mcPath, side string) {
	if !new.CheckVersion() {
		return
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Warnf("Failed to get absolute version of %s: %s", path, err)
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		log.Warnf("Failed to get absolute version of %s: %s", mcPath, err)
	}

	log.Infof("Updating %[1]s by %[3]s to v%[2]s (%[4]s-side)", gp.Name, gp.Version, gp.Author, side)

	if side == CLIENT {
		err = gp.InstallProfile(path, mcPath)
		if err != nil {
			log.Errorf("Profile install failed: %s", err)
		}

		gp.MCLVersion.Update(new.MCLVersion, filepath.Join(mcPath, "versions", gp.SimpleName), filepath.Join(mcPath, "versions", new.SimpleName), "", side)
	}
	gp.Files.Update(new.Files, path, path, "", side)
	gp.InstallForge(path, mcPath, side)

	log.Infof("Saving goPack definition to %s", filepath.Join(path, "gopacked.json"))
	err = new.Save(filepath.Join(path, "gopacked.json"))
	if err != nil {
		log.Errorf("goPack definition save failed: %s", err)
	}
}

// Uninstall this GoPack.
func (gp GoPack) Uninstall(path, mcPath, side string) {
	linec := []rune(log.Inputf("Are you sure you wish to uninstall %s v%s [y/N] ", gp.Name, gp.Version))
	if linec[0] != 'y' && linec[0] != 'Y' {
		log.Infof("Uninstall cancelled")
		return
	}

	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		log.Warnf("Failed to get absolute version of %s: %s", path, err)
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		log.Warnf("Failed to get absolute version of %s: %s", mcPath, err)
	}

	log.Infof("Uninstalling %[1]s v%[2]s by %[3]s from %[4]s (%[5]s-side)", gp.Name, gp.Version, gp.Author, path, side)

	if side == CLIENT {
		err = gp.UninstallProfile(path, mcPath)
		if err != nil {
			log.Errorf("Profile uninstall failed: %s", err)
		}

		gp.MCLVersion.Remove(filepath.Join(mcPath, "versions", gp.SimpleName), "", side)
	}
	gp.Files.Remove(path, "", side)
	err = os.RemoveAll(path)
	if err != nil {
		log.Warnf("Failed to remove %s: %s", path, err)
	}
}

// Save saves the gopack definion to the given path.
func (gp GoPack) Save(path string) error {
	data, err := json.Marshal(gp)
	if err != nil {
		return fmt.Errorf("failed to marshal: %s", err)
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write: %s", err)
	}
	return nil
}
