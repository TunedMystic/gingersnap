package app

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gingersnap/app/utils"
)

//go:embed "assets/bin/cwebp"
var webpBinary []byte

// ------------------------------------------------------------------
//
//
// Type: webp
//
//
// ------------------------------------------------------------------

// webp is responsible for converting images to the `.webp` format.
//
// Main method: `Convert()`
//
// webp is a thin wrapper which manages storage and execution of
// the `cwebp` binary. It is constructed with a given root dir,
// which is typically set to the user cache dir.
// .
type webp struct {
	// The location of the cwebp binary.
	path string

	// The function to execute the webp conversion.
	execFunc func(string) error
}

// NewWebp returns a new *webp object with a root directory.
// The root directory is where the `cwebp` binary will be stored.
// Typically, the root directory will be set to the user cache dir:
// .
func NewWebp() *webp {

	var path string

	// The user cache dir, used for user-specific cached data.
	dir, err := os.UserCacheDir()

	if err != nil {
		path = filepath.Join("gingersnap", "cwebp")
	} else {
		path = filepath.Join(dir, "gingersnap", "cwebp")
	}

	w := webp{}
	w.path = path
	w.execFunc = w.convertFunc
	return &w
}

// Convert takes the src image and converts it to a
// webp image in the same location.
// .
func (w webp) Convert(srcs ...string) error {

	// If the `cwebp` binary does not exist,
	// then unpack it from the embedded assets.
	if !utils.Exists(w.path) {
		if err := w.unpack(); err != nil {
			return fmt.Errorf("convert: %w", err)
		}
	}

	// Execute the convert command for each path.
	for _, src := range srcs {
		if err := w.execFunc(src); err != nil {
			return fmt.Errorf("convert: %w", err)
		}
	}

	return nil
}

// Clean removes the `cwebp` binary from the user cache dir.
// .
func (w webp) Clean() error {
	return os.RemoveAll(filepath.Dir(w.path))
}

// convertFunc executes the `cwebp` binary to convert
// the `src` image into webp format.
// .
func (w webp) convertFunc(src string) error {
	// Execute the `cwebp` binary.
	if _, err := exec.Command(w.path, w.args(src)...).Output(); err != nil {
		return fmt.Errorf("failed webp conversion: %w", err)
	}

	// Remove the non-webp image.
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove original image: %w", err)
	}

	return nil
}

// args computes the arguments for the convert command.
// .
func (w webp) args(src string) []string {
	return []string{
		"-q", "85", src,
		"-o", strings.ReplaceAll(src, filepath.Ext(src), ".webp"),
	}
}

// unpack writes the `cwebp` binary to root dir.
//
//	$rootDir/gingersnap/cwebp
//
// .
func (w webp) unpack() error {
	if err := utils.EnsurePath(w.path); err != nil {
		return err
	}

	mode := os.FileMode(0755)

	if err := os.WriteFile(w.path, webpBinary, mode); err != nil {
		// return err
		return fmt.Errorf("unpack: %w", err)
	}

	if err := os.Chmod(w.path, mode); err != nil {
		return err
	}

	return nil
}
