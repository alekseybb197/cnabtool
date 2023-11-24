package client

import (
	"testing"
)

func TestParseReference(t *testing.T) {

	testlines := []string{
		"registry.example.com/repository/image:v1.2.3@sha256:456",
		"registry.example.com/repository/image@sha256:456",
		"registry.example.com/repository/image:v1.2.3",
	}
	expectedlines := [][]string{
		{"registry.example.com", "repository/image", "v1.2.3", "sha256:456"},
		{"registry.example.com", "repository/image", "", "sha256:456"},
		{"registry.example.com", "repository/image", "v1.2.3", ""},
	}

	for i, tl := range testlines {
		cl := RegClient{
			Reference: tl,
		}

		if err := ParseReference(&cl); err != nil {
			t.Errorf("Should not produce an error for %s", tl)
		}

		actual := cl.Registry
		expected := expectedlines[i][0]
		if actual != expected {
			t.Errorf("Registry was incorrect, got: %s, want: %s.", actual, expected)
		}
		actual = cl.Repository
		expected = expectedlines[i][1]
		if actual != expected {
			t.Errorf("Repository was incorrect, got: %s, want: %s.", actual, expected)
		}
		actual = cl.Tag
		expected = expectedlines[i][2]
		if actual != expected {
			t.Errorf("Tag was incorrect, got: %s, want: %s.", actual, expected)
		}
		actual = cl.Digest
		expected = expectedlines[i][3]
		if actual != expected {
			t.Errorf("Digest was incorrect, got: %s, want: %s.", actual, expected)
		}

	}
}
