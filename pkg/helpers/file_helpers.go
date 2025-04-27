package helpers

import (
	"io"
	"os"
)

// FindStartPosForLastNLines finds the position in a file where the last N lines begin.
// It returns the byte offset from the start of the file and any error encountered.
// If n <= 0, it returns 0 (start of file).
// If the file has fewer lines than n, it returns 0 (start of file).
func FindStartPosForLastNLines(filename string, n int) (int64, error) {
	if n <= 0 {
		return 0, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			}
		}
	}()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}
	fileSize := stat.Size()
	if fileSize == 0 {
		return 0, nil
	}

	// Start from end
	pos := fileSize
	// Use a 4KB buffer size
	buf := make([]byte, 4096)
	linesFound := 0

	// Handle the case where the file doesn't end with a newline
	// by treating the end of file as an implicit line ending
	lastByteRead := false
	lastByteWasNewline := false

	for pos > 0 && linesFound <= n {
		// Calculate how much to read
		bytesToRead := int64(len(buf))
		if pos < bytesToRead {
			bytesToRead = pos
		}

		// Read a chunk from the right position
		pos -= bytesToRead
		_, err := file.Seek(pos, io.SeekStart)
		if err != nil {
			return 0, err
		}

		// Read the chunk
		bytesRead, err := file.Read(buf[:bytesToRead])
		if err != nil {
			return 0, err
		}

		// Count newlines in this chunk, going backwards
		for i := bytesRead - 1; i >= 0; i-- {
			if !lastByteRead {
				lastByteRead = true
				lastByteWasNewline = buf[i] == '\n'
				continue
			}

			_ = lastByteWasNewline

			if buf[i] == '\n' {
				linesFound++
				if linesFound >= n {
					// Add 1 to skip the newline itself
					return pos + int64(i) + 1, nil
				}
			}
		}
	}

	// If we get here, we need to read from the start of the file
	return 0, nil
}
