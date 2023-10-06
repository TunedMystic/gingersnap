package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ------------------------------------------------------------------
//
//
// Utility functions
//
//
// ------------------------------------------------------------------

// Slugify builds a slug from the given string.
// .
func Slugify(s string) string {
	return strings.TrimSpace(strings.ToLower(strings.ReplaceAll(s, " ", "-")))
}

// SafeDir returns a filepath directory.
// If the given path is a file, then the parent directory of the file will be returned.
// If the given path is a directory, then the directory itself will be returned.
// .
func SafeDir(p string) string {
	if ext := filepath.Ext(p); ext != "" {
		return filepath.Dir(p)
	}
	return p
}

// Exists checks if a file/directory exists.
// .
func Exists(source string) bool {
	_, err := os.Stat(source)

	if os.IsNotExist(err) {
		return false
	}
	return true
}

// OpenFile opens and returns a file from a filesystem.
// .
func OpenFile(fsys fs.FS, source string) (fs.File, error) {
	// Open source file.
	f, err := fsys.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}

// ReadFile reads a file and returns the byte slice.
// .
func ReadFile(source string) ([]byte, error) {
	// Read the source file.
	srcBytes, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return srcBytes, nil
}

// EnsurePath creates the directory paths if they do not exist.
// .
func EnsurePath(source string) error {
	// Create parent directories, if necessary.
	if err := os.MkdirAll(filepath.Dir(source), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}
	return nil
}

// CreateFile creates and returns a file.
// .
func CreateFile(source string) (*os.File, error) {
	if err := EnsurePath(source); err != nil {
		return nil, err
	}

	// Create destination file.
	f, err := os.Create(source)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return f, nil
}

// WriteFile writes data to the named file.
// .
func WriteFile(source string, data []byte) error {
	if err := EnsurePath(source); err != nil {
		return err
	}

	// Write the contents to the file.
	if err := os.WriteFile(source, data, os.FileMode(0644)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CopyDir recursively copies a directory to a destination.
// .
func CopyDir(fsys fs.FS, root, dst string) error {
	prefix, _ := filepath.Split(root)

	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {

		if !d.IsDir() {
			strippedP, ok := strings.CutPrefix(p, prefix)
			if !ok {
				return fmt.Errorf("failed to cut prefix %s from %s", prefix, p)
			}

			if err := CopyFile(fsys, p, filepath.Join(dst, strippedP)); err != nil {
				return err
			}
		}
		return nil
	})
}

// CopyFile copies a file to a destination.
// .
func CopyFile(fsys fs.FS, src, dst string) error {
	// Open the source file.
	srcFile, err := OpenFile(fsys, src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file.
	destFile, err := CreateFile(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the source file to the destination.
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Glob returns a list of files matching the pattern.
// The pattern can include **/ to match any number of directories.
//
// Ref: https://github.com/guillermo/doubleglob/blob/main/double_glob.go
// .
func Glob(inputFS fs.FS, root, glob string) ([]string, error) {
	files := make([]string, 0, 10)

	// Convert the given glob into a regex.
	// ex:  *.png  -->  [^/]*\.png
	//
	replPattern := regexp.MustCompile(`(\.)|(\*\*\/)|(\*)|([^\/\*]+)|(\/)`)
	globPattern := replPattern.ReplaceAllStringFunc(glob, func(s string) string {
		switch s {
		case "/":
			return "\\/"
		case ".":
			return "\\."
		case "**/":
			return ".*"
		case "*":
			return "[^/]*"
		default:
			return s
		}
	})

	pathPattern := regexp.MustCompile("^" + globPattern + "$")

	// Walk the directory, and match each file against the regex.
	// If it is a regex match, then collect the path.
	err := fs.WalkDir(inputFS, root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || err != nil {
			return nil
		}

		if pathPattern.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})

	// Return all collected paths.
	return files, err
}

// LocalGlob returns a list of files matching the given extensions,
// starting from the root directory.
// .
func LocalGlob(root string, ext ...string) ([]string, error) {
	ret := make([]string, 0, 10)

	for _, ext := range ext {
		paths, err := Glob(os.DirFS("."), root, "**/*."+ext)
		if err != nil {
			return nil, err
		}
		ret = append(ret, paths...)
	}

	return ret, nil
}
