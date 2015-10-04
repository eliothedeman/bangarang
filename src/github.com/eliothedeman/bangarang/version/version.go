package version

import "fmt"

var Current = &Version{
	Major: 0,
	Minor: 10,
	Patch: 4,
}

// Version prepresents the semantic version of something
type Version struct {
	Major,
	Minor,
	Patch int
}

// String returns the string representation of a version
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Newer compares two versions and returns if v > x
func (v *Version) Newer(x *Version) bool {
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
