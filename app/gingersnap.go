package app

import (
	"bytes"
	"embed"
	"fmt"
	htmlTmp "html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	textTmp "text/template"
	"time"

	"gingersnap/app/utils"
)

//go:embed "assets"
var assets embed.FS

//go:embed "assets/templates"
var templates embed.FS

// ------------------------------------------------------------------
//
//
// Type: Gingersnap
//
//
// ------------------------------------------------------------------

// Gingersnap is the main application engine.
// .
type Gingersnap struct {
	Debug      bool
	ConfigPath string
	PostsPath  string
	MediaPath  string
	ExportPath string

	// The main logger
	logger *log.Logger

	// Internal assets
	assets embed.FS

	// Internal templates
	templates *htmlTmp.Template

	// Internal HTTP server.
	httpServer *http.Server

	// The user's config
	config *config

	// The user's content
	store *store

	// The user's media
	media http.FileSystem
}

// NewGingersnap returns a *Gingernap engine.
// .
func NewGingersnap() *Gingersnap {
	return &Gingersnap{
		Debug:      true,
		ConfigPath: "assets/config/gingersnap.json",
		PostsPath:  "assets/posts",
		MediaPath:  "assets/media",
		ExportPath: "dist",
	}
}

// Configure wipes and reconfigures the gingersnap engine.
// .
func (g *Gingersnap) Configure() {

	// [1/3] Wipe the gingersnap engine -------------------

	if g.httpServer != nil {
		g.httpServer.Close()
		g.httpServer = nil
	}
	g.logger = nil
	g.templates = nil
	g.config = nil
	g.store = nil

	// [2/3] Configure the engine components --------------

	// Construct the logger.
	logger := log.New(os.Stderr, "", log.Ltime)

	// Read the config file.
	configBytes, err := utils.ReadFile(g.ConfigPath)
	if err != nil {
		logger.Fatalf("read config: %s\n", err)
	}

	// Construct the config.
	config, err := newConfig(configBytes, g.Debug)
	if err != nil {
		logger.Fatalf("parse config: %s", err)
	}

	// Gather the markdown post files.
	filePaths, err := utils.LocalGlob(g.PostsPath, "md")
	if err != nil {
		logger.Fatalf("gather posts: %s", err)
	}

	// Parse the markdown posts.
	pr := newProcessor(filePaths)
	if err := pr.process(); err != nil {
		logger.Fatalf("process posts: %s", err)
	}

	// Construct the store from the processed posts.
	store := newStore()
	store.InitPosts(pr.postsBySlug)
	store.InitCategories(pr.categoriesBySlug)
	store.InitSections()

	// Construct the templates, using the embedded FS.
	templates, err := newTemplate(templates)
	if err != nil {
		logger.Fatalf("parse templates: %s", err)
	}

	// [3/3] Construct the gingersnap engine --------------

	g.logger = logger
	g.assets = assets
	g.media = http.Dir(g.MediaPath)
	g.templates = templates
	g.config = config
	g.store = store
	g.httpServer = &http.Server{
		Addr:         g.config.ListenAddr,
		Handler:      g.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// Export exports the server as a static site.
// .
func (g *Gingersnap) Export() error {

	ex, err := g.newExporter()
	if err != nil {
		return err
	}

	return ex.export()
}

// Repository returns the path of the git repository
// where the static site will be managed.
// .
func (g *Gingersnap) Repository() string {
	return g.config.Repository
}

// RunServer runs the gingersnap server.
// .
func (g *Gingersnap) RunServer() {
	g.logger.Printf("Starting server on %s 🤖\n\n", g.config.ListenAddr)
	g.httpServer.ListenAndServe()
}

// Unpack copies the relevant resources into the current directory.
// This is used when initializing a new gingersnap project.
// .
func (g *Gingersnap) Unpack() {
	utils.CopyDir(assets, "assets/media", ".")
	utils.CopyDir(assets, "assets/posts", ".")
	utils.CopyFile(assets, "assets/config/gingersnap.json", ".")
}

// ------------------------------------------------------------------
//
//
// Server HTTP Routes
//
//
// ------------------------------------------------------------------

// Routes constructs and returns the complete http.Handler for the server.
// .
func (g *Gingersnap) routes() http.Handler {
	r := http.NewServeMux()

	r.Handle("/", g.handleIndex())
	r.Handle("/styles.css", g.serveFile(g.assets, "assets/css/styles.css"))
	r.Handle("/sitemap/", g.handleSitemapHtml())
	r.Handle("/sitemap.xml", g.handleSitemapXml())
	r.Handle("/robots.txt", g.handleRobotsTxt())
	r.Handle("/CNAME", g.handleCname())
	r.Handle("/404/", g.handle404())
	r.Handle("/media/", g.cacheControl(http.StripPrefix("/media", http.FileServer(g.media))))

	// Build routes for all blog posts.
	for _, p := range g.store.posts {
		r.Handle(p.route(), g.handlePost(p))
	}

	// Build routes for all standalone posts (pages).
	for _, p := range g.store.pages {
		r.Handle(p.route(), g.handlePost(p))
	}

	// Build category routes
	for _, cat := range g.store.categories {
		r.Handle(cat.route(), g.handleCategory(cat))
	}

	return g.recoverPanic(g.logRequest(g.secureHeaders(r)))
}

// ------------------------------------------------------------------
//
//
// Server HTTP Handlers
//
//
// ------------------------------------------------------------------

func (g *Gingersnap) handleIndex() http.HandlerFunc {
	sections := make([]section, 0, len(g.config.Homepage))

	// Create sections for rendering the homepage.
	for _, slug := range g.config.Homepage {

		section, ok := g.store.sections[slug]
		if !ok {
			panic(fmt.Sprintf("cannot find Section '%s'", slug))
		}
		sections = append(sections, section)
	}

	return func(w http.ResponseWriter, r *http.Request) {

		// Handle 404
		if r.URL.Path != "/" {
			g.errNotFound(w)
			return
		}

		rd := g.newRenderData(r)
		rd.Sections = sections

		g.render(w, http.StatusOK, "index", &rd)
	}
}

func (g *Gingersnap) handlePost(post *post) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rd := g.newRenderData(r)
		rd.Title = post.title
		rd.Description = post.description
		rd.Heading = post.heading
		rd.Post = post
		rd.LatestPosts = g.store.postsLatestSm
		rd.RelatedPosts = g.store.RelatedPosts(post)

		if post.image.isEmpty() {
			rd.Image = g.config.Site.Image
		} else {
			rd.Image = post.image
		}

		g.render(w, http.StatusOK, "post", &rd)
	}
}

