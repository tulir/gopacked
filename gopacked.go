package main

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

// GoPack is the base struct for a GoPacked modpack.
type GoPack struct {
	Name        string            `json:"name"`
	SimpleName  string            `json:"simplename"`
	Author      string            `json:"author"`
	Version     string            `json:"version"`
	ProfileArgs map[string]string `json:"profile-settings"`
	MCLVersion  FileEntry         `json:"mcl-version"`
	Files       FileEntry         `json:"files"`
}

func (mp GoPack) install(path, mcPath string) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	if !strings.HasSuffix(mcPath, "/") {
		mcPath = mcPath + "/"
	}

	fmt.Printf("Installing %[1]s v%[2]s by %[3]s to %[4]s\n", mp.Name, mp.Version, mp.Author, path)
	mp.MCLVersion.install(mcPath+"versions/"+mp.SimpleName, "")

	profileData, err := ioutil.ReadFile(mcPath + "launcher_profiles.json")
	if err != nil {
		panic(err)
	}

	profiles, err := gabs.ParseJSON(profileData)
	if err != nil {
		panic(err)
	}

	packProfile := gabs.New()
	packProfile.Set(mp.Name, "name")
	packProfile.Set(path, "gameDir")
	packProfile.Set(mp.SimpleName, "lastVersionId")
	for key, value := range mp.ProfileArgs {
		packProfile.Set(value, key)
	}
	profiles.Set(packProfile, "profiles", mp.Name)
	println(profiles.StringIndent("", "  "))
	err = ioutil.WriteFile(mcPath+"launcher_profiles.json", []byte(profiles.StringIndent("", "  ")), 0644)
	if err != nil {
		panic(err)
	}

	mp.Files.install(path, "")
}

// FileEntry ...
type FileEntry struct {
	Type     string               `json:"type"`
	FileName string               `json:"filename, omitempty"`
	Version  string               `json:"version,omitempty"`
	URL      string               `json:"url,omitempty"`
	Children map[string]FileEntry `json:"children,omitempty"`
}

func (fe FileEntry) install(path, name string) {
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	if fe.Type == "directory" {
		if len(fe.FileName) != 0 {
			if fe.FileName != "//" {
				path = path + fe.FileName
				fmt.Printf("Creating directory %s\n", fe.FileName)
			}
		} else if len(name) != 0 {
			path = path + name
			fmt.Printf("Creating directory %s\n", name)
		}
		os.MkdirAll(path, 0755)
		for key, value := range fe.Children {
			value.install(path, key)
		}
	} else if fe.Type == "file" {
		if len(fe.FileName) != 0 {
			path = path + fe.FileName
		} else {
			split := strings.Split(fe.URL, "/")
			path = path + split[len(split)-1]
		}
		fmt.Printf("Downloading %[1]s v%[2]s\n", name, fe.Version)
		Download(fe.URL, path)
	}
}

var installPath = flag.StringP("path", "p", "", "")
var minecraftPath = flag.StringP("minecraft", "m", "", "")

var help = `Simple command-line modpack manager.

Usage:
  gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL>

Help options:
  -h, --help               Show this help page

Application options:
  -p, --path=PATH          The path to save the modpack in.
  -m, --minecraft=PATH     The minecraft directory (for things like launcher-profiles.json)
`

func init() {
	flag.Usage = func() {
		println(help)
	}
	flag.Parse()
	if flag.NArg() < 2 {
		panic(fmt.Errorf("Not enough arguments"))
	}

	if minecraftPath == nil || len(*minecraftPath) == 0 {
		switch strings.ToLower(runtime.GOOS) {
		case "windows":
			*minecraftPath = os.Getenv("APPDATA") + "./minecraft"
		case "darwin":
			*minecraftPath = os.Getenv("HOME") + "/Library/Application Support/minecraft"
		default:
			*minecraftPath = os.Getenv("HOME") + "/.minecraft"
		}
	}
	if !strings.HasSuffix(*minecraftPath, "/") {
		*minecraftPath = *minecraftPath + "/"
	}
}

func main() {
	if strings.ToLower(flag.Arg(0)) == "install" {
		fmt.Println("Fetching goPack definition from", flag.Arg(1))
		data := HTTPGet(flag.Arg(1))
		if len(data) == 0 {
			panic(fmt.Errorf("No data received!"))
		}
		var mp GoPack
		err := json.Unmarshal(data, &mp)
		println(mp.Name)
		if err != nil {
			panic(err)
		}
		mp.install(*installPath, *minecraftPath)
	}
}
