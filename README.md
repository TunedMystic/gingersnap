![gingersnap](https://github.com/TunedMystic/gingersnap-cli/assets/6523726/32d70e4c-1818-485d-8ed5-ee1116c0e70a)

Gingersnap is a static site generator built in Go.



<br />



### How to build the Gingersnap cli

**Clone the repo**
```shell
git clone https://github.com/tunedmystic/gingersnap-cli && cd gingersnap-cli
```

**Build and install the app binary**
```shell
make install
```



<br />



## Getting started with Gingersnap

Gingersnap is a static site generator built in Go. It converts markdown files into a fully functional website.

With a single command, users can setup an entire project, allowing them to jump straight into editing and content creation.

Projects have a limited set of features and configuration. This helps keep the blogging workflow streamlined and ensures a more straightforward experience for users.

<br />

<details open>
    <summary><strong>Table of Contents</strong></summary>

- [Quickstart](#quickstart)
- [Project structure](#project-structure)
- [Config](#config)
- [Examples](#examples)
</details>



<br />



### Quickstart

First, create an empty directory and navigate to it.

```shell
mkdir mysite && cd mysite
```


Next, use gingersnap to initialize a project.
A starter project will be scaffolded, complete with blog posts and media assets.

```shell
gingersnap init
```


Then, use gingersnap to start a development server on `localhost:4000`.
You can add/edit content, and the server will restart to reflect the changes.

```shell
gingersnap dev
```


Finally, use gingersnap to export the project as a static site.
The site will be exported to the `dist/` directory.

```shell
gingersnap export
```


<br />


---



<br />



### Project structure

A gingersnap project contains the following directory structure:

```plaintext
├─ posts/
├─ media/
├─ gingersnap.json
```

**Posts** - The `posts` directory contains all the site content as markdown files. Gingersnap uses the metadata within each markdown file to determine the post's title, slug, category etc.

**Media** - The `media` directory contains all media assets that are used in the posts. These assets can be images, audio, video, etc.

**Config** - The config file stores settings and layout configurations for the site. More details about the config file [below](#config).



<br />



---



<br />



### Config

The `gingersnap.json` config file stores settings and layout configurations for the site. Here is an overview of the config options.

#### Site
Defines site-specific settings.

| | |
| ----------- | ----------- |
| `name` | The name of the site |
| `host` | The host of the site |
| `tagline` | A short description of the site _(50-70 characters)_ |
| `description` | A long description of the site _(70-155 characters)_ |
| `theme` | The color theme of the site _([see available themes](#themes))_ |
| `display` | Render the posts in "grid" or "list" style |

```json
"site": {
    "name": "MySite",
    "host": "mysite.com",
    "tagline": "short descr ...",
    "description": "longer descr ...",
    "theme": "pink",
    "display": "grid"
}
```

<br />

#### Homepage
Defines sections for the homepage. This _(optional)_ setting requires a list of categories.

The `"$latest"` tag is a custom section that represents all latest posts.
The `"$featured"` tag is a custom section which represents all featured posts.

Gingersnap will then build the homepage based on this given list.

```json
"homepage": ["category-slug", "$latest"]
```

<br />

#### Navbar Links
Defines anchor links for the top navbar. This _(optional)_ setting requires a list of anchor link objects.

| | |
| ----------- | ----------- |
| Text | The anchor link text |
| Href | The anchor link href |

```json
"navbarLinks": [
    {"text": "About Us", "href": "/about/"},
    {"text": "My Article", "href": "/abc/"}
]
```

<br />

#### Footer Links
Defines anchor links for the footer. This _(optional)_ setting requires a list of anchor link objects.

| | |
| ----------- | ----------- |
| Text | The anchor link text |
| Href | The anchor link href |

```json
"footerLinks": [
    {"text": "About Us", "href": "/about/"},
    {"text": "My Article", "href": "/abc/"}
]
```

<br />

#### Static Repository
Defines the export destination. This _(optional)_ setting requires a repository path where the site will be exported to.


```json
"staticRepository": "/path/to/static/repo"
```

<br />

### Example config

```json
{
    "site": {
        "name": "Gingersnap",
        "host": "gingersnap.dev",
        "tagline": "The Snappy Way to Build Static Sites.",
        "description": "Gingersnap is a simple and effortless static site generator. Get up and running with one command, and export the site when you're ready to publish!",
		"theme": "pink",
		"display": "grid"
    },
    "homepage": [
        "$featured",
        "go",
        "$latest"
    ],
    "navbarLinks": [
        {"text": "Go", "href": "/category/go/"},
        {"text": "Python", "href": "/category/python/"},
        {"text": "SQL", "href": "/category/sql/"},
        {"text": "About Us", "href": "/about/"}
    ],
    "footerLinks": [
        {"text": "Home", "href": "/"},
        {"text": "About Us", "href": "/about/"},
        {"text": "Sitemap", "href": "/sitemap/"},
        {"text": "Privacy Policy", "href": "/privacy-policy/"}
    ],
    "staticRepository": "/path/to/static/repo"
}
```

<br />

### Example display options

The `site.display` setting will tell Gingersnap to render Post blocks as a list or as a grid. Post blocks are used in the [homepage](https://gingersnap.dev/), in [category pages](https://gingersnap.dev/category/go/), and in [post detail pages](https://gingersnap.dev/golang-working-with-time/).

Here is an example of how these two displays are rendered.

![gingersnap-display-types](https://github.com/TunedMystic/gingersnap-cli/assets/6523726/60a68500-ce0b-41d0-924a-9d90cce6c3f1)



<br />



---



<br />



### Examples

Here are some tips for working with Gingersnap projects.


<br />


#### Draft Posts

Any post marked as draft will never be processed or exported by Gingersnap. You can make a post draft by adding `draft: true` in the markdown front matter.


<br />


#### Featured Posts

Featured posts are pinned posts that Gingersnap displays on the homepage.

You can make a post featured by adding `featured: true` in the markdown front matter.

You can display featured posts on the homepage by adding the `"$featured"` section to the [homepage settings](#homepage).


<br />


#### Standalone Posts

A standalone post has no relation to other content across the site. Examples of standalone posts are a "contact" page, an "about" page, or a "privacy-policy" page.

When Gingersnap displays a standalone post, it will not include related or latest posts on the page. A standalone post does not need the following front matter fields: `category`, `image_url`, `pubdate` or `updated`.

You can make a post standalone by adding `page: true` to the markdown front matter.


<br />


#### Lead Image

Each post must contain a lead image. You can set the lead in the markdown front matter with the `image_url` and `image_alt` fields.

To keep things simple, all lead images must be in `webp` format and must have a resolution of `1280x720`.

Gingersnap will display the lead image in the post detail page. Alternatively, you can hide the lead image in the post detail page with `hide_image: true`.


<br />


#### Themes

Gingersnap comes with the following color themes, each with a primary _(left)_ and secondary _(right)_ color. The primary color is applied to the site header and the category links. The secondary color is applied to all heading tags, except `h1`.

<img width="748" alt="gingersnap-themes" src="https://github.com/TunedMystic/gingersnap-cli/assets/6523726/4345ab30-9350-4f17-814e-c8270cf01ea9">


You can also simplify the themes by adding `"-simple"` to the theme name in the `gingersnap.json` file.

```json
{
    "theme": "pink-simple"
}
```

When you specify a simple theme, Gingersnap will use the primary color of the root theme, and will use black as the secondary color.

<img width="748" alt="gingersnap-themes-simplified" src="https://github.com/TunedMystic/gingersnap-cli/assets/6523726/de5d92e4-ed4c-4fe2-b1e7-4926736418f1">


<br />

And that's it! Have fun building sites with Gingersnap! 🍪 ❤️
