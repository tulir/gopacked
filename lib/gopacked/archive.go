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

package gopacked

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func untargz(from io.Reader, target string) error {
	var buf bytes.Buffer
	err := ungz(from, &buf)
	if err != nil {
		return err
	}
	return untar(&buf, target)
}

func untarbz2(from io.Reader, target string) error {
	var buf bytes.Buffer
	err := unbz2(from, &buf)
	if err != nil {
		return err
	}
	return untar(&buf, target)
}

func ungz(from io.Reader, to io.Writer) error {
	reader, err := gzip.NewReader(from)
	if err != nil {
		return err
	}
	_, err = io.Copy(to, reader)
	return err
}

func unbz2(from io.Reader, to io.Writer) error {
	reader := bzip2.NewReader(from)
	_, err := io.Copy(to, reader)
	return err
}

func untar(from io.Reader, target string) error {
	tarReader := tar.NewReader(from)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		err = unarchiveTar(header, target, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func unarchiveTar(header *tar.Header, target string, reader io.Reader) error {
	path := filepath.Join(target, header.Name)
	info := header.FileInfo()
	if info.IsDir() {
		if err := os.MkdirAll(path, info.Mode()); err != nil {
			return err
		}
		return nil
	}

	return unarchive(path, info, reader)
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		err = unarchiveZip(file, target)
		if err != nil {
			return err
		}
	}

	return nil
}

func unarchiveZip(file *zip.File, target string) error {
	path := filepath.Join(target, file.Name)
	if file.FileInfo().IsDir() {
		_ = os.MkdirAll(path, file.Mode())
		return nil
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()

	return unarchive(path, file.FileInfo(), fileReader)
}

func unarchive(toPath string, fileInfo os.FileInfo, reader io.Reader) error {
	file, err := os.OpenFile(toPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileInfo.Mode())
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}
	return nil
}
