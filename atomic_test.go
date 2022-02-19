package atomic

import (
	"bytes"
	"os"
	"testing"
)

func TestWriteFile(t *testing.T) {
	file := "foo.txt"
	content := bytes.NewBufferString("foo")
	defer func() { _ = os.Remove(file) }()
	if err := WriteFile(file, content); err != nil {
		t.Errorf("Failed to write file: %q: %v", file, err)
	}
	fi, err := os.Stat(file)
	if err != nil {
		t.Errorf("Failed to stat file: %q: %v", file, err)
	}
	if fi.Mode() != 0600 {
		t.Errorf("File mode not correct")
	}
}

func TestWriteDefaultFileMode(t *testing.T) {
	file := "bar.txt"
	content := bytes.NewBufferString("bar")
	defer func() { _ = os.Remove(file) }()
	if err := WriteFile(file, content, DefaultFileMode(0644)); err != nil {
		t.Errorf("Failed to write file: %q: %v", file, err)
	}
	fi, err := os.Stat(file)
	if err != nil {
		t.Errorf("Failed to stat file: %q: %v", file, err)
	}
	if fi.Mode() != 0644 {
		t.Errorf("File mode not correct")
	}
}
