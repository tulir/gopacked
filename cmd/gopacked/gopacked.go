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
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	flag "maunium.net/go/mauflag"

	"maunium.net/go/gopacked/lib/gopacked"
	"maunium.net/go/gopacked/lib/log"
)

var installPath = flag.MakeFull("p", "path", "The path to save the modpack in.", "").String()
var minecraftPath = flag.MakeFull("m", "minecraft", "The minecraft directory.", "").String()
var side = flag.MakeFull("s", "side", "The side (client or server) to install.", string(gopacked.SideClient)).String()
var wantHelp, _ = flag.MakeHelpFlag()

const help = `goPacked v0.4.0 - Simple command-line Minecraft modpack manager.

Usage:
  gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL/NAME>

Available actions:
  install               Install the modpack from the given URL.
  update                Update the modpack by URL, name or install path.
  uninstall             Uninstall the modpack by URL, name or install path.

Help options:
  -h, --help            Show this help page.

Application options:
  -p, --path=PATH       The path to save the modpack in.
  -m, --minecraft=PATH  The minecraft directory.
  -s, --side=SIDE       The side (client or server) to install.`

func init() {
	flag.SetHelpTitles("goPacked "+gopacked.GPVersion.String()+" - Simple command-line modpack manager.",
		"gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL/NAME>")
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
	if *side != string(gopacked.SideClient) && *side != string(gopacked.SideServer) {
		log.Fatalf("Couldn't recognize side %[1]s!", *side)
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
	var gp gopacked.GoPack
	log.Infof("Fetching goPack definition from %s", flag.Arg(1))
	err := fetchDefinition(&gp, flag.Arg(1))
	if err != nil {
		log.Fatalf("Failed to fetch goPack definition: %s", err)
		return
	}

	if installPath == nil || len(*installPath) == 0 {
		*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
	}

	gp.Install(*installPath, *minecraftPath, gopacked.Side(*side))
}

func updateOrUninstall(action string) {
	if flag.NArg() < 2 && (installPath == nil || len(*installPath) == 0) {
		log.Fatalf("goPack URL or install location not specified!")
		return
	}
	gp, updated, ok := getUpdateDefinitions()
	if !ok {
		return
	}
	if installPath == nil || len(*installPath) == 0 {
		*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
	}

	if action == "update" {
		update(gp, updated)
	} else if action == "uninstall" {
		gp.Uninstall(*installPath, *minecraftPath, gopacked.Side(*side))
	}
}

func getUpdateDefinitions() (gp gopacked.GoPack, updated gopacked.GoPack, ok bool) {
	if flag.NArg() > 1 {
		if strings.HasPrefix(flag.Arg(1), "http") {
			log.Infof("Fetching goPack definition from %s", flag.Arg(1))
			err := fetchDefinition(&updated, flag.Arg(1))
			if err != nil {
				log.Fatalf("Failed to fetch goPack definition: %s", err)
				ok = false
			}
		} else {
			*installPath = filepath.Join(*minecraftPath, "gopacked", flag.Arg(1))
			log.Infof("Reading goPack definition from %s", *installPath)
			err := readDefinition(&gp, *installPath)
			if err != nil {
				log.Fatalf("Failed to read goPack definition: %s", err)
				ok = false
			}
		}
	} else {
		log.Infof("Reading goPack definition from %s", *installPath)
		err := readDefinition(&gp, *installPath)
		if err != nil {
			log.Fatalf("Failed to read goPack definition: %s", err)
			ok = false
		}
	}
	ok = true
	return
}

func update(gp, updated gopacked.GoPack) {
	if len(gp.Name) == 0 {
		log.Infof("Reading installed goPack definition from %s", *installPath)
		err := readDefinition(&gp, *installPath)
		if err != nil {
			log.Fatalf("Failed to read local goPack definition: %s", err)
		}
	}

	if len(updated.Name) == 0 {
		log.Infof("Fetching updated goPack definition from %s", gp.UpdateURL)
		err := fetchDefinition(&updated, gp.UpdateURL)
		if err != nil {
			log.Fatalf("Failed to updated goPack definition: %s", err)
		}
	}

	gp.Update(updated, *installPath, *minecraftPath, gopacked.Side(*side))
}

func fetchDefinition(gp *gopacked.GoPack, rawURL string) error {
	fromURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if len(fromURL.Scheme) == 0 {
		fromURL.Scheme = "http"
	}
	response, err := http.Get(fromURL.String())
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

func readDefinition(gp *gopacked.GoPack, path string) error {
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
