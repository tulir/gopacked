package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// HTTPGet performs a HTTP GET request on the given URL
func HTTPGet(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		return []byte{}
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}
	}
	response.Body.Close()
	return contents
}

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
