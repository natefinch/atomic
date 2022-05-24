// package atomic provides functions to atomically change files.
package atomic

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Copier is a type of function that copies the contents of an io.Reader to a
// an io.Writer. A simple example of this is io.Copy.
type Copier func(io.Writer, io.Reader) (int64, error)

// WrapCopier wraps a Copier function and returns an atomic writer that uses it.
// If an error occurs, the target file is guaranteed to be either fully written, or
// not written at all.  WriteFile overwrites any file that exists at the
// location (but only if the write fully succeeds, otherwise the existing file
// is unmodified).
func WrapCopier(w Copier) func(string, io.Reader) error {
	return func(filename string, r io.Reader) error {
		// write to a temp file first, then we'll atomically replace the target file
		// with the temp file.
		dir, file := filepath.Split(filename)
		if dir == "" {
			dir = "."
		}

		f, err := ioutil.TempFile(dir, file)
		if err != nil {
			return fmt.Errorf("cannot create temp file: %v", err)
		}
		defer func() {
			if err != nil {
				// Don't leave the temp file lying around on error.
				_ = os.Remove(f.Name()) // yes, ignore the error, not much we can do about it.
			}
		}()
		// ensure we always close f. Note that this does not conflict with  the
		// close below, as close is idempotent.
		defer f.Close()
		name := f.Name()
		if _, err := w(f, r); err != nil {
			return fmt.Errorf("cannot write data to tempfile %q: %v", name, err)
		}
		// fsync is important, otherwise os.Rename could rename a zero-length file
		if err := f.Sync(); err != nil {
			return fmt.Errorf("can't flush tempfile %q: %v", name, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("can't close tempfile %q: %v", name, err)
		}

		// get the file mode from the original file and use that for the replacement
		// file, too.
		destInfo, err := os.Stat(filename)
		if os.IsNotExist(err) {
			// no original file
		} else if err != nil {
			return err
		} else {
			sourceInfo, err := os.Stat(name)
			if err != nil {
				return err
			}

			if sourceInfo.Mode() != destInfo.Mode() {
				if err := os.Chmod(name, destInfo.Mode()); err != nil {
					return fmt.Errorf("can't set filemode on tempfile %q: %v", name, err)
				}
			}
		}
		if err := ReplaceFile(name, filename); err != nil {
			return fmt.Errorf("cannot replace %q with tempfile %q: %v", filename, name, err)
		}
		return nil
	}
}

// WriteFile atomically writes the contents of r to the specified filepath using io.Copy.
func WriteFile(filename string, r io.Reader) (err error) {
	return WrapCopier(io.Copy)(filename, r)
}
