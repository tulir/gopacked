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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Version contains four version number levels. Level 1 is the highest in importance.
// Higher importance means overriding of lower importances.
type Version []int

// Compare compares this version to the given one.
// Return value 1 means that this version is greater than the given one.
// Return value 0 means that the versions are equal.
// Return value -1 means that this version is smaller than the given one.
func (ver Version) Compare(ver2 Version) int {
	for i := 0; i < len(ver) || i < len(ver2); i++ {
		var val1, val2 int
		if len(ver) < i {
			val1 = ver[i]
		}
		if len(ver2) < i {
			val2 = ver2[i]
		}
		if val1 < val2 {
			return -1
		} else if val1 > val2 {
			return 1
		}
	}
	return 0
}

// IsGreater checks if this version is greater than the given one.
func (ver Version) IsGreater(ver2 Version) bool {
	return ver.Compare(ver2) == 1
}

// IsEqual checks if this version is equal to the given one.
func (ver Version) IsEqual(ver2 Version) bool {
	return ver.Compare(ver2) == 0
}

// IsSmaller checks if this version is smaller than the given one.
func (ver Version) IsSmaller(ver2 Version) bool {
	return ver.Compare(ver2) == -1
}

func (ver Version) String() string {
	var buf strings.Builder
	for i, val := range ver {
		buf.WriteString(strconv.Itoa(val))
		if i < len(ver)-1 {
			buf.WriteRune('.')
		}
	}
	return buf.String()
}

func (ver *Version) UnmarshalJSON(blob []byte) error {
	var data string
	err := json.Unmarshal(blob, &data)
	if err != nil {
		return err
	}
	*ver, err = ParseVersion(data)
	return err
}

func (ver Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", ver.String())), nil
}

// ParseVersion parses a version from a string.
func ParseVersion(str string) (Version, error) {
	pieces := strings.Split(str, ".")

	ver := make(Version, len(pieces))
	var err error
	for i, piece := range pieces {
		ver[i], err = strconv.Atoi(piece)
		if err != nil {
			return ver, err
		}
	}
	return ver, nil
}
