package jsserver

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

func (s *JSWebServer) archiveCode(code, name, executionID string) (string, error) {
	// Create filename with timestamp and execution ID
	timestamp := time.Now().Format("2006-01-02T15-04-05Z")
	
	var filename string
	if name != "" {
		// Sanitize name for filesystem
		safeName := sanitizeFilename(name)
		filename = fmt.Sprintf("%s-%s-%s.js", timestamp, executionID[len(executionID)-8:], safeName)
	} else {
		filename = fmt.Sprintf("%s-%s.js", timestamp, executionID[len(executionID)-8:])
	}

	// Ensure archive directory exists
	if err := os.MkdirAll(s.config.ArchiveDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create archive directory")
	}

	// Full path to archived file
	filepath := filepath.Join(s.config.ArchiveDir, filename)

	// Add metadata header to the code
	header := fmt.Sprintf(`/*
 * Archived JavaScript Code
 * Execution ID: %s
 * Timestamp: %s
 * Name: %s
 */

`, executionID, timestamp, name)

	fullContent := header + code

	// Write to file
	if err := os.WriteFile(filepath, []byte(fullContent), 0644); err != nil {
		return "", errors.Wrap(err, "failed to write archived file")
	}

	return filepath, nil
}

func (s *JSWebServer) listArchivedFiles() ([]map[string]interface{}, error) {
	files, err := os.ReadDir(s.config.ArchiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]interface{}{}, nil
		}
		return nil, errors.Wrap(err, "failed to read archive directory")
	}

	var archivedFiles []map[string]interface{}
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".js" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		archivedFiles = append(archivedFiles, map[string]interface{}{
			"name":     file.Name(),
			"size":     info.Size(),
			"modified": info.ModTime(),
			"path":     filepath.Join(s.config.ArchiveDir, file.Name()),
		})
	}

	return archivedFiles, nil
}

func (s *JSWebServer) getArchivedFile(filename string) (string, error) {
	// Sanitize filename to prevent directory traversal
	filename = filepath.Base(filename)
	if filepath.Ext(filename) != ".js" {
		return "", errors.New("invalid file extension")
	}

	filepath := filepath.Join(s.config.ArchiveDir, filename)
	
	content, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("archived file not found")
		}
		return "", errors.Wrap(err, "failed to read archived file")
	}

	return string(content), nil
}

func (s *JSWebServer) deleteArchivedFile(filename string) error {
	// Sanitize filename to prevent directory traversal
	filename = filepath.Base(filename)
	if filepath.Ext(filename) != ".js" {
		return errors.New("invalid file extension")
	}

	filepath := filepath.Join(s.config.ArchiveDir, filename)
	
	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return errors.New("archived file not found")
		}
		return errors.Wrap(err, "failed to delete archived file")
	}

	return nil
}

func (s *JSWebServer) cleanupArchivedFiles(keepDays int) error {
	files, err := os.ReadDir(s.config.ArchiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to cleanup
		}
		return errors.Wrap(err, "failed to read archive directory")
	}

	cutoffTime := time.Now().AddDate(0, 0, -keepDays)
	deletedCount := 0

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".js" {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			filepath := filepath.Join(s.config.ArchiveDir, file.Name())
			if err := os.Remove(filepath); err == nil {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		// Log cleanup activity
		s.storeExecution(
			"archive-cleanup-"+time.Now().Format("2006-01-02T15:04:05Z"),
			"/* automatic archive cleanup */",
			fmt.Sprintf("Cleaned up %d old archived files", deletedCount),
			true,
			"",
		)
	}

	return nil
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func sanitizeFilename(filename string) string {
	// Replace common unsafe characters
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	safe := filename
	
	for _, char := range unsafe {
		safe = filepath.Clean(safe)
		// Replace unsafe characters with underscores
		for i := 0; i < len(safe); i++ {
			if string(safe[i]) == char {
				safe = safe[:i] + "_" + safe[i+1:]
			}
		}
	}

	// Limit length to 50 characters
	if len(safe) > 50 {
		safe = safe[:50]
	}

	return safe
}