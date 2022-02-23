// package atomic provides functions to atomically change files.
package atomic

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileOptions define the behaviour of `FileWrite()`.
type FileOptions struct {
	defaultFileMode os.FileMode
	fileMode        os.FileMode
	keepFileMode    bool
}

// Option functions modify FileOptions.
type Option func(*FileOptions)

// FileMode sets the file mode to the desired value and has precedence over all
// other options.
func FileMode(mode os.FileMode) Option {
	return func(opts *FileOptions) {
		opts.fileMode = mode
	}
}

// DefaultFileMode sets the default file mode instead of using the
// `ioutil.TempFile()` default of `0600`.
func DefaultFileMode(mode os.FileMode) Option {
	return func(opts *FileOptions) {
		opts.defaultFileMode = mode
	}
}

// KeepFileMode preserves the file mode of an existing file instead of using the
// default value.
func KeepFileMode(keep bool) Option {
	return func(opts *FileOptions) {
		opts.keepFileMode = keep
	}
}

// WriteFile atomically writes the contents of r to the specified filepath.  If
// an error occurs, the target file is guaranteed to be either fully written, or
// not written at all.  WriteFile overwrites any file that exists at the
// location (but only if the write fully succeeds, otherwise the existing file
// is unmodified).  Additional option arguments can be used to change the
// default configuration for the target file.
func WriteFile(filename string, r io.Reader, opts ...Option) (err error) {
	// original behaviour is to preserve the mode of an existing file.
	fopts := &FileOptions{
		keepFileMode: true,
	}
	for _, opt := range opts {
		opt(fopts)
	}

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
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("cannot write data to tempfile %q: %v", name, err)
	}
	// fsync is important, otherwise os.Rename could rename a zero-length file
	if err := f.Sync(); err != nil {
		return fmt.Errorf("can't flush tempfile %q: %v", name, err)
	}
	// get file info via file descriptor before closing it.
	sourceInfo, err := f.Stat()
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("can't close tempfile %q: %v", name, err)
	}

	var fileMode os.FileMode
	// change default file mode for when file does not exist yet.
	if fopts.defaultFileMode != 0 {
		fileMode = fopts.defaultFileMode
	}
	// get the file mode from the original file and use that for the replacement
	// file, too.
	if fopts.keepFileMode {
		destInfo, err := os.Stat(filename)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if destInfo != nil {
			fileMode = destInfo.Mode()
		}
	}
	// given file mode always takes precedence
	if fopts.fileMode != 0 {
		fileMode = fopts.fileMode
	}
	// apply possible file mode change
	if fileMode != 0 && fileMode != sourceInfo.Mode() {
		if err := os.Chmod(name, fileMode); err != nil {
			return fmt.Errorf("can't set filemode on tempfile %q: %v", name, err)
		}
	}
	if err := ReplaceFile(name, filename); err != nil {
		return fmt.Errorf("cannot replace %q with tempfile %q: %v", filename, name, err)
	}
	return nil
}
