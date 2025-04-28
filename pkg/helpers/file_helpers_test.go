package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindStartPosForLastNLines(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "file_helpers_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	tests := []struct {
		name     string
		content  string
		n        int
		wantPos  int64
		wantErr  bool
		fileFunc func(dir string) string // Function to create test file and return its path
	}{
		{
			name:    "empty file",
			content: "",
			n:       5,
			wantPos: 0,
		},
		{
			name:    "single line no newline",
			content: "single line",
			n:       1,
			wantPos: 0,
		},
		{
			name:    "single line with newline",
			content: "single line\n",
			n:       1,
			wantPos: 0,
		},
		{
			name:    "two lines last line",
			content: "first line\nsecond line\n",
			n:       1,
			wantPos: 11, // Position after "first line\n"
		},
		{
			name:    "three lines last two",
			content: "first line\nsecond line\nthird line\n",
			n:       2,
			wantPos: 11, // Position after "first line\n"
		},
		{
			name:    "three lines all",
			content: "first line\nsecond line\nthird line\n",
			n:       3,
			wantPos: 0,
		},
		{
			name:    "three lines more than exist",
			content: "first line\nsecond line\nthird line\n",
			n:       5,
			wantPos: 0,
		},
		{
			name:    "no newline at end",
			content: "first line\nsecond line\nthird line",
			n:       2,
			wantPos: 11, // Position after "first line\n"
		},
		{
			name:    "empty lines",
			content: "\n\n\n",
			n:       2,
			wantPos: 1, // Position after first "\n"
		},
		{
			name:    "zero lines requested",
			content: "first line\nsecond line\n",
			n:       0,
			wantPos: 0,
		},
		{
			name:    "negative lines requested",
			content: "first line\nsecond line\n",
			n:       -1,
			wantPos: 0,
		},
		{
			name:    "mixed line lengths",
			content: "short\nmedium line\nreally long line here\n",
			n:       2,
			wantPos: 6, // Position after "short\nmedium line\n"
		},
		{
			// Test for a large file that's bigger than our buffer size
			name: "large file",
			fileFunc: func(dir string) string {
				path := filepath.Join(dir, "large.txt")
				f, err := os.Create(path)
				if err != nil {
					t.Fatalf("Failed to create large test file: %v", err)
				}
				defer func() {
					if err := f.Close(); err != nil {
						t.Errorf("Failed to close test file: %v", err)
					}
				}()

				offset := 0
				// Write 100 numbered lines
				for i := 0; i < 100; i++ {
					s := "Line number /" + string(rune('0'+i%10)) + "\n"
					_, _ = f.WriteString(s)
					offset += len(s)
				}

				for i := 0; i < 10; i++ {
					s := "Line number /" + string(rune('0'+i%10)) + "\n"
					_, _ = f.WriteString(s)
				}

				return path
			},
			n:       10,
			wantPos: 1500, // Position for last 10 lines
		},
		{
			name: "nonexistent file",
			fileFunc: func(dir string) string {
				return filepath.Join(dir, "nonexistent.txt")
			},
			n:       5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			var filepath string
			if tt.fileFunc != nil {
				filepath = tt.fileFunc(tmpDir)
			} else {
				f, err := os.CreateTemp(tmpDir, "test")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				defer func() {
					if err := f.Close(); err != nil {
						t.Errorf("Failed to close temp file: %v", err)
					}
				}()
				filepath = f.Name()
				if _, err := f.WriteString(tt.content); err != nil {
					t.Fatalf("Failed to write test content: %v", err)
				}
			}

			// Run the test
			gotPos, err := FindStartPosForLastNLines(filepath, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStartPosForLastNLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && gotPos != tt.wantPos {
				t.Errorf("FindStartPosForLastNLines() = %v, want %v", gotPos, tt.wantPos)

				// For debugging: print the actual content from the found position
				if file, err := os.Open(filepath); err == nil {
					defer func() {
						if err := file.Close(); err != nil {
							t.Errorf("Failed to close file: %v", err)
						}
					}()
					_, _ = file.Seek(gotPos, 0)
					content := make([]byte, 1024)
					n, _ := file.Read(content)
					t.Logf("Content from position %d: %q", gotPos, content[:n])
				}
			}
		})
	}
}

// TestFindStartPosForLastNLinesPermissions tests permission-related errors
func TestFindStartPosForLastNLinesPermissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir, err := os.MkdirTemp("", "file_helpers_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a file with no read permissions
	noReadFile := filepath.Join(tmpDir, "no_read.txt")
	if err := os.WriteFile(noReadFile, []byte("test\n"), 0200); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = FindStartPosForLastNLines(noReadFile, 1)
	if err == nil {
		t.Error("Expected error for file with no read permissions, got nil")
	}
}
