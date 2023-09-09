package gingersnap

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"time"
)

// ------------------------------------------------------------------
//
//
// Embedded files
//
//
// ------------------------------------------------------------------

//go:embed "assets"
var Assets embed.FS

//go:embed "assets/templates"
var Templates embed.FS

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
	ListenAddr string
	Logger     *log.Logger

	Assets    embed.FS
	Templates *template.Template

	Config     *Config
	Posts      *PostModel
	Categories *CategoryModel
	Media      http.FileSystem

	HttpServer *http.Server
}

// ------------------------------------------------------------------
//
//
// Server HTTP Handlers
//
//
// ------------------------------------------------------------------

func (g *Gingersnap) Routes() http.Handler {
	r := http.NewServeMux()

	r.Handle("/", g.HandleIndex())
	r.Handle("/styles.css", g.ServeFile(g.Assets, "assets/css/styles.css"))
	r.Handle("/media/", g.CacheControl(http.StripPrefix("/media", http.FileServer(g.Media))))

	// Build category routes
	for _, cat := range g.Categories.All() {
		r.Handle(fmt.Sprintf("/category/%s/", cat.Slug), g.HandleCategory(cat))
	}

	// Build post routes
	for _, post := range g.Posts.All() {
		r.Handle(fmt.Sprintf("/%s/", post.Slug), g.HandlePost(post))
	}

	return g.RecoverPanic(g.LogRequest(g.SecureHeaders(r)))
}

func (g *Gingersnap) HandleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Handle 404
		if r.URL.Path != "/" {
			g.ErrNotFound(w)
			return
		}

		rd := g.NewRenderData(r)

		g.Render(w, http.StatusOK, "index", &rd)
	}
}

func (g *Gingersnap) HandlePost(post Post) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rd := g.NewRenderData(r)
		rd.Title = post.Title
		rd.Description = post.Description
		rd.Heading = post.Heading
		rd.Post = post
		rd.Image = post.Image
		rd.LatestPosts = g.Posts.Latest()
		rd.FeaturedPosts = g.Posts.Featured()

		g.Render(w, http.StatusOK, "post", &rd)
	}
}

func (g *Gingersnap) HandleCategory(cat Category) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the Posts by Category
		posts, ok := g.Posts.ByCategory(cat)
		if !ok {
			g.Logger.Printf("Cannot find Posts for Category '%s'", cat.Slug)
		}

		rd := g.NewRenderData(r)
		rd.Title = fmt.Sprintf("%s related Posts - Explore our Content on %s", cat.Title, g.Config.Site.Name)
		rd.Description = fmt.Sprintf("Browse through the %s category on %s and take a look at our posts.", cat.Title, g.Config.Site.Name)
		rd.Heading = cat.Title
		rd.Category = cat
		rd.Posts = posts

		g.Render(w, http.StatusOK, "category", &rd)
	}
}

