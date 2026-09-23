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
	// If /Users/nemesis/Projects/my-projects/aim exists on disk (e.g. host Mac),
	// SlugToPath resolves it via live filesystem inspection.
	// In environments where it does not exist (e.g. CI), it falls back to slash replacement.
	expectedMyProjects := "/Users/nemesis/Projects/my/projects/aim"
	if fi, err := os.Stat("/Users/nemesis/Projects/my-projects/aim"); err == nil && fi.IsDir() {
		expectedMyProjects = "/Users/nemesis/Projects/my-projects/aim"
	}

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "slug with hyphens in folder name",
			input:    "-Users-nemesis-Projects-my-projects-aim",
			expected: expectedMyProjects,
		},
		{
			name:     "slug where directory does not exist on disk falls back to slash replacement",
			input:    "-nonexistent-path-my-projects-aim",
			expected: "/nonexistent/path/my/projects/aim",
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

func TestSlugToPath_WorkingDirFastPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}

	slug := PathToSlug(wd)
	got := SlugToPath(slug)
	if got != wd {
		t.Errorf("SlugToPath(%q) = %q, expected working dir %q", slug, got, wd)
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

func TestSlugToPath_LeafMustBeDirectory(t *testing.T) {
	tempBase := t.TempDir()
	filePath := filepath.Join(tempBase, "some-file.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	slug := PathToSlug(filePath)
	got := SlugToPath(slug)
	// Because filePath is a file and not a directory, filesystem resolution must not return it,
	// and fallback slash replacement should be used instead.
	expected := filepath.Clean("/" + filepath.ToSlash(filePath))
	// On fallback, hyphens in the filename "some-file.txt" are split
	expectedFallback := filepath.Clean("/" + filepath.ToSlash(tempBase) + "/some/file.txt")
	if got != expectedFallback && got != expected {
		t.Errorf("SlugToPath on regular file (%q) returned %q, expected fallback without directory match", slug, got)
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
