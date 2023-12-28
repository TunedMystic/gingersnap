package app

import "fmt"

// ------------------------------------------------------------------
//
//
// Type: post
//
//
// ------------------------------------------------------------------

// post represents an article or a web page.
// .
type post struct {
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
	Category category

	// The lead image
	Image image

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

	// The index of the post in the `PostsByCategory` map.
	idxCategory int
}

// LatestTS returns the Post's latest timestamped date.
// .
func (p *post) LatestTS() int {
	if p.UpdatedTS > p.PubdateTS {
		return p.UpdatedTS
	}
	return p.PubdateTS
}

// Route returns the url path for the Post.
//
// ex: "/post-slug/"
// .
func (p *post) Route() string {
	return fmt.Sprintf("/%s/", p.Slug)
}

// ------------------------------------------------------------------
//
//
// Type: category
//
//
// ------------------------------------------------------------------

// category represents a post category.
// .
type category struct {
	Slug  string
	Title string
}

// IsEmpty reports if the category is empty.
// .
func (c category) IsEmpty() bool {
	return c.Slug == ""
}

// Route returns the url path for the category.
//
// ex: "/category/some-slug/"
// .
func (c category) Route() string {
	return fmt.Sprintf("/category/%s/", c.Slug)
}

// ------------------------------------------------------------------
//
//
// Type: section
//
//
// ------------------------------------------------------------------

// section represents a section of a web page.
// It stores a collection of posts for a category.
//
// For non-conventional sections like "Latest Posts",
// the category can be a pseudo-category.
// .
type section struct {
	Category category
	Posts    []*post
}

// A homepage section which represents all latest posts.
const sectionLatest = "$latest"

// A homepage section which represents all featured posts.
const sectionFeatured = "$featured"

// A homepage section which represents all posts.
const sectionAll = "$all"

// ------------------------------------------------------------------
//
//
// Type: image
//
//
// ------------------------------------------------------------------

// image represents an image media object.
// .
type image struct {
	Url    string
	Alt    string
	Type   string
	Width  string
	Height string
}

// isEmpty reports if the image is empty.
// .
func (i *image) IsEmpty() bool {
	return i.Url == ""
}

// Image settings for all images in the site.
const imageWidth = "800"
const imageHeight = "450"
const imageType = "webp"
