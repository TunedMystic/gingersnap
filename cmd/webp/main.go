package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gingersnap"
)

func Convert(paths []string) error {

	for _, path := range paths {
		srcPath := path
		dstPath := strings.ReplaceAll(srcPath, filepath.Ext(path), ".webp")
		fmt.Printf("✨🏞️ ✨converting %s to webp\n", srcPath)

		cmd := []string{"./bin/cwebp", "-q", "85", srcPath, "-o", dstPath}

		// Convert the image to webp.
		if _, err := exec.Command(cmd[0], cmd[1:]...).Output(); err != nil {
			return fmt.Errorf("failed webp conversion: %w", err)
		}

		// Remove the non-webp image.
		if err := os.Remove(srcPath); err != nil {
			return fmt.Errorf("failed to remove original image: %w", err)
		}
	}

	return nil
}

// ------------------------------------------------------------------
//
//
// Main entrypoint
//
//
// ------------------------------------------------------------------

func main() {
	paths := make([]string, 0, 10)
	globs := []string{
		"**/*.png",
		"**/*.jpg",
		"**/*.jpeg",
	}

	for _, pattern := range globs {
		names, err := gingersnap.Glob(os.DirFS("."), "assets/media", pattern)
		if err != nil {
			log.Fatal(err)
		}
		paths = append(paths, names...)
	}

	if err := Convert(paths); err != nil {
		log.Fatal(err)
	}
}
