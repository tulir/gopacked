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
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var installPath = flag.StringP("path", "p", "", "")
var minecraftPath = flag.StringP("minecraft", "m", "", "")

var side = flag.StringP("side", "s", "client", "")

var help = `goPacked 0.2 - Simple command-line modpack manager.

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
	flag.Usage = func() {
		println(help)
	}
	flag.Parse()

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
	if *side != "client" && *side != "server" {
		Fatalf("Couldn't recognize side %[1]s!", *side)
		os.Exit(1)
	}
}

func main() {
	if *side == "server" && runtime.GOOS != "windows" {
		*minecraftPath = os.Getenv("HOME")
	}

	action := strings.ToLower(flag.Arg(0))
	var gp GoPack
	if action == "install" && flag.NArg() > 1 {
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
	} else if action == "uninstall" || action == "update" {
		if flag.NArg() < 2 && (installPath == nil || len(*installPath) == 0) {
			Fatalf("goPack URL or install location not specified!")
			return
		}

		var updated GoPack
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

		if installPath == nil || len(*installPath) == 0 {
			*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
		}

		if action == "update" {
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
		} else if action == "uninstall" {
			gp.Uninstall(*installPath, *minecraftPath, *side)
		}
	} else {
		flag.Usage()
	}
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
	response.Body.Close()

	if len(data) == 0 {
		return fmt.Errorf("No data received!")
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
