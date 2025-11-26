package fileutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CopyDir(source, dest string) error {
	srcInfo, err := validateSourceDirectory(source)
	if err != nil {
		return err
	}

	if err := createDestinationDirectory(dest, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := readDirectoryContents(source)
	if err != nil {
		return err
	}

	if err := copyDirectoryEntries(entries, source, dest); err != nil {
		return err
	}

	return nil
}

func validateSourceDirectory(source string) (os.FileInfo, error) {
	srcInfo, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("failed to access source directory %s: %v", source, err)
	}

	if !srcInfo.IsDir() {
		return nil, fmt.Errorf("source path %s is not a directory", source)
	}

	return srcInfo, nil
}

func createDestinationDirectory(dest string, mode os.FileMode) error {
	if err := os.MkdirAll(dest, mode); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %v", dest, err)
	}
	return nil
}

func readDirectoryContents(source string) ([]os.FileInfo, error) {
	dir, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open source directory %s: %v", source, err)
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read source directory %s: %v", source, err)
	}

	return entries, nil
}

func copyDirectoryEntries(entries []os.FileInfo, source, dest string) error {
	var errors []error

	for _, entry := range entries {
		if err := copyEntry(entry, source, dest); err != nil {
			errors = append(errors, err)
		}
	}

	return combineErrors(errors)
}

func copyEntry(entry os.FileInfo, source, dest string) error {
	sourcePath := filepath.Join(source, entry.Name())
	destPath := filepath.Join(dest, entry.Name())

	if entry.IsDir() {
		if err := CopyDir(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy directory %s: %v", sourcePath, err)
		}
	} else {
		if err := CopyFile(sourcePath, destPath); err != nil {
			return fmt.Errorf("failed to copy file %s: %v", sourcePath, err)
		}
	}

	return nil
}

func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var errorMessages []string
	for _, err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return fmt.Errorf("directory copy completed with errors:\n%s", strings.Join(errorMessages, "\n"))
}
