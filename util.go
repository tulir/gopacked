package main

import (
	"io"
	"net/http"
	"os"
)

// Download downloads the file from the given URL and saves it into the given path.
func Download(url, saveTo string) error {
	out, err := os.Create(saveTo)
	defer out.Close()
	if err != nil {
		return err
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
