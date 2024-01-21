package app

import (
	"fmt"
	"net/http"
	"time"
)

// ------------------------------------------------------------------
//
//
// renderData
//
//
// ------------------------------------------------------------------

// renderData stores the necessary data for template rendering.
// .
type renderData struct {
	// Site info
	SiteHost    string
	SiteUrl     string
	SiteName    string
	SiteTagline string
	SiteEmail   string
	PageUrl     string

	// The meta image / lead image
	Image image

	// The title tag, meta title tag and og:title tag
	Title string

	// The meta description
	Description string

	// The page heading, h1
	Heading string

	// The copyright year
	Copyright string

	// Post data
	Post          *post
	Posts         []*post
	LatestPosts   []*post
	RelatedPosts  []*post
	FeaturedPosts []*post

	// Category data
	Category   category
	Categories []category

	// Layout and styling
	Sections    []section
	NavbarLinks []siteLink
	FooterLinks []siteLink
	Theme       theme
	Display     display

	// Metrics
	AnalyticsTag string

	// Application info
	AppDebug bool
	AppError string
	AppTrace string
}

func (g *Gingersnap) newRenderData(r *http.Request) renderData {
	pageUrl := ""
	if r != nil {
		pageUrl = fmt.Sprintf("%s%s", g.config.Site.Url, r.URL.RequestURI())
	}

	return renderData{
		SiteHost:    g.config.Site.Host,
		SiteUrl:     g.config.Site.Url,
		SiteName:    g.config.Site.Name,
		SiteTagline: g.config.Site.Tagline,
		SiteEmail:   g.config.Site.Email,
		PageUrl:     pageUrl,

		Image: g.config.Site.Image,

		Title:       g.config.Site.Title,
		Description: g.config.Site.Description,
		Heading:     g.config.Site.Tagline,

		NavbarLinks: g.config.NavbarLinks,
		FooterLinks: g.config.FooterLinks,
		Theme:       g.config.Theme,
		Display:     g.config.Site.Display,

		AnalyticsTag: g.config.AnalyticsTag,

		Copyright: fmt.Sprintf("%d", time.Now().Year()),
		AppDebug:  g.config.Debug,
	}
}
