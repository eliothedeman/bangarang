package version

import "testing"

func TestVersionGreater(t *testing.T) {
	x := []struct {
		a, b     Version
		expected bool
	}{
		{
			a: Version{
				Major: 1,
			},
			b: Version{
				Major: 0,
			},
			expected: true,
		},
		{
			a: Version{
				Minor: 1,
			},
			b: Version{
				Minor: 0,
			},
			expected: true,
		},
		{
			a: Version{
				Patch: 1,
			},
			b: Version{
				Patch: 0,
			},
			expected: true,
		},
		{
			a: Version{
				Patch: 1,
			},
			b: Version{
				Major: 33,
				Patch: 0,
			},
			expected: false,
		},
	}

	for _, v := range x {
		if newer := v.a.Greater(v.b); newer != v.expected {
			t.Fatalf("Expected %t got %t a: %s, b: %s", v.expected, newer, v.a, v.b)
		}

	}
}
