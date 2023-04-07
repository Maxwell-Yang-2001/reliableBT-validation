package utils

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const bufSize = PieceLength * 80 // 10MB per

// Create a directory at said location.
// Fails if any file IO fails.
func CreateDir(t *testing.T, path string) {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		t.FailNow()
	}
}

// Create a file at specified location with randomized content of specifed amount,
// And duplicated it among all spoecified locations,
// Returns the first one of them.
// Fails if any file IO fails (but not when the file already exists) or no path is given.
func CreateFiles(t *testing.T, paths []string, size int64) (file *os.File) {
	require.NotZero(t, len(paths))
	// First ensure the directories are good
	for _, path := range paths {
		CreateDir(t, filepath.Dir(path))
	}

	files := make([]*os.File, len(paths))

	for i, path := range paths {
		fmt.Printf("Creating test file at path %s\n", path)
		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			t.FailNow()
		}
		defer file.Close()
		files[i] = file
	}

	// Write chunks by chunks of randomized data
	buf := make([]byte, bufSize)
	for size >= bufSize {
		_, err := rand.Read(buf)
		if err != nil {
			t.FailNow()
		}
		for _, file := range files {
			_, err = file.Write(buf)
			if err != nil {
				t.FailNow()
			}
		}
		size -= bufSize
	}

	// Write the remaining incomplete chunk
	if size != 0 {
		buf = make([]byte, size)
		_, err := rand.Read(buf)
		if err != nil {
			t.FailNow()
		}
		for _, file := range files {
			_, err = file.Write(buf)
			if err != nil {
				t.FailNow()
			}
		}
	}
	return files[0]
}

// Check whether file at each path to be checked has the same content as that at the reference path,
// Fails if any file IO fails or any file content mismatch is found.
func VerifyFileContent(t *testing.T, name string, refDir string, checkDirs []string) {
	fmt.Println("Verifying file content equality")
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
			fmt.Println("SUCCESS: File content equality holds")
			return
		}
	}
}
