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
	isPage bool

	// If the Post is a blog post
	isBlog bool

	// If the post is a featured post
	isFeatured bool

	// If the lead image should be displayed
	showLead bool

	// The post slug
	slug string

	// The post meta title
	title string

	// The post heading, or main title
	heading string

	// The post description
	description string

	// The post category
	category category

	// The lead image
	image image

	// The post body
	body string

	// The publish date - January 2, 2006
	pubdate string

	// The publish date, as a UNIX timestamp
	pubdateTS int

	// The updated date - January 2, 2006
	updated string

	// The updated date, as a UNIX timestamp
	updatedTS int

	// The index of the post in the `PostsByCategory` map.
	idxCategory int
}

// latestTS returns the Post's latest timestamped date.
// .
func (p *post) latestTS() int {
	if p.updatedTS > p.pubdateTS {
		return p.updatedTS
	}
	return p.pubdateTS
}

// route returns the url path for the Post.
//
// ex: "/post-slug/"
// .
func (p *post) route() string {
	return fmt.Sprintf("/%s/", p.slug)
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

// isEmpty reports if the category is empty.
// .
func (c category) isEmpty() bool {
	return c.Slug == ""
}

// route returns the url path for the category.
//
// ex: "/category/some-slug/"
// .
func (c category) route() string {
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
	category category
	posts    []*post
}

// A homepage section which represents all latest posts.
const sectionLatest = "$latest"

// A homepage section which represents all featured posts.
const sectionFeatured = "$featured"

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
func (i *image) isEmpty() bool {
	return i.Url == ""
}

// Image settings for all images in the site.
const imageWidth = "800"
const imageHeight = "450"
const imageType = "webp"
