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
	"fmt"
	"strconv"
	"strings"
)

// Version contains four version number levels. Level 1 is the highest in importance.
// Higher importance means overriding of lower importances.
type Version struct {
	Level1, Level2, Level3, Level4 int
}

// Compare compares this version to the given one.
// Return value 1 means that this version is greater than the given one.
// Return value 0 means that the versions are equal.
// Return value -1 means that this version is smaller than the given one.
func (ver Version) Compare(ver2 Version) int {
	if ver.Level1 > ver2.Level1 {
		return 1
	} else if ver.Level1 < ver2.Level1 {
		return -1
	}

	if ver.Level2 > ver2.Level2 {
		return 1
	} else if ver.Level2 < ver2.Level2 {
		return -1
	}

	if ver.Level3 > ver2.Level3 {
		return 1
	} else if ver.Level3 < ver2.Level3 {
		return -1
	}

	if ver.Level4 > ver2.Level4 {
		return 1
	} else if ver.Level4 < ver2.Level4 {
		return -1
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
	return fmt.Sprintf("%d.%d.%d.%d", ver.Level1, ver.Level2, ver.Level3, ver.Level4)
}

// ParseAndCompare parses a version out of the two strings and compares them.
func ParseAndCompare(str1, str2 string) (int, error) {
	ver1, err := ParseVersion(str1)
	if err != nil {
		return -2, err
	}
	ver2, err := ParseVersion(str2)
	if err != nil {
		return -2, err
	}
	return ver1.Compare(ver2), nil
}

// ParseVersion parses a version from a string.
func ParseVersion(str string) (Version, error) {
	pieces := strings.Split(str, ".")

	var ver Version
	var err error

	if len(pieces) != 4 {
		return ver, fmt.Errorf("The amount of levels (%d) is incorrect", len(pieces))
	}

	ver.Level1, err = strconv.Atoi(pieces[0])
	if err != nil {
		return ver, fmt.Errorf("The first level is not an integer")
	}

	ver.Level2, err = strconv.Atoi(pieces[1])
	if err != nil {
		return ver, fmt.Errorf("The second level is not an integer")
	}

	ver.Level3, err = strconv.Atoi(pieces[2])
	if err != nil {
		return ver, fmt.Errorf("The third level is not an integer")
	}

	ver.Level4, err = strconv.Atoi(pieces[3])
	if err != nil {
		return ver, fmt.Errorf("The fourth level is not an integer")
	}

	return ver, nil
}
