package gingersnap

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	htmlTemplate "html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
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
	Logger    *log.Logger
	Assets    embed.FS
	Templates *htmlTemplate.Template

	Config     *Config
	Posts      *PostModel
	Categories *CategoryModel
	Media      http.FileSystem

	HttpServer *http.Server
}

// New returns an empty Gingernap engine.
// .
func New() *Gingersnap {
	return &Gingersnap{}
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
	r.Handle("/sitemap.xml", g.HandleSitemap())
	r.Handle("/robots.txt", g.HandleRobotsTxt())
	r.Handle("/CNAME", g.HandleCname())
	r.Handle("/404/", g.Handle404())
	r.Handle("/media/", g.CacheControl(http.StripPrefix("/media", http.FileServer(g.Media))))

	// Build category routes
	for _, cat := range g.Categories.All() {
		r.Handle(cat.Route(), g.HandleCategory(cat))
	}

	// Build post routes
	for _, post := range g.Posts.All() {
		r.Handle(post.Route(), g.HandlePost(post))
	}

	return g.RecoverPanic(g.LogRequest(g.SecureHeaders(r)))
}

func (g *Gingersnap) AllUrls() ([]string, error) {
	urls := []string{"/", "/styles.css", "/sitemap.xml", "/robots.txt", "/CNAME", "/404/"}

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
		name := file.Name()

		// If file is not a dot file, then build a route for it.
		if !strings.HasPrefix(name, ".") {
			urls = append(urls, fmt.Sprintf("/media/%s", name))
		}
	}

	// Build routes for all posts.
	for _, post := range g.Posts.All() {
		urls = append(urls, post.Route())
	}

	// Build routes for all categories.
	for _, cat := range g.Categories.All() {
		urls = append(urls, cat.Route())
	}

	return urls, nil
}

func (g *Gingersnap) HandleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Handle 404
		if r.URL.Path != "/" {
			g.ErrNotFound(w)
			return
		}

		rd := g.NewRenderData(r)
		rd.FeaturedPosts = g.Posts.Featured()
		rd.LatestPosts = g.Posts.Latest()

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
		rd.LatestPosts = g.Posts.Latest()
		rd.FeaturedPosts = g.Posts.Featured()

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
		posts, ok := g.Posts.ByCategory(cat)
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

func (g *Gingersnap) HandleSitemap() http.HandlerFunc {

	// Prepare the sitemap template.
	tmpl, err := textTemplate.New("").Parse(strings.TrimPrefix(SitemapTemplate, "\n"))
	if err != nil {
		panic(err)
	}

	// The urlSet is a map of urls to lastmod dates.
	// It is used to render the sitemap.
	urlSet := make(map[string]string)

	// permalink is a helper function which generates
	// the permalink for a given path
	permalink := func(urlPath string) string {
		return fmt.Sprintf("%v%v", g.Config.Site.Url, urlPath)
	}

	// Add sitemap entries for the index page.
	urlSet[permalink("/")] = ""

	// Add sitemap entries for all the categories.
	for _, cat := range g.Categories.All() {
		urlSet[permalink(cat.Route())] = ""
	}

	// Add sitemap entries for all the posts.
	for _, post := range g.Posts.All() {
		lastMod := ""

		if ts := post.LatestTS(); ts > 0 {
			lastMod = time.Unix(int64(ts), 0).Format("2006-01-02T00:00:00+00:00")
		}

		urlSet[permalink(post.Route())] = lastMod
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
	rd.LatestPosts = g.Posts.Latest()

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
	rd.LatestPosts = g.Posts.Latest()

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
	return log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
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
	Debug       bool
	ListenAddr  string
	Site        Site   `json:"site"`
	NavbarLinks []Link `json:"navbarLinks"`
	FooterLinks []Link `json:"footerLinks"`
}

// Site stores site-specific settings
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
}

