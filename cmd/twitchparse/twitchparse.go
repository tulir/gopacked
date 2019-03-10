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
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"maunium.net/go/gopacked/lib/archive"
	"maunium.net/go/gopacked/lib/log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	flag "maunium.net/go/mauflag"

	"maunium.net/go/gopacked/lib/gopacked"
)

var inputPath = flag.MakeFull("i", "input", "The Twitch modpack as a zip file to read.", "").String()
var outputPath = flag.MakeFull("o", "output", "The file to output the modpack to.", "modpack.json").String()
var extraOutputPath = flag.MakeFull("e", "extra-output", "The directory to output extra files that need to to be served under --web-prefix.", "modpackextra").String()
var webPrefix = flag.MakeFull("w", "web-prefix", "The URL prefix for files that need to be hosted somewhere (e.g. https://example.com/modpack)", "https://example.com/modpack").String()
var wantHelp, _ = flag.MakeHelpFlag()

const help = `goPacked Twitch modpack parser v0.1.0

Usage:
  twitchparse [-h] <-i PATH> [-o PATH] [-w HOST]

Help options:
  -h, --help            Show this help page.

Application options:
  -i, --input=PATH         The Twitch modpack as a zip file to read.
  -o, --output=PATH        The file to output the modpack to.
  -e, --extra-output=PATH  The directory to output extra files that need to to
                           be served under --web-prefix.
  -w, --web-prefix=HOST    The URL prefix for files that need to be hosted
                           somewhere (e.g. https://example.com/modpack).`