func (g *Gingersnap) handleCategory(cat category) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the Posts by Category
		posts, ok := g.store.postsByCategory[cat]
		if !ok {
			g.logger.Printf("Cannot find Posts for Category '%s'", cat.Slug)
			g.errNotFound(w)
			return
		}

		rd := g.newRenderData(r)
		rd.Title = fmt.Sprintf("%s related Posts - Explore our Content on %s", cat.Title, g.config.Site.Name)
		rd.Description = fmt.Sprintf("Browse through the %s category on %s and take a look at our posts.", cat.Title, g.config.Site.Name)
		rd.Heading = cat.Title
		rd.Category = cat
		rd.Posts = posts

		g.render(w, http.StatusOK, "category", &rd)
	}
}

func (g *Gingersnap) handleCname() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(g.config.Site.Host))
	}
}

func (g *Gingersnap) handle404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		g.render404(w, http.StatusOK)
	}
}

func (g *Gingersnap) handleRobotsTxt() http.HandlerFunc {

	robotsTemplate := `
User-agent: *
Disallow:

Sitemap: {{.}}/sitemap.xml
`

	// Prepare the robots template.
	tmpl, err := textTmp.New("").Parse(strings.TrimPrefix(robotsTemplate, "\n"))
	if err != nil {
		g.logger.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)

		// Write the template to the buffer first.
		// If error, then respond with a server error and return.
		if err := tmpl.Execute(buf, g.config.Site.Url); err != nil {
			g.internalServerError(w, err)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

func (g *Gingersnap) handleSitemapXml() http.HandlerFunc {

	sitemapTemplate := `
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	{{- range $key, $value := .}}
	<url>
		<loc>{{$key}}</loc>
		{{if $value}}<lastmod>{{$value}}</lastmod>{{end}}
	</url>
	{{- end}}
</urlset>
`

	// Prepare the sitemap template.
	tmpl, err := textTmp.New("").Parse(strings.TrimPrefix(sitemapTemplate, "\n"))
	if err != nil {
		panic(err)
	}

	// The urlSet is a map of urls to lastmod dates.
	// It is used to render the sitemap.
	urlSet := make(map[string]string, len(g.store.posts)*2)

	// permalink is a helper function which generates
	// the permalink for a given path
	permalink := func(urlPath string) string {
		return fmt.Sprintf("%v%v", g.config.Site.Url, urlPath)
	}

	// Add sitemap entries for the index page.
	urlSet[permalink("/")] = ""

	// Add sitemap entries for all the blog posts.
	for _, post := range g.store.posts {
		lastMod := ""

		if ts := post.latestTS(); ts > 0 {
			lastMod = time.Unix(int64(ts), 0).UTC().Format("2006-01-02T00:00:00+00:00")
		}

		urlSet[permalink(post.route())] = lastMod
	}

	// Add sitemap entries for all the standalone posts (pages).
	for _, post := range g.store.pages {
		urlSet[permalink(post.route())] = ""
	}

	// Add sitemap entries for all the categories.
	for _, cat := range g.store.categories {
		urlSet[permalink(cat.route())] = ""
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)

		// Write the template to the buffer first.
		// If error, then respond with a server error and return.
		if err := tmpl.Execute(buf, urlSet); err != nil {
			g.internalServerError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

func (g *Gingersnap) handleSitemapHtml() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		rd := g.newRenderData(r)
		rd.Title = fmt.Sprintf("Sitemap - Browse through all Posts on %s", g.config.Site.Name)
		rd.Description = fmt.Sprintf("Browse through the sitemap on %s and take a look at our posts.", g.config.Site.Name)
		rd.Heading = "Posts"
		rd.Posts = g.store.posts

		g.render(w, http.StatusOK, "sitemap", &rd)
	}
}

// ------------------------------------------------------------------
//
//
// Server Reusable HTTP Handlers
//
//
// ------------------------------------------------------------------

// ServeFile returns a http.Handler that serves a specific file.
// .
func (g *Gingersnap) serveFile(efs embed.FS, fileName string) http.Handler {
	ext := filepath.Ext(fileName)

	contentTypes := map[string]string{
		".css": "text/css; charset=utf-8",
		".txt": "text/plain; charset=utf-8",
		".xml": "application/xml; charset=utf-8",
	}

	// Check that the content type exists for the given extension.
	if _, ok := contentTypes[ext]; !ok {
		g.logger.Fatalf("content type for [%s] not supported", ext)
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		data, err := efs.ReadFile(fileName)
		if err != nil {
			g.errInternalServer(w, err)
		}
		w.Header().Set("Content-Type", contentTypes[ext])
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
	return http.HandlerFunc(fn)
}

// ------------------------------------------------------------------
//
//
// Render helpers
//
//
// ------------------------------------------------------------------

// errNotFound renders the 404.html template.
// .
func (g *Gingersnap) errNotFound(w http.ResponseWriter) {
	g.render404(w, http.StatusNotFound)
}

// render404 renders the error template with the assigned status code.
// Usually, we render the not-found template with a 404 error code.
// But when exporting the site, we need to render the not-found template with a 200 instead.
// .
func (g *Gingersnap) render404(w http.ResponseWriter, status int) {
	rd := g.newRenderData(nil)
	rd.AppError = "404"
	rd.Title = fmt.Sprintf("Page Not Found - %s", g.config.Site.Name)
	rd.LatestPosts = g.store.postsLatest

	g.render(w, status, "error", &rd)
}

// errInternalServer renders the 500.html template.
// If debug is enabled, then the stack trace is also shown.
// .
func (g *Gingersnap) errInternalServer(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	g.logger.Output(2, trace)
	status := http.StatusInternalServerError

	rd := g.newRenderData(nil)
	rd.AppError = "500"
	rd.Title = fmt.Sprintf("Internal Server Error - %s", g.config.Site.Name)
	rd.LatestPosts = g.store.postsLatest

	if g.config.Debug {
		rd.AppTrace = trace
	}

	g.render(w, status, "error", &rd)
}

// internalServerError writes the 500 error without using the render pipeline.
// This is used in the case where the `Render()` method fails.
// .
func (g *Gingersnap) internalServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	g.logger.Output(2, trace)
	status := http.StatusInternalServerError

	if g.config.Debug {
		http.Error(w, trace, status)
		return
	}

	http.Error(w, http.StatusText(status), status)
}

// Render writes a template to the http.ResponseWriter.
// .
func (g *Gingersnap) render(w http.ResponseWriter, status int, page string, data *renderData) {
	buf := new(bytes.Buffer)

	// Write the template to the buffer first.
	// If error, then respond with a server error and return.
	if err := g.templates.ExecuteTemplate(buf, page, data); err != nil {
		g.internalServerError(w, err)
		return
	}

	w.WriteHeader(status)

	// Write the contents of the buffer to the http.ResponseWriter.
	buf.WriteTo(w)
}

// ------------------------------------------------------------------
//
//
// Server Middleware
//
//
// ------------------------------------------------------------------

// logResponseWriter allows us to capture the response status code.
// .
type logResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *logResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Logger is a middleware which logs the http request and response status.
// .
func (g *Gingersnap) logRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ww := &logResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Defer the logging call.
		defer func(start time.Time) {

			g.logger.Printf(
				"%s %d %s %s %s",
				"[request]",
				ww.status,
				r.Method,
				r.URL.RequestURI(),
				time.Since(start),
			)

		}(time.Now())

		// Call the next handler
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}

// recoverPanic is a middleware which recovers from panics and
// logs a HTTP 500 (Internal Server Error) if possible.
// .
func (g *Gingersnap) recoverPanic(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Create a deferred function (which will always be run in the event
		// of a panic as Go uwinds the stack).
		defer func() {

			// Use the builtin recover function to check
			// if there has been a panic or not.
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				g.errInternalServer(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// secureHeaders is a middleware which adds HTTP security headers
// to every response, inline with current OWASP guidance.
// .
func (g *Gingersnap) secureHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("X-Powered-By", "Go")
		w.Header().Set("X-Built-By", "ad9280c159074d9ec90899b584f520606e83d10e")

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// cacheControl is a middleware which sets the caching policy for assets.
// .
func (g *Gingersnap) cacheControl(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if g.config.Debug {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "max-age=172800") // 2 days
		}
		next.ServeHTTP(w, r)
	}
}

// ------------------------------------------------------------------
//
//
// Templates
//
//
// ------------------------------------------------------------------

// NewTemplate parses and loads all templates from
// the the given filesystem interface.
// .
func newTemplate(files fs.FS) (*htmlTmp.Template, error) {
	funcs := htmlTmp.FuncMap{
		"safe": func(content string) htmlTmp.HTML {
			return htmlTmp.HTML(content)
		},
	}

	return htmlTmp.New("").Funcs(funcs).ParseFS(files, "assets/templates/*.html")
}
