package gingersnap

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	htmlTemplate "html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
	high "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	"go.abhg.dev/goldmark/frontmatter"
)

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
	// The main logger
	Logger *log.Logger

	// Internal assets
	Assets embed.FS

	// Internal templates
	Templates *htmlTemplate.Template

	// Internal HTTP server.
	HttpServer *http.Server

	// The user's config
	Config *Config

	// The user's content
	Store *Store

	// The user's media
	Media http.FileSystem
}

// New returns an empty Gingernap engine.
// .
func New() *Gingersnap {
	return &Gingersnap{}
}

// Routes constructs and returns the complete http.Handler for the server.
// .
func (g *Gingersnap) Routes() http.Handler {
	r := http.NewServeMux()

	r.Handle("/", g.HandleIndex())
	r.Handle("/styles.css", g.ServeFile(g.Assets, "assets/css/styles.css"))
	r.Handle("/sitemap/", g.HandleSitemapHtml())
	r.Handle("/sitemap.xml", g.HandleSitemapXml())
	r.Handle("/robots.txt", g.HandleRobotsTxt())
	r.Handle("/CNAME", g.HandleCname())
	r.Handle("/404/", g.Handle404())
	r.Handle("/media/", g.CacheControl(http.StripPrefix("/media", http.FileServer(g.Media))))

	// Build routes for all blog posts.
	for _, post := range g.Store.Posts {
		r.Handle(post.Route(), g.HandlePost(post))
	}

	// Build routes for all standalone posts (pages).
	for _, post := range g.Store.Pages {
		r.Handle(post.Route(), g.HandlePost(post))
	}

	// Build category routes
	for _, cat := range g.Store.Categories {
		r.Handle(cat.Route(), g.HandleCategory(cat))
	}

	return g.RecoverPanic(g.LogRequest(g.SecureHeaders(r)))
}

// AllUrls returns all urls of the server.
// .
func (g *Gingersnap) AllUrls() ([]string, error) {

	urls := make([]string, 0, max(len(g.Store.Posts), 20))
	urls = append(urls, "/", "/styles.css", "/sitemap/", "/sitemap.xml")
	urls = append(urls, "/robots.txt", "/CNAME", "/404/")

	// Note: for "/media/", we read media files
	//       directly from the filesystem.

	// Open the media directory.
	f, err := g.Media.Open(".")
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
	for _, post := range g.Store.Posts {
		urls = append(urls, post.Route())
	}

	// Build routes for all standalone posts (pages).
	for _, post := range g.Store.Pages {
		urls = append(urls, post.Route())
	}

	// Build routes for all categories.
	for _, cat := range g.Store.Categories {
		urls = append(urls, cat.Route())
	}

	return urls, nil
}

// ------------------------------------------------------------------
//
//
// Server HTTP Handlers
//
//
// ------------------------------------------------------------------

func (g *Gingersnap) HandleIndex() http.HandlerFunc {
	sections := make([]Section, 0, len(g.Config.Homepage))

	// Create sections for rendering the homepage.
	for _, slug := range g.Config.Homepage {

		// If specified, then gather the "latest" posts.
		if slug == SectionLatest {

			categoryLatest := Category{
				Slug:  "",
				Title: "Latest Posts",
			}

			// Create the section.
			section := Section{
				Category: categoryLatest,
				Posts:    g.Store.PostsLatest,
			}

			// Add the section.
			sections = append(sections, section)
			continue
		}

		// If specified, then gather the "featured" posts.
		if slug == SectionFeatured {

			categoryFeatured := Category{
				Slug:  "",
				Title: "Featured Posts",
			}

			// Create the section.
			section := Section{
				Category: categoryFeatured,
				Posts:    g.Store.PostsFeatured,
			}

			// Add the section.
			sections = append(sections, section)
			continue
		}

		// Gather the posts for the specified category.
		// Raise error if category is not found.
		cat, ok := g.Store.CategoriesBySlug[slug]
		if !ok {
			panic(fmt.Sprintf("homepage section error: cannot find category [%s]", slug))
		}

		posts, ok := g.Store.PostsByCategory[cat]
		if !ok {
			panic(fmt.Sprintf("homepage section error: no posts found for category [%s]", slug))
		}

		// Create the section.
		section := Section{
			Category: cat,
			Posts:    posts[:min(LimitSection, len(posts))],
		}

		// Add the section.
		sections = append(sections, section)

	}

	return func(w http.ResponseWriter, r *http.Request) {

		// Handle 404
		if r.URL.Path != "/" {
			g.ErrNotFound(w)
			return
		}

		rd := g.NewRenderData(r)
		rd.Sections = sections

		g.Render(w, http.StatusOK, "index", &rd)
	}
}

