package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gingersnap/app/utils"
)

// ------------------------------------------------------------------
//
//
// Type: exporter
//
//
// ------------------------------------------------------------------

// exporter is responsible for exporting the server as a static site.
//
// Main method: `export()`
// .
type exporter struct {
	// The http handler responsible for rendering the urls.
	handler http.Handler

	// The set of URLs to export.
	urls []string

	// The directory where the site will be exported to.
	outputPath string
}

// newExporter constructs and returns an *exporter
// with the server's handlers, urls and export dir.
// .
func (g *Gingersnap) newExporter() (*exporter, error) {

	// [1/2] Collect the urls to export -------------------

	urls := make([]string, 0, max(len(g.store.posts), 20))
	urls = append(urls, "/", "/styles.css", "/sitemap/", "/sitemap.xml", "/robots.txt", "/CNAME", "/404/")

	// For "/media/", we read media files
	// directly from the filesystem.

	// Open the media FS.
	f, err := g.media.Open(".")
	if err != nil {
		return nil, err
	}

	// Read the media directory.
	files, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Build routes for all media files.
	for _, file := range files {
		if name := file.Name(); !strings.HasPrefix(name, ".") {
			urls = append(urls, fmt.Sprintf("/media/%s", name))
		}
	}

	// Build routes for all blog posts.
	for _, post := range g.store.posts {
		urls = append(urls, post.route())
	}

	// Build routes for all standalone posts (pages).
	for _, post := range g.store.pages {
		urls = append(urls, post.route())
	}

	// Build routes for all categories.
	for _, cat := range g.store.categories {
		urls = append(urls, cat.route())
	}

	// [2/2] Construct the exporter -----------------------

	return &exporter{
		handler:    g.routes(),
		urls:       urls,
		outputPath: g.ExportPath,
	}, nil
}

// export exports the configured server routes as a static site.
// .
func (e *exporter) export() error {
	// Remove the output path, if it exists.
	if err := os.RemoveAll(e.outputPath); err != nil {
		return err
	}

	// Create the output directory.
	if err := utils.EnsurePath(e.outputPath); err != nil {
		return err
	}

	// Write the timestamp file.
	tsPath := filepath.Join(e.outputPath, ".gingersnap")
	ts := time.Now().Format(time.UnixDate)

	if err := utils.WriteFile(tsPath, []byte(ts)); err != nil {
		return err
	}

	// Render all the paths.
	for _, url := range e.urls {
		if err := e.exportPage(url, e.makePath(url)); err != nil {
			return err
		}
	}

	return nil
}

// exportPage exports a single page to the filesystem
// by using the httptest framework to render the page.
// .
func (e *exporter) exportPage(url, dstPath string) error {
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	e.handler.ServeHTTP(w, r)

	// For nested media assets like `/media/other/book.webp`
	// the exporter will attempt to request `/media/other/`.
	//
	// However, partial paths return an http 301, so we have
	// to account for those status codes and skip them.

	if c := w.Result().StatusCode; c != http.StatusOK && c != http.StatusMovedPermanently {
		return fmt.Errorf("expected URL %s to return %d, but it returned %d instead", url, http.StatusOK, c)
	}

	// Create the destination directory.
	if err := utils.EnsurePath(dstPath); err != nil {
		return err
	}

	// Write the contents of the response body.
	if err := utils.WriteFile(dstPath, w.Body.Bytes()); err != nil {
		return err
	}

	return nil
}

// makePath builds the output path based on the given route being rendered.
//
// Ex:
//
//	/404/             =>  /404.html
//	/CNAME            =>  /CNAME
//	/some-post/       =>  /some-post/index.html
//
// .
func (e *exporter) makePath(url string) string {
	if url == "/CNAME" {
		return filepath.Join(e.outputPath, "/CNAME")
	}

	if url == "/404/" {
		return filepath.Join(e.outputPath, "404.html")
	}

	p := filepath.Join(e.outputPath, url)

	if ext := filepath.Ext(p); ext == "" {
		p = filepath.Join(p, "index.html")
	}

	return p
}
