package atomic

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteFileLongName(t *testing.T) {
	names := []string{"a", ".a", strings.Repeat("a", 240), "." + strings.Repeat("a", 239), strings.Repeat("a", 255), "." + strings.Repeat("a", 254)}
	for _, name := range names {
		length := len(name)
		for _, existing := range []bool{false, true} {
			t.Run(fmt.Sprintf("length=%d/hidden=%t/existing=%t", length, name[0] == '.', existing), func(t *testing.T) {
				if runtime.GOOS == "windows" && length > 2 {
					t.Skip("Windows long paths are a separate issue (#13)")
				}
				dir, err := ioutil.TempDir("", "atomic-")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(dir)
				filename := filepath.Join(dir, name)
				// Verify that this basename is supported by the filesystem.
				if err := ioutil.WriteFile(filename, []byte("original"), 0640); err != nil {
					t.Skipf("filesystem does not support a %d-byte basename: %v", length, err)
				}
				if !existing {
					if err := os.Remove(filename); err != nil {
						t.Fatal(err)
					}
				}
				if err := WriteFile(filename, strings.NewReader("replacement")); err != nil {
					t.Fatal(err)
				}
				data, err := ioutil.ReadFile(filename)
				if err != nil || string(data) != "replacement" {
					t.Fatalf("ReadFile = %q, %v; want replacement", data, err)
				}
				info, err := os.Stat(filename)
				if err != nil {
					t.Fatal(err)
				}
				wantMode := os.FileMode(0600)
				if existing {
					wantMode = 0640
				}
				if runtime.GOOS != "windows" && info.Mode().Perm() != wantMode {
					t.Fatalf("mode = %v; want %v", info.Mode().Perm(), wantMode)
				}
				files, err := ioutil.ReadDir(dir)
				if err != nil || len(files) != 1 || files[0].Name() != filepath.Base(filename) {
					t.Fatalf("unexpected directory contents: %v, %v", files, err)
				}
			})
		}
	}
}

func TestWriteFileReaderError(t *testing.T) {
	dir, err := ioutil.TempDir("", "atomic-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	filename := filepath.Join(dir, "destination")
	if err := ioutil.WriteFile(filename, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	r := io.MultiReader(strings.NewReader("partial replacement"), failingReader{})
	if err := WriteFile(filename, r); err == nil {
		t.Fatal("WriteFile succeeded with a failing reader")
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil || string(data) != "original" {
		t.Fatalf("ReadFile = %q, %v; want original", data, err)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil || len(files) != 1 || files[0].Name() != "destination" {
		t.Fatalf("unexpected directory contents: %v, %v", files, err)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("injected read failure")
}