func main() {
	flag.SetHelpTitles("goPacked Twitch modpack parser v0.1.0",
		"twitchparse [-h] <-i PATH> [-o PATH] [-w HOST]")
	err := flag.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stdout, help)
		os.Exit(1)
	} else if *wantHelp {
		fmt.Fprintln(os.Stdout, help)
		os.Exit(0)
	}

	log.Infof("Reading input zip")
	reader, err := zip.OpenReader(*inputPath)
	if err != nil {
		panic(err)
	}

	packFiles := map[string]gopacked.FileEntry{}

	tempPath := "/tmp/gopacked/" + strconv.FormatInt(time.Now().Unix(), 16)
	err = os.MkdirAll(tempPath, 0755)
	if err != nil {
		panic(err)
	}

	log.Infof("Looking for manifest.json")
	var packManifest TwitchManifest
	for _, file := range reader.File {
		if file.Name == "manifest.json" {
			reader, err := file.Open()
			if err != nil {
				panic(err)
			}
			manifest, err := ioutil.ReadAll(reader)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(manifest, &packManifest)
			if err != nil {
				panic(err)
			}
		} else {
			err = archive.UnarchiveZipFile(file, tempPath)
			if err != nil {
				panic(err)
			}
		}
	}

	if len(packManifest.OverridesDir) > 0 {
		log.Infof("Extracting overrides and converting them to goPack format")
		eop, err := filepath.Abs(*extraOutputPath)
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll(eop, 0755)
		if err != nil {
			panic(err)
		}

		overridesPath := filepath.Join(tempPath, packManifest.OverridesDir)
		overrides, err := ioutil.ReadDir(overridesPath)
		if err != nil {
			panic(err)
		}
		for _, file := range overrides {
			fileTempPath := filepath.Join(overridesPath, file.Name())
			if file.IsDir() {
				if file.Name() != "mods" {
					outputFile, err := os.OpenFile(filepath.Join(eop, file.Name() + ".zip"), os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						panic(err)
					}
					zipFile := zip.NewWriter(outputFile)
					err = archive.MakeZip(zipFile, fileTempPath, "")
					_ = zipFile.Close()
					_ = outputFile.Close()
					if err != nil {
						panic(err)
					}
					packFiles["override/" + file.Name()] = gopacked.FileEntry{
						Type: gopacked.TypeZipArchive,
						FileName: file.Name(),
						URL: *webPrefix + "/" + file.Name() + ".zip",
					}
				} else {
					modOverrides, err := ioutil.ReadDir(fileTempPath)
					if err != nil {
						panic(err)
					}
					modOutputDir := filepath.Join(eop, "mods")
					err = os.MkdirAll(modOutputDir, 0755)
					if err != nil {
						panic(err)
					}
					for _, modOverride := range modOverrides {
						modTempPath := filepath.Join(fileTempPath, modOverride.Name())
						if modOverride.IsDir() {
							outputFile, err := os.OpenFile(filepath.Join(modOutputDir, modOverride.Name() + ".zip"), os.O_CREATE|os.O_WRONLY, 0644)
							if err != nil {
								panic(err)
							}
							zipFile := zip.NewWriter(outputFile)
							err = archive.MakeZip(zipFile, modTempPath, "")
							_ = zipFile.Close()
							_ = outputFile.Close()
							if err != nil {
								panic(err)
							}
							packFiles["override/mods/" + modOverride.Name()] = gopacked.FileEntry{
								Type: gopacked.TypeZipArchive,
								FileName: modOverride.Name(),
								URL: *webPrefix + "/mods/" + modOverride.Name() + ".zip",
							}
						} else {
							packFiles["override/mods/"+modOverride.Name()] = gopacked.FileEntry{
								Type:     gopacked.TypeFile,
								FileName: modOverride.Name(),
								URL:      *webPrefix + "/mods/" + modOverride.Name(),
							}
							err = os.Rename(modTempPath, filepath.Join(modOutputDir, modOverride.Name()))
							if err != nil {
								panic(err)
							}
						}
					}
				}
			} else {
				packFiles["override/"+file.Name()] = gopacked.FileEntry{
					Type:     gopacked.TypeFile,
					FileName: file.Name(),
					URL:      *webPrefix + "/" + file.Name(),
				}
				err = os.Rename(fileTempPath, filepath.Join(eop, file.Name()))
				if err != nil {
					panic(err)
				}
			}
		}
	}

	packManifest.LoadFileData()
	packManifest.LoadModData()

	log.Infof("Looking for Forge version...")
	var forgeVer string
	for _, loader := range packManifest.Minecraft.ModLoaders {
		if strings.HasPrefix(loader.ID, "forge-") {
			forgeVer = packManifest.Minecraft.Version + "-" + loader.ID[len("forge-"):]
		}
	}
	log.Infof("Forge version found: %s", forgeVer)

	log.Infof("Converting mods to goPack format")
	mods := map[string]gopacked.FileEntry{}
	for _, mod := range packManifest.Files {
		mods[mod.ModData.Name] = gopacked.FileEntry{
			Type:     gopacked.TypeFile,
			Version:  gopacked.Version{mod.FileData.ID},
			FileName: mod.FileData.DiskFileName,
			URL:      mod.FileData.URL,
		}
	}
	packFiles["mods"] = gopacked.FileEntry{
		Type:     gopacked.TypeDirectory,
		Children: mods,
	}

	log.Infof("Converting pack info to goPack format")
	simpleName := strings.ToLower(strings.Replace(packManifest.Name, " ", "", -1))
	gopack := gopacked.GoPack{
		Name:        packManifest.Name,
		SimpleName:  simpleName,
		Author:      packManifest.Author,
		Version:     packManifest.Version,
		ForgeVer:    forgeVer,
		UpdateURL:   *webPrefix,
		ProfileArgs: map[string]interface{}{},
		GoPackedMin: gopacked.Version{0, 4, 0, 0},
		MCLVersion: gopacked.FileEntry{
			Type: gopacked.TypeDirectory,
			Children: map[string]gopacked.FileEntry{
				"Version JSON": {
					Type:     gopacked.TypeFile,
					FileName: simpleName + ".json",
					Version:  gopacked.Version{1},
					URL:      *webPrefix + "/version.json",
				},
			},
		},
		Files: gopacked.FileEntry{
			Type: gopacked.TypeDirectory,
			Children: packFiles,
		},
	}

	log.Infof("Marshaling and writing finished goPack file to disk")
	data, err := json.Marshal(&gopack)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(*outputPath, data, 0644)
	if err != nil {
		panic(err)
	}
	log.Infof("All done")
}