func (g *Gingersnap) HandlePost(post *Post) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rd := g.NewRenderData(r)
		rd.Title = post.Title
		rd.Description = post.Description
		rd.Heading = post.Heading
		rd.Post = post
		rd.LatestPosts = g.Store.PostsLatest[:min(LimitLatestPostDetail, len(g.Store.PostsLatest))]
		rd.RelatedPosts = g.Store.RelatedPosts(post)

		if post.Image.IsEmpty() {
			rd.Image = g.Config.Site.Image
		} else {
			rd.Image = post.Image
		}

		g.Render(w, http.StatusOK, "post", &rd)
	}
}

func (g *Gingersnap) HandleCategory(cat Category) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the Posts by Category
		posts, ok := g.Store.PostsByCategory[cat]
		if !ok {
			g.Logger.Printf("Cannot find Posts for Category '%s'", cat.Slug)
			g.ErrNotFound(w)
			return
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

func (g *Gingersnap) HandleCname() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(g.Config.Site.Host))
	}
}

func (g *Gingersnap) Handle404() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		g.render404(w, http.StatusOK)
	}
}

const RobotsTemplate = `
User-agent: *
Disallow:

Sitemap: {{.}}/sitemap.xml
`

func (g *Gingersnap) HandleRobotsTxt() http.HandlerFunc {

	// Prepare the robots template.
	tmpl, err := textTemplate.New("").Parse(strings.TrimPrefix(RobotsTemplate, "\n"))
	if err != nil {
		g.Logger.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)

		// Write the template to the buffer first.
		// If error, then respond with a server error and return.
		if err := tmpl.Execute(buf, g.Config.Site.Url); err != nil {
			g.errInternalServer(w, err)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

const SitemapTemplate = `
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

func (g *Gingersnap) HandleSitemapXml() http.HandlerFunc {

	// Prepare the sitemap template.
	tmpl, err := textTemplate.New("").Parse(strings.TrimPrefix(SitemapTemplate, "\n"))
	if err != nil {
		panic(err)
	}

	// The urlSet is a map of urls to lastmod dates.
	// It is used to render the sitemap.
	urlSet := make(map[string]string, len(g.Store.Posts)*2)

	// permalink is a helper function which generates
	// the permalink for a given path
	permalink := func(urlPath string) string {
		return fmt.Sprintf("%v%v", g.Config.Site.Url, urlPath)
	}

	// Add sitemap entries for the index page.
	urlSet[permalink("/")] = ""

	// Add sitemap entries for all the blog posts.
	for _, post := range g.Store.Posts {
		lastMod := ""

		if ts := post.LatestTS(); ts > 0 {
			lastMod = time.Unix(int64(ts), 0).UTC().Format("2006-01-02T00:00:00+00:00")
		}

		urlSet[permalink(post.Route())] = lastMod
	}

	// Add sitemap entries for all the standalone posts (pages).
	for _, post := range g.Store.Pages {
		urlSet[permalink(post.Route())] = ""
	}

	// Add sitemap entries for all the categories.
	for _, cat := range g.Store.Categories {
		urlSet[permalink(cat.Route())] = ""
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)

		// Write the template to the buffer first.
		// If error, then respond with a server error and return.
		if err := tmpl.Execute(buf, urlSet); err != nil {
			g.errInternalServer(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(buf.Bytes())
	}
}

func (g *Gingersnap) HandleSitemapHtml() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		rd := g.NewRenderData(r)
		rd.Title = fmt.Sprintf("Sitemap - Browse through all Posts on %s", g.Config.Site.Name)
		rd.Description = fmt.Sprintf("Browse through the sitemap on %s and take a look at our posts.", g.Config.Site.Name)
		rd.Heading = "Posts"
		rd.Posts = g.Store.Posts

		g.Render(w, http.StatusOK, "sitemap", &rd)
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
	g.render404(w, http.StatusNotFound)
}

// render404 renders the error template with the assigned status code.
// Usually, we render the not-found template with a 404 error code.
// But when exporting the site, we need to render the not-found template with a 200 instead.
// .
func (g *Gingersnap) render404(w http.ResponseWriter, status int) {
	rd := g.NewRenderData(nil)
	rd.AppError = "404"
	rd.Title = fmt.Sprintf("Page Not Found - %s", g.Config.Site.Name)
	rd.LatestPosts = g.Store.PostsLatest

	g.Render(w, status, "error", &rd)
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
	rd.LatestPosts = g.Store.PostsLatest

	if g.Config.Debug {
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

	if g.Config.Debug {
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
	if err := g.Templates.ExecuteTemplate(buf, page, data); err != nil {
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
		if g.Config.Debug {
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
func NewTemplate(files fs.FS) (*htmlTemplate.Template, error) {
	funcs := htmlTemplate.FuncMap{
		"safe": func(content string) htmlTemplate.HTML {
			return htmlTemplate.HTML(content)
		},
	}

	return htmlTemplate.New("").Funcs(funcs).ParseFS(files, "assets/templates/*.html")
}

// NewLogger constructs and returns a new Logger.
// .
func NewLogger() *log.Logger {
	return log.New(os.Stderr, "", log.Ltime)
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
	Debug      bool
	ListenAddr string

	// Site-specific settings
	Site Site `json:"site"`

	// Site styling
	Theme Theme

	// Homepage sections
	Homepage []string `json:"homepage"`

	// Anchor links for the navbar
	NavbarLinks []Link `json:"navbarLinks"`

	// Anchor links for the footer
	FooterLinks []Link `json:"footerLinks"`

	// The git repository where the production static site will be managed
	ProdRepo string `json:"staticRepository"`
}

// Site stores site-specific settings.
// .
type Site struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Tagline     string `json:"tagline"`
	Description string `json:"description"`
	Title       string
	Url         string
	Email       string
	Image       Image
	Theme       string `json:"theme"`
}

// Link stores data for an anchor link.
// .
type Link struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

// NewConfig parses the settings into a Config struct.
// .
func NewConfig(configBytes []byte, debug bool) (*Config, error) {

	config := &Config{
		Debug:      debug,
		ListenAddr: ":4000",
	}

	// Parse the config file.
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}

	config.Site.Url = fmt.Sprintf("https://%s", config.Site.Host)
	config.Site.Email = fmt.Sprintf("admin@%s", config.Site.Host)
	config.Site.Title = fmt.Sprintf("%s - %s", config.Site.Name, config.Site.Tagline)
	config.Site.Image = Image{
		Url:    "/media/meta-img.webp",
		Alt:    config.Site.Title,
		Type:   ImageType,
		Width:  ImageWidth,
		Height: ImageHeight,
	}

	// In "debug" mode, we change the host to localhost.
	if config.Debug {
		config.Site.Host = fmt.Sprintf("localhost%s", config.ListenAddr)
		config.Site.Url = fmt.Sprintf("http://%s", config.Site.Host)
		config.Site.Email = fmt.Sprintf("admin@%s", config.Site.Host)
	}

	// If no Homepage sections are defined, then create
	// a default setup with the "$latest" posts only.
	if config.Homepage == nil {
		config.Homepage = []string{SectionLatest}
	}

	// Retrieve the theme.
	theme, err := NewTheme(config.Site.Theme)
	if err != nil {
		return nil, err
	}

	config.Theme = theme

	return config, nil
}

// ------------------------------------------------------------------
//
//
// Type: Theme
//
//
// ------------------------------------------------------------------

// Theme represents a simple color profile for the site.
// .
type Theme struct {
	Heading   string
	Primary   string
	Secondary string
	Link      string
}

// NewTheme retrieves the requested theme.
// Simplified themes can also be requested. A simplified theme
// modifies the secondary color, which makes it less colorful.
//
// ex:
//
//	// will retrieve the pink theme
//	NewTheme("pink")
//
//	// will retrieve the pink theme and modify it
//	NewTheme("pink-simple")
//
// .
func NewTheme(themeStr string) (Theme, error) {

	// If no theme is given, then return the default.
	if themeStr == "" {
		return DefaultTheme, nil
	}

	// Determine if a "simple" theme is requested.
	// With "simple" themes, the secondary color
	// is modified to make it less colorful.
	themeStr, isSimple := strings.CutSuffix(themeStr, "-simple")

	// Retrieve the theme.
	if t, ok := Themes[themeStr]; ok {

		theme := Theme{
			Primary:   t.Primary,
			Secondary: t.Secondary,
			Link:      t.Link,
		}

		// If a "simple" theme is requested,
		// then change the secondary color to black.
		if isSimple {
			theme.Secondary = "#0f172a" // slate-900
		}

		return theme, nil
	}

	return Theme{}, fmt.Errorf("Could not load theme [%s]", themeStr)
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

// IsEmpty reports if the image is empty.
// .
func (i *Image) IsEmpty() bool {
	return i.Url == ""
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
	Post          *Post
	Posts         []*Post
	LatestPosts   []*Post
	RelatedPosts  []*Post
	FeaturedPosts []*Post

	// Category data
	Category   Category
	Categories []Category

	// Layout and styling
	Sections    []Section
	NavbarLinks []Link
	FooterLinks []Link
	Theme       Theme

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
		Theme:       g.Config.Theme,

		Copyright: fmt.Sprintf("%d", time.Now().Year()),
		AppDebug:  g.Config.Debug,
	}
}

// ------------------------------------------------------------------
//
//
// Type: Post
//
//
// ------------------------------------------------------------------

// Post represents an article or a web page.
// .
type Post struct {
	// If the Post is a standalone post (page)
	IsPage bool

	// If the Post is a blog post
	IsBlog bool

	// If the post is a featured post
	IsFeatured bool

	// If the lead image should be displayed
	ShowLead bool

	// The post slug
	Slug string

	// The post meta title
	Title string

	// The post heading, or main title
	Heading string

	// The post description
	Description string

	// The post category
	Category Category

	// The lead image
	Image Image

	// The post body
	Body string

	// The publish date - January 2, 2006
	Pubdate string

	// The publish date, as a UNIX timestamp
	PubdateTS int

	// The updated date - January 2, 2006
	Updated string

	// The updated date, as a UNIX timestamp
	UpdatedTS int
}

// LatestTS returns the Post's latest timestamped date.
// .
func (p *Post) LatestTS() int {
	if p.UpdatedTS > p.PubdateTS {
		return p.UpdatedTS
	}
	return p.PubdateTS
}

// Route returns the url path for the Post.
//
// ex: "/post-slug/"
// .
func (p *Post) Route() string {
	return fmt.Sprintf("/%s/", p.Slug)
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

// IsEmpty reports if the category is empty.
// .
func (c Category) IsEmpty() bool {
	return c.Slug == ""
}

// Route returns the url path for the Category.
//
// ex: "/category/some-slug/"
// .
func (c Category) Route() string {
	return fmt.Sprintf("/category/%s/", c.Slug)
}

// ------------------------------------------------------------------
//
//
// Type: Section
//
//
// ------------------------------------------------------------------

// Section represents a section of a web page.
// It stores a collection of posts for a category.
// .
type Section struct {
	Category Category
	Posts    []*Post
}

// ------------------------------------------------------------------
//
//
// Type: Store
//
//
// ------------------------------------------------------------------

// Store is responsible for storing Posts and Categories
// in an organized way, making easy to access.
// .
type Store struct {
	Posts           []*Post
	Pages           []*Post
	PostsLatest     []*Post
	PostsFeatured   []*Post
	PostsBySlug     map[string]*Post
	PostsByCategory map[Category][]*Post

	Categories       []Category
	CategoriesBySlug map[string]Category
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) InitCategories(categoriesBySlug map[string]Category) {

	s.CategoriesBySlug = categoriesBySlug

	s.Categories = make([]Category, 0, len(s.CategoriesBySlug))

	for slug := range s.CategoriesBySlug {
		cat := s.CategoriesBySlug[slug]
		s.Categories = append(s.Categories, cat)
	}
}

func (s *Store) InitPosts(postsBySlug map[string]*Post) {

	// Note: the `postsBySlug` map contains
	//       blog posts AND standalone posts (pages).

	s.PostsBySlug = postsBySlug

	postsLen := len(s.PostsBySlug)

	s.Posts = make([]*Post, 0, postsLen)
	s.Pages = make([]*Post, 0, postsLen)
	s.PostsLatest = make([]*Post, 0, postsLen)
	s.PostsFeatured = make([]*Post, 0, postsLen)
	s.PostsByCategory = make(map[Category][]*Post, postsLen)

	// [1/5] Separate the posts into blog posts and standalone posts (pages).
	for slug := range s.PostsBySlug {
		post := s.PostsBySlug[slug]

		if post.IsBlog {
			s.Posts = append(s.Posts, post)
		} else {
			s.Pages = append(s.Pages, post)
		}
	}

	// [2/5] Sort the posts by pubdate timestamp.
	sort.SliceStable(s.Posts, func(i, j int) bool {
		return s.Posts[i].PubdateTS > s.Posts[j].PubdateTS
	})

	// [3/5] Prepare the posts by category.
	for _, post := range s.Posts {
		if !post.Category.IsEmpty() {

			cat := post.Category

			// Create slice if it does not exist.
			if len(s.PostsByCategory[cat]) == 0 {
				s.PostsByCategory[cat] = make([]*Post, 0, postsLen)
			}

			// Add post to the category slice.
			s.PostsByCategory[cat] = append(s.PostsByCategory[cat], post)
		}
	}

	// [4/5] Prepare the latest posts.
	s.PostsLatest = s.Posts[:min(LimitLatest, len(s.Posts))]

	// [5/5] Prepare the featured posts.
	for i := range s.Posts {
		post := s.Posts[i]

		if post.IsFeatured {
			s.PostsFeatured = append(s.PostsFeatured, post)
		}
	}

	s.PostsFeatured = s.PostsFeatured[:min(LimitFeatured, len(s.PostsFeatured))]
}

func (s *Store) RelatedPosts(post *Post) []*Post {
	// If the post is a standalone post, then return nil.
	if post.IsPage {
		return nil
	}

	// Retrieve the posts for the category.
	cPosts, ok := s.PostsByCategory[post.Category]

	// If there are no posts for the category,
	// then return nil.
	if !ok {
		return nil
	}

	// If there are not enough posts in the category to gather,
	// then return nil. This way, the number of category posts
	// must cross the `LimitRelated` threshold before they start
	// being recommended.
	if len(cPosts) <= LimitRelated {
		return nil
	}

	// Find the index of the given post.
	idx := 0
	for i, p := range cPosts {
		if p.Slug == post.Slug {
			idx = i
			break
		}
	}

	// Starting from the index of the current post,
	// gather the next x number of posts for the category.
	// Use modulo calculation to ensure the selection wraps
	// around the category posts.
	related := make([]*Post, 0, LimitRelated)
	for i := 0; i < LimitRelated; i++ {
		related = append(related, cPosts[(idx+i+1)%len(cPosts)])
	}

	return related
}

// ------------------------------------------------------------------
//
//
// Type: Processor
//
//
// ------------------------------------------------------------------

// Processor is responsible for parsing markdown posts and
// storing them as in-memory structs.
//
// Main method: `Process()`
// .
type Processor struct {
	// The markdown parser
	Markdown goldmark.Markdown

	// A slice of markdown posts filepaths to process
	FilePaths []string

	// The prepared Posts and Categories
	PostsBySlug      map[string]*Post
	CategoriesBySlug map[string]Category
}

func NewProcessor(filePaths []string) *Processor {
	return &Processor{
		//
		Markdown: goldmark.New(
			goldmark.WithExtensions(
				high.NewHighlighting(high.WithStyle("tango")),
				&frontmatter.Extender{
					Mode: frontmatter.SetMetadata,
				},
				extension.Table,
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
				html.WithHardWraps(),
			),
		),
		//
		FilePaths: filePaths,
		//
		PostsBySlug: make(map[string]*Post),
		//
		CategoriesBySlug: make(map[string]Category),
	}
}

// The Process method parses all markdown posts and
// stores it in memory.
// .
func (pr *Processor) Process() error {

	for _, filePath := range pr.FilePaths {

		// Read the markdown file.
		fileBytes, err := ReadFile(filePath)
		if err != nil {
			return err
		}

		// Construct Post item and add it to the database.
		if err := pr.processPost(fileBytes); err != nil {
			return err
		}
	}

	return nil
}

// processPost constructs a Post and optional Category
// from the given markdown file bytes.
// .
func (pr *Processor) processPost(mkdownBytes []byte) error {
	// Parse the file contents.
	doc := pr.Markdown.Parser().Parse(text.NewReader(mkdownBytes))

	m := MetadataParser{}

	// Get the document metadata, and construct a metadata parser.
	if metadata := doc.OwnerDocument().Meta(); metadata != nil {

		m.label = "unknown post"

		// Use the slug as the label.
		// This is useful for error messages.
		if slug, ok := metadata["slug"]; ok {
			m.label = slug.(string)
		}

		m.metadata = metadata
	}

	// Skip processing if the document is marked as draft.
	if isDraft := m.GetBool("draft", false); isDraft {
		return nil
	}

	// Parse title from metadata --------------------------
	title, err := m.GetRequiredString("title")
	if err != nil {
		return err
	}

	// Parse heading from metadata ------------------------
	heading, err := m.GetRequiredString("heading")
	if err != nil {
		return err
	}

	// Parse slug from metadata ---------------------------
	slug, err := m.GetRequiredString("slug")
	if err != nil {
		return err
	}

	// Parse description from metadata --------------------
	description, err := m.GetRequiredString("description")
	if err != nil {
		return err
	}

	// Parse featured from metadata -----------------------
	isFeatured := m.GetBool("featured", false)

	// Parse page from metadata ---------------------------
	isPage := m.GetBool("page", false)
	isBlog := !isPage

	// This check ensures that post slugs remain unique by guarding
	// against slug collision.
	if _, exists := pr.PostsBySlug[slug]; exists {
		return fmt.Errorf("post collision [%s]\n", slug)
	}

	// Parse pubdate from metadata ------------------------
	pubdate := ""
	pubdateTs := 0

	updated := ""
	updatedTs := 0

	if isBlog {
		pubdate, pubdateTs, err = m.GetRequiredDate("pubdate")
		if err != nil {
			return err
		}

		updated, updatedTs, err = m.GetDate("updated")
		if err != nil {
			return err
		}
	}

	// Parse category from metadata -----------------------
	cat := Category{}

	if isBlog {
		catTitle, err := m.GetRequiredString("category")
		if err != nil {
			return err
		}

		catSlug := Slugify(catTitle)

		existingCat, ok := pr.CategoriesBySlug[catSlug]
		// Handle the case where the category exists.
		if ok {
			// If multiple categories differ in case (ex 'Gardening Tips' and 'GarDENing TIPS'),
			// then they produce categories with the SAME SLUG, which is a problem.
			// This check ensures that category slugs remain unique by guarding
			// against category collision.
			if catTitle != existingCat.Title {
				return fmt.Errorf("category collision [%s] and [%s]\n", catTitle, existingCat.Title)
			}

			// Assign the existing category to `cat`, so that
			// it can be referenced in the post construction.
			cat = existingCat
		}

		// Handle the case where the category does NOT exist.
		if !ok {
			cat.Title = catTitle
			cat.Slug = catSlug

			// Save the category.
			pr.CategoriesBySlug[catSlug] = cat
		}
	}

	// Parse hide_image from metadata ---------------------
	showLead := !m.GetBool("hide_image", false)

	// Parse image from metadata --------------------------
	img := Image{}

	if isBlog {
		img.Url, err = m.GetRequiredString("image_url")
		if err != nil {
			return err
		}

		img.Alt, err = m.GetRequiredString("image_alt")
		if err != nil {
			return err
		}

		img.Type = ImageType
		img.Width = ImageWidth
		img.Height = ImageHeight
	}

	// Render the markdown content to a buffer.
	buf := new(bytes.Buffer)
	if err := pr.Markdown.Renderer().Render(buf, mkdownBytes, doc); err != nil {
		return fmt.Errorf("error when rendering body: %w", err)
	}

	// Save the post.
	pr.PostsBySlug[slug] = &Post{
		IsPage:      isPage,
		IsBlog:      isBlog,
		IsFeatured:  isFeatured,
		ShowLead:    showLead,
		Slug:        slug,
		Title:       title,
		Heading:     heading,
		Description: description,
		Category:    cat,
		Image:       img,
		Body:        buf.String(),
		Pubdate:     pubdate,
		PubdateTS:   pubdateTs,
		Updated:     updated,
		UpdatedTS:   updatedTs,
	}

	return nil
}

// ------------------------------------------------------------------
//
//
// Type: MetadataParser
//
//
// ------------------------------------------------------------------

// MetadataParser helps to parse markdown metadata.
// .
type MetadataParser struct {
	label    string
	metadata map[string]interface{}
}

// GetBool retrieves and converts a metadata value into a boolean.
// .
func (m *MetadataParser) GetBool(key string, defaultVal bool) bool {
	if !m.exists(key) {
		return defaultVal
	}
	return m.metadata[key].(bool)
}

// GetString retrieves and converts a metadata value into a string.
// .
func (m *MetadataParser) GetString(key string, defaultVal string) string {
	if !m.exists(key) {
		return defaultVal
	}
	return m.metadata[key].(string)
}

// GetRequiredString retrieves and converts a metadata value into a string.
// If not found, then an error is returned.
// .
func (m *MetadataParser) GetRequiredString(key string) (string, error) {
	if !m.exists(key) {
		return "", fmt.Errorf("%s is required [%s]", key, m.label)
	}
	return m.metadata[key].(string), nil
}

// GetDate retrieves and converts a metadata value into
// a time-formatted string and a unix timestamp.
// .
func (m *MetadataParser) GetDate(key string) (string, int, error) {
	if !m.exists(key) {
		return "", 0, nil
	}
	return m.parseDate(key)
}

// GetRequiredDate retrieves and converts a metadata value into
// a time-formatted string and a unix timestamp.
// If not found, then an error is returned.
// .
func (m *MetadataParser) GetRequiredDate(key string) (string, int, error) {
	if !m.exists(key) {
		return "", 0, fmt.Errorf("%s is required [%s]", key, m.label)
	}
	return m.parseDate(key)
}

func (m *MetadataParser) parseDate(key string) (string, int, error) {
	d := m.metadata[key].(time.Time)
	return d.Format("January 2, 2006"), int(d.Unix()), nil
}

func (m *MetadataParser) exists(key string) bool {
	_, ok := m.metadata[key]
	return ok
}

// ------------------------------------------------------------------
//
//
// Type: Exporter
//
//
// ------------------------------------------------------------------

// Exporter is responsible for exporting the server as a static site.
//
// Main method: `Export()`
// .
type Exporter struct {
	// The http handler responsible for rendering the urls.
	Handler http.Handler

	// The set of URLs to export.
	Urls []string

	// The directory where the site will be exported to.
	OutputPath string
}

// Export exports the configured server routes into a static site.
// .
func (e *Exporter) Export() error {
	// Remove the output path, if it exists.
	if err := os.RemoveAll(e.OutputPath); err != nil {
		return err
	}

	// Create the output directory.
	if err := EnsurePath(e.OutputPath); err != nil {
		return err
	}

	// Write the timestamp file.
	tsPath := filepath.Join(e.OutputPath, ".gingersnap")
	ts := time.Now().Format(time.UnixDate)

	if err := WriteFile(tsPath, []byte(ts)); err != nil {
		return err
	}

	// Render all the paths.
	for _, url := range e.Urls {
		if err := e.exportPage(url, e.makePath(url)); err != nil {
			return err
		}
	}

	return nil
}

// exportPage exports a single page to the filesystem
// by using the httptest framework to render the page.
// .
func (e *Exporter) exportPage(url, dstPath string) error {
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	e.Handler.ServeHTTP(w, r)

	if c := w.Result().StatusCode; c != http.StatusOK {
		return fmt.Errorf("expected URL %s to return %d, but it returned %d instead", url, http.StatusOK, c)
	}

	// Create the destination directory.
	if err := EnsurePath(dstPath); err != nil {
		return err
	}

	// Write the contents of the response body.
	if err := WriteFile(dstPath, w.Body.Bytes()); err != nil {
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
func (e *Exporter) makePath(url string) string {
	if url == "/CNAME" {
		return filepath.Join(e.OutputPath, "/CNAME")
	}

	if url == "/404/" {
		return filepath.Join(e.OutputPath, "404.html")
	}

	p := filepath.Join(e.OutputPath, url)

	if ext := filepath.Ext(p); ext == "" {
		p = filepath.Join(p, "index.html")
	}

	return p
}

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

// ------------------------------------------------------------------
//
//
// External config and utilities
//
//
// ------------------------------------------------------------------

// Settings stores external settings for gingersnap resources.
// These settings are used to configure the engine differently for
// DEBUG and for PROD.
// .
type Settings struct {
	// Debug mode
	Debug bool

	// The path to the json config file
	ConfigPath string

	// The glob pattern for the markdown posts
	PostsGlob string

	// The directory for media resources
	MediaDir string
}

// SafeDir returns a filepath directory.
// If the given path is a file, then the parent directory of the file will be returned.
// If the given path is a directory, then the directory itself will be returned.
// .
func (s Settings) SafeDir(p string) string {
	if ext := filepath.Ext(p); ext != "" {
		return filepath.Dir(p)
	}
	return p
}

// Configure wipes and reconfigures the gingersnap engine.
// .
func (g *Gingersnap) Configure(s Settings) {

	// ------------------------------------------
	//
	// [1/3] Wipe the gingersnap engine.
	//
	// ------------------------------------------

	if g.HttpServer != nil {
		g.HttpServer.Close()
		g.HttpServer = nil
	}
	g.Logger = nil
	g.Templates = nil
	g.Config = nil
	g.Store = nil

	// ------------------------------------------
	//
	// [2/3] Configure the engine components.
	//
	// ------------------------------------------

	// Construct the logger.
	logger := NewLogger()

	// Construct the config
	configBytes, err := ReadFile(s.ConfigPath)
	if err != nil {
		logger.Fatal(err)
	}

	config, err := NewConfig(configBytes, s.Debug)
	if err != nil {
		logger.Fatal(err)
	}

	// Gather the markdown post files.
	filePaths, err := filepath.Glob(s.PostsGlob)
	if err != nil {
		logger.Fatal(err)
	}

	// Parse the markdown posts.
	pr := NewProcessor(filePaths)
	if err := pr.Process(); err != nil {
		logger.Fatal(err)
	}

	// Construct the store from the processed markdown posts.
	store := NewStore()
	store.InitPosts(pr.PostsBySlug)
	store.InitCategories(pr.CategoriesBySlug)

	// Construct the templates, using the embedded FS.
	templates, err := NewTemplate(Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// ------------------------------------------
	//
	// [3/3] Construct the gingersnap engine.
	//
	// ------------------------------------------

	g.Logger = logger
	g.Assets = Assets
	g.Media = http.Dir(s.MediaDir)
	g.Templates = templates
	g.Config = config
	g.Store = store
	g.HttpServer = &http.Server{
		Addr:         g.Config.ListenAddr,
		Handler:      g.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// RunServerWithWatcher runs the server and and watches for file changes.
// On file change, it resets the gingersnap engine and restarts the server.
// .
func (g *Gingersnap) RunServerWithWatcher(s Settings) {
	// Create new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	if err = w.Add(s.SafeDir(s.ConfigPath)); err != nil {
		log.Fatal(err)
	}

	if err = w.Add(s.SafeDir(s.PostsGlob)); err != nil {
		log.Fatal(err)
	}

	if err = w.Add(s.SafeDir(s.MediaDir)); err != nil {
		log.Fatal(err)
	}

	g.Logger.Printf("Watching for file changes")

	go g.RunServer()

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				g.Logger.Println("Files changed. Restarting server")

				g.Configure(s)
				go g.RunServer()
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			g.Logger.Println("error:", err)
		}
	}
}

// RunServer runs the gingersnap server.
// .
func (g *Gingersnap) RunServer() {
	g.Logger.Printf("Starting server on %s 🤖\n\n", g.Config.ListenAddr)

	_ = g.HttpServer.ListenAndServe()
}

// Export exports the server as a static site.
// .
func (g *Gingersnap) Export() error {

	urls, err := g.AllUrls()
	if err != nil {
		return err
	}

	exporter := &Exporter{
		Handler:    g.Routes(),
		Urls:       urls,
		OutputPath: ExportDir,
	}

	return exporter.Export()
}

// ------------------------------------------------------------------
//
//
// Constants and embedded assets
//
//
// ------------------------------------------------------------------

//go:embed "assets"
var Assets embed.FS

//go:embed "assets/templates"
var Templates embed.FS

// Image settings for all images in the site.
const ImageWidth = "800"
const ImageHeight = "450"
const ImageType = "webp"

// Cutoff values for different post lists.
const LimitFeatured = 3
const LimitRelated = 3
const LimitLatest = 9
const LimitSection = 6
const LimitLatestPostDetail = 4

// A homepage section which represents all latest posts.
const SectionLatest = "$latest"

// A homepage section which represents all featured posts.
const SectionFeatured = "$featured"

// The directory where the static site will be exported to.
// This is only a temporary directory. To fully deploy the site,
// the exported site must be moved to the production site repository.
const ExportDir = "dist"

// THe default color theme.
var DefaultTheme = Theme{
	Primary:   "#4338ca", // indigo-700
	Secondary: "#0f172a", // slate-900
	Link:      "#1d4ed8", // blue-700
}

// Color themes for the site.
var Themes = map[string]Theme{
	"purple": {
		Primary:   "#4f46e5", // indigo-600
		Secondary: "#4338ca", // indigo-700
		Link:      "#2563eb", // blue-600
	},
	"green": {
		Primary:   "#0f766e", // teal-700
		Secondary: "#0f766e", // teal-700
		Link:      "#2563eb", // blue-600
	},
	"pink": {
		Primary:   "#db2777", // pink-600
		Secondary: "#be185d", // pink-700
		Link:      "#4f46e5", // indigo-600
	},
	"blue": {
		Primary:   "#0284c7", // sky-600
		Secondary: "#0284c7", // sky-600
		Link:      "#2563eb", // blue-600
	},
	"red": {
		Primary:   "#b91c1c", // red-700
		Secondary: "#be123c", // rose-700
		Link:      "#4f46e5", // indigo-600
	},
	"black": {
		Primary:   "#0f172a", // slate-900
		Secondary: "#0f172a", // slate-900
		Link:      "#2563eb", // blue-600
	},
}
