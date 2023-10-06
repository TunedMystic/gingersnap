package app

import (
	"bytes"
	"fmt"
	"time"

	"github.com/yuin/goldmark"
	high "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"

	"gingersnap/app/utils"
)

// ------------------------------------------------------------------
//
//
// Type: processor
//
//
// ------------------------------------------------------------------

// processor is responsible for parsing markdown posts and
// storing them as in-memory structs.
//
// Main method: `process()`
// .
type processor struct {
	// The markdown parser
	markdown goldmark.Markdown

	// A slice of markdown posts filepaths to process
	filePaths []string

	// The collected Posts and Categories
	postsBySlug      map[string]*post
	categoriesBySlug map[string]category
}

func newProcessor(filePaths []string) *processor {
	return &processor{
		//
		markdown: goldmark.New(
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
		filePaths: filePaths,
		//
		postsBySlug: make(map[string]*post, 20),
		//
		categoriesBySlug: make(map[string]category, 20),
	}
}

// The Process method parses all markdown posts and
// stores it in memory.
// .
func (pr *processor) process() error {

	for _, filePath := range pr.filePaths {

		// Read the markdown file.
		fileBytes, err := utils.ReadFile(filePath)
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
func (pr *processor) processPost(mkdownBytes []byte) error {
	// Parse the file contents.
	doc := pr.markdown.Parser().Parse(text.NewReader(mkdownBytes))

	m := metadataParser{}

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
	if isDraft := m.getBool("draft", false); isDraft {
		return nil
	}

	// Parse title from metadata --------------------------
	title, err := m.mustGetString("title")
	if err != nil {
		return err
	}

	// Parse heading from metadata ------------------------
	heading, err := m.mustGetString("heading")
	if err != nil {
		return err
	}

	// Parse slug from metadata ---------------------------
	slug, err := m.mustGetString("slug")
	if err != nil {
		return err
	}

	// Parse description from metadata --------------------
	description, err := m.mustGetString("description")
	if err != nil {
		return err
	}

	// Parse featured from metadata -----------------------
	isFeatured := m.getBool("featured", false)

	// Parse page from metadata ---------------------------
	isPage := m.getBool("page", false)
	isBlog := !isPage

	// This check ensures that post slugs remain unique by guarding
	// against slug collision.
	if _, exists := pr.postsBySlug[slug]; exists {
		return fmt.Errorf("post collision [%s]\n", slug)
	}

	// Parse pubdate from metadata ------------------------
	pubdate := ""
	pubdateTs := 0

	updated := ""
	updatedTs := 0

	if isBlog {
		pubdate, pubdateTs, err = m.mustGetDate("pubdate")
		if err != nil {
			return err
		}

		updated, updatedTs, err = m.getDate("updated")
		if err != nil {
			return err
		}
	}

	// Parse category from metadata -----------------------
	cat := category{}

	if isBlog {
		catTitle, err := m.mustGetString("category")
		if err != nil {
			return err
		}

		catSlug := utils.Slugify(catTitle)

		existingCat, ok := pr.categoriesBySlug[catSlug]
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
			pr.categoriesBySlug[catSlug] = cat
		}
	}

	// Parse hide_image from metadata ---------------------
	showLead := !m.getBool("hide_image", false)

	// Parse image from metadata --------------------------
	img := image{}

	if isBlog {
		img.Url, err = m.mustGetString("image_url")
		if err != nil {
			return err
		}

		img.Alt, err = m.mustGetString("image_alt")
		if err != nil {
			return err
		}

		img.Type = imageType
		img.Width = imageWidth
		img.Height = imageHeight
	}

	// Render the markdown content to a buffer.
	buf := new(bytes.Buffer)
	if err := pr.markdown.Renderer().Render(buf, mkdownBytes, doc); err != nil {
		return fmt.Errorf("error when rendering body: %w", err)
	}

	// Save the post.
	pr.postsBySlug[slug] = &post{
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
// Type: metadataParser
//
//
// ------------------------------------------------------------------

// metadataParser helps to parse markdown metadata.
// .
type metadataParser struct {
	label    string
	metadata map[string]interface{}
}

// getBool retrieves and converts a metadata value into a boolean.
// .
func (m *metadataParser) getBool(key string, defaultVal bool) bool {
	if !m.exists(key) {
		return defaultVal
	}
	return m.metadata[key].(bool)
}

// getString retrieves and converts a metadata value into a string.
// .
func (m *metadataParser) getString(key string, defaultVal string) string {
	if !m.exists(key) {
		return defaultVal
	}
	return m.metadata[key].(string)
}

// mustGetString retrieves and converts a metadata value into a string.
// If not found, then an error is returned.
// .
func (m *metadataParser) mustGetString(key string) (string, error) {
	if !m.exists(key) {
		return "", fmt.Errorf("%s is required [%s]", key, m.label)
	}
	return m.metadata[key].(string), nil
}

// getDate retrieves and converts a metadata value into
// a time-formatted string and a unix timestamp.
// .
func (m *metadataParser) getDate(key string) (string, int, error) {
	if !m.exists(key) {
		return "", 0, nil
	}
	return m.parseDate(key)
}

// mustGetDate retrieves and converts a metadata value into
// a time-formatted string and a unix timestamp.
// If not found, then an error is returned.
// .
func (m *metadataParser) mustGetDate(key string) (string, int, error) {
	if !m.exists(key) {
		return "", 0, fmt.Errorf("%s is required [%s]", key, m.label)
	}
	return m.parseDate(key)
}

func (m *metadataParser) parseDate(key string) (string, int, error) {
	d := m.metadata[key].(time.Time)
	return d.Format("January 2, 2006"), int(d.Unix()), nil
}

func (m *metadataParser) exists(key string) bool {
	_, ok := m.metadata[key]
	return ok
}
