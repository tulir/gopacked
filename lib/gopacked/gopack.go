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

// GoPack is the base struct for a goPacked modpack.
type GoPack struct {
	Name        string                 `json:"name"`
	SimpleName  string                 `json:"simplename"`
	UpdateURL   string                 `json:"update-url"`
	Author      string                 `json:"author"`
	Version     Version                `json:"version"`
	ForgeVer    string                 `json:"forge-version,omitempty"`
	GoPackedMin Version                `json:"gopacked-version-minimum,omitempty"`
	GoPackedMax Version                `json:"gopacked-version-maximum,omitempty"`
	ProfileArgs map[string]interface{} `json:"profile-settings"`
	MCLVersion  FileEntry              `json:"mcl-version"`
	Files       FileEntry              `json:"files"`
}

type FileType string

const (
	TypeDirectory  FileType = "directory"
	TypeFile                = "file"
	TypeZipArchive          = "zip-archive"
)

type Side string

const (
	SideClient Side = "client"
	SideServer Side = "server"
)

// FileEntry contains the data of a file or directory.
type FileEntry struct {
	Type     FileType             `json:"type"`
	FileName string               `json:"filename,omitempty"`
	Version  Version              `json:"version,omitempty"`
	Side     Side                 `json:"side,omitempty"`
	URL      string               `json:"url,omitempty"`
	Children map[string]FileEntry `json:"children,omitempty"`
}
