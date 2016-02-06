package main

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// GoPack is the base struct for a GoPacked modpack.
type GoPack struct {
	Name        string                 `json:"name"`
	SimpleName  string                 `json:"simplename"`
	Author      string                 `json:"author"`
	Version     string                 `json:"version"`
	ProfileArgs map[string]interface{} `json:"profile-settings"`
	MCLVersion  FileEntry              `json:"mcl-version"`
	Files       FileEntry              `json:"files"`
}

// Install installs the GoPack to the given path and minecraft directory.
func (gp GoPack) Install(path, mcPath string) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	mcPath, err = filepath.Abs(mcPath)
	if err != nil {
		panic(err)
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	if !strings.HasSuffix(mcPath, "/") {
		mcPath = mcPath + "/"
	}

	fmt.Printf("Installing %[1]s v%[2]s by %[3]s to %[4]s\n", gp.Name, gp.Version, gp.Author, path)
	gp.MCLVersion.Install(mcPath+"versions/"+gp.SimpleName, "")

	profileData, err := ioutil.ReadFile(mcPath + "launcher_profiles.json")
	if err != nil {
		panic(err)
	}

	profiles, err := gabs.ParseJSON(profileData)
	if err != nil {
		panic(err)
	}

	profiles.Set(gp.Name, "profiles", gp.Name, "name")
	profiles.Set(path, "profiles", gp.Name, "gameDir")
	profiles.Set(gp.SimpleName, "profiles", gp.Name, "lastVersionId")
	for key, value := range gp.ProfileArgs {
		profiles.Set(value, "profiles", gp.Name, key)
	}
	println(profiles.StringIndent("", "  "))
	err = ioutil.WriteFile(mcPath+"launcher_profiles.json", []byte(profiles.StringIndent("", "  ")), 0644)
	if err != nil {
		panic(err)
	}

	gp.Files.Install(path, "")
	gp.Save(path + "gopacked.json")
}

// Update updates this GoPack.
func (gp GoPack) Update(path, mcPath string) {
	// TODO Implement me
}

// Save saves the gopack definion to the given path.
func (gp GoPack) Save(path string) {
	json, err := json.Marshal(gp)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(path, json, 0644)
	if err != nil {
		panic(err)
	}
}
