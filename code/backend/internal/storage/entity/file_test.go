package entity

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewFileAcceptsUnicodeAndCommonFileNameCharacters(t *testing.T) {
	name := strings.Repeat("Документ + версия #", 12)
	if len([]rune(name)) > maxFileNameLength {
		t.Fatalf("test name unexpectedly exceeds the limit: %d", len([]rune(name)))
	}

	file, err := NewFile(
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
		"",
		name,
		"application/x-custom-format",
		0,
		"",
	)
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if file.MimeType != "application/x-custom-format" {
		t.Fatalf("MimeType = %q, want the supplied MIME type", file.MimeType)
	}
}

func TestNewFileUsesFallbackMimeType(t *testing.T) {
	file, err := NewFile(
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
		"",
		"без типа",
		"",
		0,
		"",
	)
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	if file.MimeType != "application/octet-stream" {
		t.Fatalf("MimeType = %q, want application/octet-stream", file.MimeType)
	}
}

func TestNewFileRejectsUnsafeNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "folder/file.txt", `folder\file.txt`, "line\nbreak"} {
		t.Run(name, func(t *testing.T) {
			_, err := NewFile(
				uuid.NewString(),
				uuid.NewString(),
				uuid.NewString(),
				"",
				name,
				"text/plain",
				0,
				"",
			)
			if err == nil {
				t.Fatalf("NewFile() accepted unsafe name %q", name)
			}
		})
	}
}

func TestNewFileRejectsNamesOverLimit(t *testing.T) {
	_, err := NewFile(
		uuid.NewString(),
		uuid.NewString(),
		uuid.NewString(),
		"",
		strings.Repeat("а", maxFileNameLength+1),
		"text/plain",
		0,
		"",
	)
	if err == nil {
		t.Fatal("NewFile() accepted a name over the maximum length")
	}
}
