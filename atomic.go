// package atomic provides functions to atomically change files.
package atomic

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// WriteFile atomically writes the contents of r to the specified filepath.  If
// an error occurs, the target file is guaranteed to be either fully written, or
// not written at all.  WriteFile overwrites any file that exists at the
// location (but only if the write fully succeeds, otherwise the existing file
// is unmodified).
func WriteFile(filename string, r io.Reader) (res *AtomicResult) {
	// write to a temp file first, then we'll atomically replace the target file
	// with the temp file.
	dir, file := filepath.Split(filename)
	if dir == "" {
		dir = "."
	}

	f, err := ioutil.TempFile(dir, file)
	if err != nil {
		return NewAtomicError(fmt.Sprintf("cannot create temp file: %v", err), "")
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
	if _, err := io.Copy(f, r); err != nil {
		return NewAtomicError(fmt.Sprintf("cannot write data to temp file %q: %v", name, err), name)
	}
	// fsync is important, otherwise os.Rename could rename a zero-length file
	if err := f.Sync(); err != nil {
		return NewAtomicError(fmt.Sprintf("cannot flush temp file %q: %v", name, err), name)
	}
	if err := f.Close(); err != nil {
		return NewAtomicError(fmt.Sprintf("cannot close temp file %q: %v", name, err), name)
	}

	// get the file mode from the original file and use that for the replacement
	// file, too.
	destInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// no original file
	} else if err != nil {
		return NewAtomicError(fmt.Sprintf("cannot get permissions info from original file %q: %v", filename, err), name)
	} else {
		sourceInfo, err := os.Stat(name)
		if err != nil {
			return NewAtomicError(fmt.Sprintf("cannot get permissions info from temp file %q: %v", name, err), name)
		}

		if sourceInfo.Mode() != destInfo.Mode() {
			if err := os.Chmod(name, destInfo.Mode()); err != nil {
				return NewAtomicError(fmt.Sprintf("cannot set filemode of temp file %q: %v", name, err), name)
			}
		}
	}
	if err := ReplaceFile(name, filename); err != nil {
		return NewAtomicError(fmt.Sprintf("cannot replace file %q with temp file %q: %v", filename, name, err), name)
	}

	return NewAtomicResult(name)
}
