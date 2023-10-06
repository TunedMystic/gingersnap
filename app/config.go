package app

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ------------------------------------------------------------------
//
//
// Type: config
//
//
// ------------------------------------------------------------------

// config stores project settings
// .
type config struct {
	// Site-specific settings
	Site site `json:"site"`

	// Homepage sections
	Homepage []string `json:"homepage"`

	// Anchor links for the navbar
	NavbarLinks []siteLink `json:"navbarLinks"`

	// Anchor links for the footer
	FooterLinks []siteLink `json:"footerLinks"`

	// The git repository where the static site will be managed
	Repository string `json:"repository"`

	// Site styling
	Theme theme

	// If the program is running in DEBUG mode
	Debug bool

	// The address for the http.server to listen on
	ListenAddr string
}

// newConfig parses the settings and returns a *Config struct.
// .
func newConfig(configBytes []byte, debug bool) (*config, error) {

	c := &config{
		Debug:      debug,
		ListenAddr: ":4000",
	}

	// Parse the config file.
	if err := json.Unmarshal(configBytes, c); err != nil {
		return nil, err
	}

	c.Site.Url = fmt.Sprintf("https://%s", c.Site.Host)
	c.Site.Email = fmt.Sprintf("admin@%s", c.Site.Host)
	c.Site.Title = fmt.Sprintf("%s - %s", c.Site.Name, c.Site.Tagline)
	c.Site.Image = image{
		Url:    "/media/meta-img.webp",
		Alt:    c.Site.Title,
		Type:   imageType,
		Width:  imageWidth,
		Height: imageHeight,
	}

	// In "debug" mode, we change the host to localhost.
	if c.Debug {
		c.Site.Host = fmt.Sprintf("localhost%s", c.ListenAddr)
		c.Site.Url = fmt.Sprintf("http://%s", c.Site.Host)
		c.Site.Email = fmt.Sprintf("admin@%s", c.Site.Host)
	}

	// If no Homepage sections are defined, then create
	// a default setup with the "$latest" posts only.
	if c.Homepage == nil {
		c.Homepage = []string{sectionLatest}
	}

	// Retrieve the theme.
	theme, err := newTheme(c.Site.Theme)
	if err != nil {
		return nil, err
	}

	c.Theme = theme

	// Retrieve the display. Set appropriate defaults.
	if c.Site.Display == "" {
		c.Site.Display = display("grid")
	}

	if d := c.Site.Display; !d.isGrid() && !d.isList() {
		return nil, fmt.Errorf("could not load display [%s]", d)
	}

	return c, nil
}

// ------------------------------------------------------------------
//
//
// Type: site
//
//
// ------------------------------------------------------------------

// site stores site-specific settings.
// .
type site struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Tagline     string `json:"tagline"`
	Description string `json:"description"`
	Title       string
	Url         string
	Email       string
	Image       image
	Theme       string  `json:"theme"`
	Display     display `json:"display"`
}

// ------------------------------------------------------------------
//
//
// Type: siteLink
//
//
// ------------------------------------------------------------------

// siteLink stores data for an anchor link.
// .
type siteLink struct {
	Text string `json:"text"`
	Href string `json:"href"`
}

// ------------------------------------------------------------------
//
//
// Type: display
//
//
// ------------------------------------------------------------------

// display describes the method for rendering
// post sections across the website.
// .
type display string

// isGrid reports if the display is set to "grid"
// .
func (d display) isGrid() bool {
	return d == "grid"
}

// isList reports if the display is set to "list"
// .
func (d display) isList() bool {
	return d == "list"
}

// ------------------------------------------------------------------
//
//
// Type: theme
//
//
// ------------------------------------------------------------------

// theme represents a simple color profile for the site.
// .
type theme struct {
	Heading   string
	Primary   string
	Secondary string
	Link      string
}

// newTheme retrieves the requested theme.
// Simplified themes can also be requested. A simplified theme
// modifies the secondary color, which makes it less colorful.
//
// ex:
//
//	// will retrieve the pink theme
//	newTheme("pink")
//
//	// will retrieve the pink theme and modify it
//	newTheme("pink-simple")
//
// .
func newTheme(themeStr string) (theme, error) {

	// If no theme is given, then return the default.
	if themeStr == "" {
		return defaultTheme, nil
	}

	// Determine if a "simple" theme is requested.
	// With "simple" themes, the secondary color
	// is modified to make it less colorful.
	themeStr, isSimple := strings.CutSuffix(themeStr, "-simple")

	// Retrieve the theme.
	if t, ok := themes[themeStr]; ok {

		theme := theme{
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

	return theme{}, fmt.Errorf("Could not load theme [%s]", themeStr)
}

// ------------------------------------------------------------------
//
//
// Themes
//
//
// ------------------------------------------------------------------

// The default color theme.
var defaultTheme = theme{
	Primary:   "#4338ca", // indigo-700
	Secondary: "#0f172a", // slate-900
	Link:      "#1d4ed8", // blue-700
}

// Color themes for the site.
var themes = map[string]theme{
	"purple": {
		Primary:   "#4f46e5", // indigo-600
		Secondary: "#4338ca", // indigo-700
		Link:      "#2563eb", // blue-600
	},
	"green": {
		Primary:   "#0f766e", // teal-700
		Secondary: "#0f766e", // teal-700
		Link:      "#0369a1", // sky-700
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
