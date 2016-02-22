package config

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	Current = Version{
		Major: 0,
		Minor: 15,
		Patch: 1,
	}

	First = Version{
		0, 10, 4,
	}
)

// Version prepresents the semantic version of something
type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

// VersionFromString parse a string and return the corosponding version
func VersionFromString(s string) Version {
	vs := strings.Split(s, ".")

	// if we have less than 3 elements, use the oldest known version
	if len(vs) < 3 {

		// the default version is 0.10.4
		return First
	}

	toInt := func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return i
	}

	return Version{toInt(vs[0]), toInt(vs[1]), toInt(vs[2])}
}

// String returns the string representation of a version
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Greater compares two versions and returns if v > x
func (v Version) Greater(x Version) bool {
	if v.Major > x.Major {
		return true
	}

	if v.Major == x.Major {
		if v.Minor > x.Minor {
			return true
		}

		if v.Minor == x.Minor {
			if v.Patch > x.Patch {
				return true
			}

			return false
		}
		return false
	}

	return false
}
