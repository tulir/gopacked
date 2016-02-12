package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Version TODO comment
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