// Link stores data for an anchor link
// .
type Link struct {
	Text  string
	Route string
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

	return config, nil
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

// Post represents a blog or an article.
// .
type Post struct {
	// An FNV-1a hash of the slug
	Hash int

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

	// Is the post featured on the homepage?
	Featured bool

	// The publish date - January 2, 2006
	Pubdate string

	// The publish date, as a UNIX timestamp
	PubdateTS int

	// The updated date - January 2, 2006
	Updated string

	// The updated date, as a UNIX timestamp
	UpdatedTS int
}

// IsStandalone reports if the Post should be rendered
// as a standalone page, or as a blog post.
// .
func (p *Post) IsStandalone() bool {
	return p.Image.IsEmpty() || p.Category.IsEmpty()
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
func (c *Category) IsEmpty() bool {
	return c.Slug == ""
}

// Route returns the url path for the Post.
//
// ex: "/category/some-slug/"
// .
func (c *Category) Route() string {
	return fmt.Sprintf("/category/%s/", c.Slug)
}

// ------------------------------------------------------------------
//
//
// Type: PostModel
//
//
// ------------------------------------------------------------------

// PostModel manages queries for Posts.
// .
type PostModel struct {
	posts           []Post
	postsLatest     []Post
	postsFeatured   []Post
	postsBySlug     map[string]Post
	postsByCategory map[Category][]Post
}

func NewPostModel(postsBySlug map[string]Post) *PostModel {
	m := &PostModel{
		posts:           []Post{},
		postsLatest:     []Post{},
		postsFeatured:   []Post{},
		postsBySlug:     postsBySlug,
		postsByCategory: make(map[Category][]Post),
	}

	// Prepare the Post structures.
	for _, post := range m.postsBySlug {
		m.posts = append(m.posts, post)

		if !post.Category.IsEmpty() {
			cat := post.Category
			m.postsByCategory[cat] = append(m.postsByCategory[cat], post)
		}
	}

	// Sort the posts by latest timestamp.
	sort.SliceStable(m.posts, func(i, j int) bool {
		return m.posts[i].PubdateTS > m.posts[j].PubdateTS
	})

	// Prepare the latest posts.
	for i, post := range m.posts {
		if i == PostLatestLimit {
			break
		}
		m.postsLatest = append(m.postsLatest, post)
	}

	// Prepare the featured posts.
	for _, post := range m.posts {
		if len(m.postsFeatured) == PostFeaturedLimit {
			break
		}

		if post.Featured {
			m.postsFeatured = append(m.postsFeatured, post)
		}
	}

	return m
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
	return m.postsLatest
}

func (m *PostModel) Featured() []Post {
	return m.postsFeatured
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

// CategoryModel manages queries for Categories.
// .
type CategoryModel struct {
	categories       []Category
	categoriesBySlug map[string]Category
}

func NewCategoryModel(categoriesBySlug map[string]Category) *CategoryModel {
	m := &CategoryModel{
		categories:       []Category{},
		categoriesBySlug: categoriesBySlug,
	}

	for _, category := range m.categoriesBySlug {
		m.categories = append(m.categories, category)
	}

	return m
}

func (m *CategoryModel) All() []Category {
	return m.categories
}

func (m *CategoryModel) BySlug(s string) (Category, bool) {
	category, ok := m.categoriesBySlug[s]
	return category, ok
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
	PostsBySlug      map[string]Post
	CategoriesBySlug map[string]Category
}

func NewProcessor(filePaths []string) *Processor {
	return &Processor{
		//
		Markdown: goldmark.New(
			goldmark.WithExtensions(
				meta.New(meta.WithStoresInDocument()),
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
		PostsBySlug: make(map[string]Post),
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

// processPost constructs a Post and optional Category struct
// from the given markdown file bytes.
// .
func (pr *Processor) processPost(mkdownBytes []byte) error {
	// Parse the file contents.
	doc := pr.Markdown.Parser().Parse(text.NewReader(mkdownBytes))

	// Get the document metadata.
	metadata := doc.OwnerDocument().Meta()

	// Render the markdown content to a buffer.
	buf := new(bytes.Buffer)
	if err := pr.Markdown.Renderer().Render(buf, mkdownBytes, doc); err != nil {
		return err
	}

	// Retrieve fields from the markdown metadata.
	title := metadata["title"].(string)
	heading := metadata["heading"].(string)
	slug := metadata["slug"].(string)
	description := metadata["description"].(string)

	// Skip processing if the document is marked as draft.
	draft := false
	if _, ok := metadata["draft"]; ok {
		draft = metadata["draft"].(bool)
		if draft {
			fmt.Printf("skipping draft: %s\n", title)
			return nil
		}
	}

	// This check ensures that post slugs remain unique by guarding
	// against slug collision.
	if _, exists := pr.PostsBySlug[slug]; exists {
		return fmt.Errorf("post collision [%s]\n", slug)
	}

	// Retrieve the featured field.
	featured := false
	if _, ok := metadata["featured"]; ok {
		featured = metadata["featured"].(bool)
	}

	// Retrieve the pubdate field.
	pubdate := ""
	pubdateTs := 0
	// check that the pubdate value is a valid date.
	pd, err := time.Parse(time.DateOnly, metadata["pubdate"].(string))

	if err != nil {
		return fmt.Errorf("failed to parse pubdate %w", err)
	} else {
		pubdate = pd.Format("January 2, 2006")
		pubdateTs = int(pd.Unix())
	}

	// Retrieve the updated field
	updated := ""
	updatedTs := 0
	if _, ok := metadata["updated"]; ok {
		// check that the updated value is a valid date.
		ud, err := time.Parse(time.DateOnly, metadata["updated"].(string))

		if err != nil {
			return fmt.Errorf("failed to parse pubdate %w", err)
		} else {
			updated = ud.Format("January 2, 2006")
			updatedTs = int(ud.Unix())
		}
	}

	// Retrieve the category field.
	category := Category{}
	if _, ok := metadata["category"]; ok {
		categoryTitle := metadata["category"].(string)
		categorySlug := Slugify(categoryTitle)

		// If multiple categories differ in case (ex 'Gardening Tips' and 'GarDENing TIPS'),
		// then they produce unique categories with the SAME SLUG.
		// This check ensures that category slugs remain unique by guarding
		// against category collision.
		if ct, exists := pr.CategoriesBySlug[categorySlug]; exists {
			if ct.Title != categoryTitle {
				return fmt.Errorf("category collision [%s] and [%s]\n", ct.Title, categoryTitle)
			}
		}

		category = Category{
			Title: categoryTitle,
			Slug:  categorySlug,
		}

		// Save the category.
		//
		// At this point, overwriting the category does not produce
		// any negative effect because it is effectively the same.
		pr.CategoriesBySlug[categorySlug] = category
	}

	// Retrieve the lead image fields.
	image := Image{}
	if _, ok := metadata["image_url"]; ok {
		image.Url = metadata["image_url"].(string)
		image.Alt = metadata["image_alt"].(string)
		image.Type = ImageType
		image.Width = ImageWidth
		image.Height = ImageHeight
	}

	// Construct a Post object.
	post := Post{
		Hash:        HashSimple(slug),
		Slug:        slug,
		Title:       title,
		Heading:     heading,
		Description: description,
		Category:    category,
		Image:       image,
		Body:        buf.String(),
		Featured:    featured,
		Pubdate:     pubdate,
		PubdateTS:   pubdateTs,
		Updated:     updated,
		UpdatedTS:   updatedTs,
	}

	pr.PostsBySlug[slug] = post

	return nil
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

// GetEnv retrieves an variable from the application environment.
// If the variable is not found, the fallback is returned instead.
// .
func GetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

// HashSimple converts a string into a FNV-1a hash.
// .
func HashSimple(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

// Slugify builds a slug from the given string.
// .
func Slugify(s string) string {
	return strings.TrimSpace(strings.ToLower(strings.ReplaceAll(s, " ", "-")))
}

// Path builds filepaths from the project root.
// .
func Path(p string) string {
	return fmt.Sprintf("%s/%s", packageRoot(), p)
}

func projectRoot() string {
	return path.Dir(packageRoot())
}

func packageRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return path.Dir(b)
}

func OpenFile(fs fs.FS, source string) (fs.File, error) {
	// Open source file.
	f, err := fs.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open source file: %w", err)
	}
	return f, nil
}

func ReadFile(src string) ([]byte, error) {
	// Read the source file.
	srcBytes, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read source file: %w", err)
	}
	return srcBytes, nil
}

func EnsurePath(source string) error {
	// Create parent directories, if necessary.
	if err := os.MkdirAll(filepath.Dir(source), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create parent directories: %w", err)
	}
	return nil
}

func CreateFile(source string) (*os.File, error) {
	if err := EnsurePath(source); err != nil {
		return nil, err
	}

	// Create destination file.
	f, err := os.Create(source)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}

	return f, nil
}

func WriteFile(source string, data []byte) error {
	if err := EnsurePath(source); err != nil {
		return err
	}

	// Write the contents to the file.
	if err := os.WriteFile(source, data, os.FileMode(0644)); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

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
		return fmt.Errorf("failed to copy source file: %w", err)
	}

	return nil
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

const ImageWidth = "800"
const ImageHeight = "450"
const ImageType = "webp"

const PostFeaturedLimit = 4
const PostLatestLimit = 20

// ------------------------------------------------------------------
//
//
// External config and utilities
//
//
// ------------------------------------------------------------------

// Settings stores external settings for Gingersnap resources.
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
	g.Posts = nil
	g.Categories = nil

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
	processor := NewProcessor(filePaths)
	if err := processor.Process(); err != nil {
		logger.Fatal(err)
	}

	// Construct the models from the processed markdown posts.
	postModel := NewPostModel(processor.PostsBySlug)
	categoryModel := NewCategoryModel(processor.CategoriesBySlug)

	// Construct the templates, using the embedded FS.
	templates, err := NewTemplate(Templates)
	if err != nil {
		logger.Fatal(err)
	}

	// ------------------------------------------
	//
	// [3/3] Construct the main gingersnap engine.
	//
	// ------------------------------------------

	g.Logger = logger
	g.Assets = Assets
	g.Media = http.Dir(s.MediaDir)
	g.Templates = templates
	g.Config = config
	g.Posts = postModel
	g.Categories = categoryModel
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
				g.Logger.Println("file changes detected, restarting server...")

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
	g.Logger.Printf("Starting server on %s", g.Config.ListenAddr)

	if err := g.HttpServer.ListenAndServe(); err != nil {
		g.Logger.Print(err)
	}
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
		OutputPath: "dist",
	}

	return exporter.Export()
}
