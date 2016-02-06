package main

import (
	"encoding/json"
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
	"runtime"
	"strings"
)

var installPath = flag.StringP("path", "p", "", "")
var minecraftPath = flag.StringP("minecraft", "m", "", "")

var help = `Simple command-line modpack manager.

Usage:
  gopacked [-h] [-p PATH] [-m PATH] <ACTION> <URL>

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
		mp.Install(*installPath, *minecraftPath)
	}
}
