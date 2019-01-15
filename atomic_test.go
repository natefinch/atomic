package atomic

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestWriteFile(t *testing.T) {
	f, err := ioutil.TempFile("", "atomic")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	contents := bytes.Repeat([]byte("abcde"), 100)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := WriteFile(f.Name(), bytes.NewReader(contents))
			if err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()

	written, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != string(contents) {
		t.Errorf("file contents should be %q but got %q", string(contents), string(written))
	}
}
