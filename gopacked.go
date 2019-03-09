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
	flag "maunium.net/go/mauflag"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var installPath = flag.Make().LongKey("path").ShortKey("p").String()
var minecraftPath = flag.Make().LongKey("minecraft").ShortKey("m").String()
var wantHelp = flag.Make().LongKey("help").ShortKey("h").Bool()

var side = flag.Make().LongKey("side").ShortKey("s").Default(CLIENT).String()

var version = Version{0, 3, 0, 0}

// Side constants
const (
	CLIENT = "client"
	SERVER = "server"
)

const help = `goPacked 0.3 - Simple command-line modpack manager.

Usage:
  gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL/NAME>

Available actions:
  install                  Install the modpack from the given URL.
  update                   Update the modpack by URL, name or install path.
  uninstall                Uninstall the modpack by URL, name or install path.

Help options:
  -h, --help               Show this help page.

Application options:
  -p, --path=PATH          The path to save the modpack in.
  -m, --minecraft=PATH     The minecraft directory.
  -s, --side=SIDE          The side (client or server) to install.`

func init() {
	err := flag.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stdout, help)
		os.Exit(1)
	} else if *wantHelp {
		fmt.Fprintln(os.Stdout, help)
		os.Exit(0)
	}

	if minecraftPath == nil || len(*minecraftPath) == 0 {
		switch strings.ToLower(runtime.GOOS) {
		case "windows":
			*minecraftPath = filepath.Join(os.Getenv("APPDATA"), ".minecraft")
		case "darwin":
			*minecraftPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "minecraft")
		default:
			*minecraftPath = filepath.Join(os.Getenv("HOME"), ".minecraft")
		}
	}

	*side = strings.ToLower(*side)
	if *side != CLIENT && *side != SERVER {
		Fatalf("Couldn't recognize side %[1]s!", *side)
		os.Exit(2)
	}
}

func main() {
	if *side == "server" && runtime.GOOS != "windows" {
		*minecraftPath = os.Getenv("HOME")
	}

	action := strings.ToLower(flag.Arg(0))
	if action == "install" && flag.NArg() > 1 {
		install()
	} else if action == "uninstall" || action == "update" {
		updateOrUninstall(action)
	} else {
		fmt.Fprintln(os.Stdout, help)
	}
}

func install() {
	var gp GoPack
	Infof("Fetching goPack definition from %s", flag.Arg(1))
	err := fetchDefinition(&gp, flag.Arg(1))
	if err != nil {
		Fatalf("Failed to fetch goPack definition: %s", err)
		return
	}

	if installPath == nil || len(*installPath) == 0 {
		*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
	}

	gp.Install(*installPath, *minecraftPath, *side)
}

func updateOrUninstall(action string) {
	if flag.NArg() < 2 && (installPath == nil || len(*installPath) == 0) {
		Fatalf("goPack URL or install location not specified!")
		return
	}
	var gp, updated = getUpdateDefinitions()
	if installPath == nil || len(*installPath) == 0 {
		*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
	}

	if action == "update" {
		update(gp, updated)
	} else if action == "uninstall" {
		gp.Uninstall(*installPath, *minecraftPath, *side)
	}
}

func getUpdateDefinitions() (GoPack, GoPack) {
	var gp, updated GoPack
	if flag.NArg() > 1 {
		if strings.HasPrefix(flag.Arg(1), "http") {
			Infof("Fetching goPack definition from %s", flag.Arg(1))
			err := fetchDefinition(&updated, flag.Arg(1))
			if err != nil {
				Fatalf("Failed to fetch goPack definition: %s", err)
			}
		} else {
			*installPath = filepath.Join(*minecraftPath, "gopacked", flag.Arg(1))
			Infof("Reading goPack definition from %s", *installPath)
			err := readDefinition(&gp, *installPath)
			if err != nil {
				Fatalf("Failed to read goPack definition: %s", err)
			}
		}
	} else {
		Infof("Reading goPack definition from %s", *installPath)
		err := readDefinition(&gp, *installPath)
		if err != nil {
			Fatalf("Failed to read goPack definition: %s", err)
		}
	}
	return gp, updated
}

func update(gp, updated GoPack) {
	if len(gp.Name) == 0 {
		Infof("Reading installed goPack definition from %s", *installPath)
		err := readDefinition(&gp, *installPath)
		if err != nil {
			Fatalf("Failed to read local goPack definition: %s", err)
		}
	}

	if len(updated.Name) == 0 {
		Infof("Fetching updated goPack definition from %s", gp.UpdateURL)
		err := fetchDefinition(&updated, gp.UpdateURL)
		if err != nil {
			Fatalf("Failed to updated goPack definition: %s", err)
		}
	}

	gp.Update(updated, *installPath, *minecraftPath, *side)
}

func fetchDefinition(gp *GoPack, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("no data received")
	}

	err = json.Unmarshal(data, &gp)
	if err != nil {
		return err
	}
	return nil
}

func readDefinition(gp *GoPack, path string) error {
	data, err := ioutil.ReadFile(filepath.Join(path, "gopacked.json"))
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, gp)
	if err != nil {
		return err
	}
	return nil
}