func (g *Gingersnap) RunServer() {
	g.Logger.Printf("Starting server on %s", g.ListenAddr)

	err := g.HttpServer.ListenAndServe()
	if err != nil {
		g.Logger.Print(err)
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
func (g *Gingersnap) ServeFile(efs embed.FS, fileName string) http.Handler {
	ext := filepath.Ext(fileName)

	contentTypes := map[string]string{
		".css": "text/css; charset=utf-8",
		".txt": "text/plain; charset=utf-8",
		".xml": "application/xml; charset=utf-8",
	}

	// Check that the content type exists for the given extension.
	if _, ok := contentTypes[ext]; !ok {
		g.Logger.Fatalf("content type for [%s] not supported", ext)
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		data, err := efs.ReadFile(fileName)
		if err != nil {
			g.ErrInternalServer(w, err)
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

// ErrNotFound renders the 404.html template.
// .
func (g *Gingersnap) ErrNotFound(w http.ResponseWriter) {
	rd := g.NewRenderData(nil)
	rd.AppError = "404"
	rd.Title = fmt.Sprintf("Page Not Found - %s", g.Config.Site.Name)
	rd.LatestPosts = g.Posts.Latest()

	g.Render(w, http.StatusNotFound, "error", &rd)
}

// ErrInternalServer renders the 500.html template.
// If debug is enabled, then the stack trace is also shown.
// .
func (g *Gingersnap) ErrInternalServer(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	g.Logger.Output(2, trace)
	status := http.StatusInternalServerError

	rd := g.NewRenderData(nil)
	rd.AppError = "500"
	rd.Title = fmt.Sprintf("Internal Server Error - %s", g.Config.Site.Name)
	rd.LatestPosts = g.Posts.Latest()

	if g.Debug {
		rd.AppTrace = trace
	}

	g.Render(w, status, "error", &rd)
}

// errInternalServer writes the 500 error without using the render pipeline.
// This is used in the case where the `Render()` method fails.
// .
func (g *Gingersnap) errInternalServer(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	g.Logger.Output(2, trace)
	status := http.StatusInternalServerError

	if g.Debug {
		http.Error(w, trace, status)
		return
	}

	http.Error(w, http.StatusText(status), status)
}

// Render writes a template to the http.ResponseWriter.
// .
func (g *Gingersnap) Render(w http.ResponseWriter, status int, page string, data *RenderData) {
	buf := new(bytes.Buffer)

	// Write the template to the buffer first.
	// If error, then respond with a server error and return.
	err := g.Templates.ExecuteTemplate(buf, page, data)
	if err != nil {
		g.errInternalServer(w, err)
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

// LogResponseWriter allows us to capture the response status code.
// .
type LogResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *LogResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Logger is a middleware which logs the http request and response status.
// .
func (g *Gingersnap) LogRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ww := &LogResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Defer the logging call.
		defer func(start time.Time) {

			g.Logger.Printf(
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

// RecoverPanic is a middleware which recovers from panics and
// logs a HTTP 500 (Internal Server Error) if possible.
// .
func (g *Gingersnap) RecoverPanic(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		// Create a deferred function (which will always be run in the event
		// of a panic as Go uwinds the stack).
		defer func() {

			// Use the builtin recover function to check
			// if there has been a panic or not.
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				g.ErrInternalServer(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// SecureHeaders is a middleware which adds HTTP security headers
// to every response, inline with current OWASP guidance.
// .
func (g *Gingersnap) SecureHeaders(next http.Handler) http.Handler {
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

// CacheControl is a middleware which sets the caching policy for assets.
// .
func (g *Gingersnap) CacheControl(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if g.Debug {
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
func NewTemplate(files fs.FS) (*template.Template, error) {
	funcs := template.FuncMap{
		"safe": func(content string) template.HTML {
			return template.HTML(content)
		},
	}

	return template.New("").Funcs(funcs).ParseFS(files, "assets/templates/*.html")
}

// ------------------------------------------------------------------
//
//
// Type: Config
//
//
// ------------------------------------------------------------------

// Config stores project settings
// .
type Config struct {
	Site        Site
	NavbarLinks []Link
	FooterLinks []Link
}

// Site stores site-specific settings
// .
type Site struct {
	Name        string
	Host        string
	Tagline     string
	Description string
	Title       string
	Url         string
	Email       string
	Image       Image
}

// Link stores data for an anchor link
// .
type Link struct {
	Text  string
	Route string
}

// ------------------------------------------------------------------
//
//
// Type: Image
//
//
// ------------------------------------------------------------------

// Image represents an Image media object.
// .
type Image struct {
	Url    string
	Alt    string
	Type   string
	Width  string
	Height string
}

// ------------------------------------------------------------------
//
//
// RenderData
//
//
// ------------------------------------------------------------------

// RenderData stores the necessary data for template rendering.
// .
type RenderData struct {
	// Site info
	SiteHost    string
	SiteUrl     string
	SiteName    string
	SiteTagline string
	SiteEmail   string
	PageUrl     string

	// The meta image / lead image
	Image Image

	// The title tag, meta title tag and og:title tag
	Title string

	// The meta description
	Description string

	// The page heading, h1
	Heading string

	// The copyright year
	Copyright string

	// Post data
	Post          Post
	Posts         []Post
	LatestPosts   []Post
	RelatedPosts  []Post
	FeaturedPosts []Post

	// Category data
	Category   Category
	Categories []Category

	// Anchor links
	NavbarLinks []Link
	FooterLinks []Link

	// Application info
	AppDebug bool
	AppError string
	AppTrace string
}

func (g *Gingersnap) NewRenderData(r *http.Request) RenderData {
	pageUrl := ""
	if r != nil {
		pageUrl = fmt.Sprintf("%s%s", g.Config.Site.Url, r.URL.RequestURI())
	}

	return RenderData{
		SiteHost:    g.Config.Site.Host,
		SiteUrl:     g.Config.Site.Url,
		SiteName:    g.Config.Site.Name,
		SiteTagline: g.Config.Site.Tagline,
		SiteEmail:   g.Config.Site.Email,
		PageUrl:     pageUrl,

		Image: g.Config.Site.Image,

		Title:       g.Config.Site.Title,
		Description: g.Config.Site.Description,
		Heading:     g.Config.Site.Tagline,

		NavbarLinks: g.Config.NavbarLinks,
		FooterLinks: g.Config.FooterLinks,

		Copyright: fmt.Sprintf("2022 - %d", time.Now().Year()),
		AppDebug:  g.Debug,
	}
}

// ------------------------------------------------------------------
//
//
// Type: Post
//
//
// ------------------------------------------------------------------

// Post represents a blog or an article.
// .
type Post struct {
	// An FNV-1a hash of the slug.
	Hash int

	// The post slug.
	Slug string

	// The post meta title.
	Title string

	// The post heading, or main title.
	Heading string

	// The post description.
	Description string

	// The post category.
	Category Category

	// The lead image.
	Image Image

	// The post body.
	Body string

	// The date the post was or will be published.
	Pubdate string

	// The pubdate, as a UNIX timestamp.
	PubdateTS int
}

// ------------------------------------------------------------------
//
//
// Type: Category
//
//
// ------------------------------------------------------------------

// Category represents a post category.
// .
type Category struct {
	Slug  string
	Title string
}

// ------------------------------------------------------------------
//
//
// Type: PostModel
//
//
// ------------------------------------------------------------------

// PostModel manages queries for Post resources.
// .
type PostModel struct {
	posts           []Post
	postsBySlug     map[string]Post
	postsByCategory map[Category][]Post
}

func (m *PostModel) All() []Post {
	return m.posts
}

func (m *PostModel) Slugs() []string {
	var slugs []string
	for _, post := range m.posts {
		slugs = append(slugs, post.Slug)
	}
	return slugs
}

func (m *PostModel) Latest() []Post {
	return nil
}

func (m *PostModel) Featured() []Post {
	return nil
}

func (m *PostModel) Related(p Post) []Post {
	return nil
}

func (m *PostModel) ByCategory(c Category) ([]Post, bool) {
	posts, ok := m.postsByCategory[c]
	return posts, ok
}

func (m *PostModel) BySlug(s string) (Post, bool) {
	post, ok := m.postsBySlug[s]
	return post, ok
}

// ------------------------------------------------------------------
//
//
// Type: CategoryModel
//
//
// ------------------------------------------------------------------

// CategoryModel manages queries for Post resources.
// .
type CategoryModel struct {
	categories       []Category
	categoriesBySlug map[string]Category
}

func (m *CategoryModel) All() []Category {
	return m.categories
}

func (m *CategoryModel) BySlug(s string) (Category, bool) {
	category, ok := m.categoriesBySlug[s]
	return category, ok
}
