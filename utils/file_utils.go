package utils

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

const bufSize = PieceLength * 80 // 10MB per chunk

// Create a file at said location with randomized content of specifed amount,
// Returns the entire file without closing it.
// Fails if any file IO fails (but not when the file already exists).
func CreateFile(t *testing.T, path string, size int64) (file *os.File) {
	// First ensure the directory is good
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		t.FailNow()
	}

	file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		t.FailNow()
	}
	defer file.Close()

	// Write chunks by chunks of randomized data
	buf := make([]byte, bufSize)
	for size >= bufSize {
		_, err = rand.Read(buf)
		if err != nil {
			t.FailNow()
		}
		_, err = file.Write(buf)
		if err != nil {
			t.FailNow()
		}
		size -= bufSize
	}

	// Write the remaining incomplete chunk
	if size != 0 {
		buf = make([]byte, size)
		_, err = rand.Read(buf)
		if err != nil {
			fmt.Println("h5")
			t.FailNow()
		}
		_, err = file.Write(buf)
		if err != nil {
			fmt.Println("h6")
			t.FailNow()
		}
	}
	return
}

// Check whether file at each path to be checked has the same content as that at the reference path,
// Fails if any file IO fails or any file content mismatch is found.
func VerifyFileContent(t *testing.T, name string, refDir string, checkDirs []string) {
	// Open the reference file and to-be-verified ones
	refFile, err := os.Open(filepath.Join(refDir, name))
	if err != nil {
		t.FailNow()
	}
	defer refFile.Close()

	checkFiles := make([]*os.File, len(checkDirs))
	for i, checkDir := range checkDirs {
		checkFiles[i], err = os.Open(filepath.Join(checkDir, name))
		if err != nil {
			t.FailNow()
		}
		defer checkFiles[i].Close()
	}

	for {
		refBuf := make([]byte, bufSize)
		refBytesRead, err := io.ReadFull(refFile, refBuf)

		endOfFile := false

		// Error acceptable only if reaching EOF
		if err != nil {
			if err != io.ErrUnexpectedEOF {
				t.FailNow()
			}
			endOfFile = true
		}

		checkBuf := make([]byte, bufSize)
		for _, checkFile := range checkFiles {
			checkBytesRead, err := io.ReadFull(checkFile, checkBuf)

			// Error acceptable only if reaching EOF when reference also does so
			if err != nil && (err != io.ErrUnexpectedEOF || !endOfFile) {
				t.FailNow()
			}

			if refBytesRead != checkBytesRead || !bytes.Equal(refBuf, checkBuf) {
				t.FailNow()
			}
		}

		if endOfFile {
			return
		}
	}
}
