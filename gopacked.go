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

var help = `Simple command-line modpack manager.

Usage:
  gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL/NAME>

Help options:
  -h, --help               Show this help page

Application options:
  -p, --path=PATH          The path to save the modpack in.
  -m, --minecraft=PATH     The minecraft directory
`

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
}

func main() {
	action := strings.ToLower(flag.Arg(0))
	var gp GoPack
	if action == "install" && flag.NArg() > 1 {
		fmt.Println("Fetching goPack definition from", flag.Arg(1))
		err := fetchDefinition(&gp, flag.Arg(1))
		if err != nil {
			panic(err)
		}

		if installPath == nil || len(*installPath) == 0 {
			*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
		}

		gp.Install(*installPath, *minecraftPath)
	} else if action == "uninstall" || action == "update" {
		if flag.NArg() < 2 && (installPath == nil || len(*installPath) == 0) {
			panic(fmt.Errorf("Gopack URL or install location not specified!"))
		}

		if flag.NArg() > 1 {
			if strings.HasPrefix(flag.Arg(1), "http") {
				fmt.Println("Fetching goPack definition from", flag.Arg(1))
				err := fetchDefinition(&gp, flag.Arg(1))
				if err != nil {
					panic(err)
				}
			} else {
				*installPath = filepath.Join(*minecraftPath, "gopacked", flag.Arg(1))
				fmt.Println("Reading goPack definition from", *installPath)
				err := readDefinition(&gp, *installPath)
				if err != nil {
					panic(err)
				}
			}
		} else {
			fmt.Println("Reading goPack definition from", *installPath)
			err := readDefinition(&gp, *installPath)
			if err != nil {
				panic(err)
			}
		}

		if installPath == nil || len(*installPath) == 0 {
			*installPath = filepath.Join(*minecraftPath, "gopacked", gp.SimpleName)
		}

		if action == "update" {
			gp.Update(*installPath, *minecraftPath)
		} else if action == "uninstall" {
			gp.Uninstall(*installPath, *minecraftPath)
		}
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
