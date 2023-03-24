package utils

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Check whether file at each path to be checked has the same content as that at the reference path
// Fails if any file IO fails or any file content mismatch is found
// TODO: consider that read does not necessarily load up the whole buffer
func VerifyFileContent(t *testing.T, refPath string, checkPaths []string) {
	const bufSize = 64000
	// Open the reference file and to-be-verified ones
	refFile, err := os.Open(refPath)
	if err != nil {
		t.FailNow()
	}
	defer refFile.Close()

	checkFiles := make([]*os.File, len(checkPaths))
	for i, path := range checkPaths {
		checkFiles[i], err = os.Open(path)
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
			if err != io.EOF {
				t.FailNow()
			}
			endOfFile = true
		}

		checkBuf := make([]byte, bufSize)
		for _, checkFile := range checkFiles {
			checkBytesRead, err := io.ReadFull(checkFile, checkBuf)

			// Error acceptable only if reaching EOF when reference also does so
			if err != nil && (err != io.EOF || !endOfFile) {
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
