package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathToSlug(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard absolute path with hyphens in folder",
			input:    "/Users/nemesis/Projects/my-projects/aim",
			expected: "-Users-nemesis-Projects-my-projects-aim",
		},
		{
			name:     "private tmp path",
			input:    "/private/tmp",
			expected: "-private-tmp",
		},
		{
			name:     "standard absolute path without hyphens",
			input:    "/Users/nemesis/Projects/aim",
			expected: "-Users-nemesis-Projects-aim",
		},
		{
			name:     "trailing slash",
			input:    "/Users/nemesis/Projects/aim/",
			expected: "-Users-nemesis-Projects-aim",
		},
		{
			name:     "multiple trailing slashes",
			input:    "/Users/nemesis/Projects/aim///",
			expected: "-Users-nemesis-Projects-aim",
		},
		{
			name:     "no leading slash",
			input:    "Users/nemesis/Projects/aim",
			expected: "-Users-nemesis-Projects-aim",
		},
		{
			name:     "no leading slash with trailing slash",
			input:    "Users/nemesis/Projects/aim/",
			expected: "-Users-nemesis-Projects-aim",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "-",
		},
		{
			name:     "multiple root slashes",
			input:    "///",
			expected: "-",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := PathToSlug(c.input)
			if got != c.expected {
				t.Errorf("PathToSlug(%q) = %q, expected %q", c.input, got, c.expected)
			}
		})
	}
}

func TestSlugToPath(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard slug with hyphens in folder name",
			input:    "-Users-nemesis-Projects-my-projects-aim",
			expected: "/Users/nemesis/Projects/my-projects/aim",
		},
		{
			name:     "private tmp slug",
			input:    "-private-tmp",
			expected: "/private/tmp",
		},
		{
			name:     "standard slug without folder hyphens",
			input:    "-Users-nemesis-Projects-aim",
			expected: "/Users/nemesis/Projects/aim",
		},
		{
			name:     "slug without leading dash",
			input:    "Users-nemesis-Projects-aim",
			expected: "/Users/nemesis/Projects/aim",
		},
		{
			name:     "slug with trailing dash",
			input:    "-Users-nemesis-Projects-aim-",
			expected: "/Users/nemesis/Projects/aim",
		},
		{
			name:     "slug without leading dash and with trailing dash",
			input:    "Users-nemesis-Projects-aim-",
			expected: "/Users/nemesis/Projects/aim",
		},
		{
			name:     "root slug",
			input:    "-",
			expected: "/",
		},
		{
			name:     "multiple dashes for root",
			input:    "---",
			expected: "/",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := SlugToPath(c.input)
			if got != c.expected {
				t.Errorf("SlugToPath(%q) = %q, expected %q", c.input, got, c.expected)
			}
		})
	}
}

func TestSlugToPath_FilesystemResolution(t *testing.T) {
	tempBase := t.TempDir()
	nestedWithHyphens := filepath.Join(tempBase, "hyphenated-project", "deeply-nested-app")
	if err := os.MkdirAll(nestedWithHyphens, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	slug := PathToSlug(nestedWithHyphens)
	got := SlugToPath(slug)
	if got != nestedWithHyphens {
		t.Errorf("SlugToPath(%q) = %q, expected %q", slug, got, nestedWithHyphens)
	}
}

func TestPathToSlug_SlugToPath_Roundtrip(t *testing.T) {
	paths := []string{
		"/private/tmp",
		"/Users/nemesis/Projects/aim",
		"/usr/local/bin",
	}

	for _, p := range paths {
		slug := PathToSlug(p)
		recovered := SlugToPath(slug)
		if recovered != p {
			t.Errorf("Roundtrip failed for %q: slug=%q, recovered=%q", p, slug, recovered)
		}
	}
}
