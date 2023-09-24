---
title: Diving into SQLite Full-Text Search - The Complete Guide
heading: SQLite Full-Text Search
slug: sqlite-full-text-search
description: In this guide, you'll learn all about SQLite's Full-Text Search feature and how to use it to efficiently search and retrieve data from your database.

category: SQL
image_url: /media/full-text-search.webp
image_alt: SQLite Full-Text Search

pubdate: 2023-01-21
updated: 2023-01-23
featured: true
---

## Getting started

Gingersnap is a static site generator built in Go. It converts markdown files into a fully functional website.

With a single command, users can setup an entire project, allowing them to jump straight into editing and content creation.

Projects have a limited set of features and configuration. This helps keep the blogging workflow streamlined and ensures a more straightforward experience for users.

<details open>
    <summary>Table of Contents</summary>

- [Quickstart](#quickstart)
- [Project structure](#project-structure)
- [Config](#config)
- [Examples](#examples)
</details>

## Quickstart

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

---

## Project structure

A gingersnap project contains the following directory structure:

```plaintext
├─ posts/
├─ media/
├─ gingersnap.json
```

**Posts** - The `posts` directory contains all the site content as markdown files. Gingersnap uses the metadata within each markdown file to determine the post's title, slug, category etc.

**Media** - The `media` directory contains all media assets that are used in the posts. These assets can be images, audio, video, etc.

**Config** - The config file stores settings and layout configurations for the site. More details about the config file [below](#config).

---

## Config

The `gingersnap.json` config file stores settings and layout configurations for the site. The config is separated into the following 5 sections.


**Site** defines site-specific settings.

| | |
| ----------- | ----------- |
| `name` | The name of the site |
| `host` | The host of the site |
| `tagline` | A short description of the site _(50-70 characters)_ |
| `description` | A long description of the site _(70-155 characters)_ |
| `theme` | The color theme of the site _([see available themes](#themes))_ |

```json
"site": {
    "name": "MySite",
    "host": "mysite.com",
    "tagline": "short descr ...",
    "description": "longer descr ...",
    "theme": "pink"
}
```

<br />

**Homepage** defines sections for the homepage. This _(optional)_ setting requires a list of categories.

The `"$latest"` tag is a pseudo category that represents the latest posts across the site. Gingersnap will then build the homepage based on this given list.

```json
"homepage": ["category-slug", "$latest"]
```

<br />

**Navbar Links** defines anchor links for the top navbar. This _(optional)_ setting requires a list of anchor link objects.

| | |
| ----------- | ----------- |
| Text | The anchor link text |
| Route | The anchor link href |

```json
"navbarLinks": [
    {"text": "About Us", "route": "/about/"},
    {"text": "My Article", "route": "/abc/"}
]
```

<br />

**Footer Links** defines anchor links for the footer. This _(optional)_ setting requires a list of anchor link objects.

| | |
| ----------- | ----------- |
| Text | The anchor link text |
| Route | The anchor link href |

```json
"footerLinks": [
    {"text": "About Us", "route": "/about/"},
    {"text": "My Article", "route": "/abc/"}
]
```

<br />

**Static Repository** defines the export destination. This _(optional)_ setting requires a repository path where the site will be exported to.


```json
"staticRepository": "/path/to/static/repo"
```

---

## Examples

Here are some tips for working with Gingersnap projects.


#### Draft Posts

Any post marked as draft will never be processed or exported by Gingersnap.

You can set a post as draft by adding `draft: true` in the markdown front matter.


#### Featured Posts

Featured posts are pinned posts that Gingersnap displays in specific areas of the site, such as the post detail page.

You can set a post as featured by adding `featured: true` in the markdown front matter.


#### Standalone Posts

A standalone post has no relation to other content across the site. Examples of standalone posts are a "contact" page, an "about" page, or a "privacy-policy" page.

When Gingersnap displays a standalone post, it will not include related or latest posts on the page. A standalone post does not need the following front matter fields: `category`, `image_url`, `pubdate` or `updated`.

You can set a post as standalone by adding `page: true` to the markdown front matter.


#### Lead Image

Each post must contain a lead image. You can set the lead in the markdown front matter with the `image_url` and `image_alt` fields.

To keep things simple, all lead images must be in `webp` format and must have a resolution of `1280x720`.

Gingersnap will display the lead image in the post detail page. Alternatively, you can hide the lead image in the post detail page with `hide_image: true`.


#### Themes

Gingersnap comes with the following color themes, each with a primary _(left)_ and secondary _(right)_ color. The primary color is applied to the site header and the category links. The secondary color is applied to all heading tags, except `h1`.

| | |
| ----------- | ----------- |
| Purple | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#4f46e5;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#4338ca;"></span></div> |
| Green | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#0f766e;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f766e;"></span></div> |
| Pink | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#db2777;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#be185d;"></span></div> |
| Blue | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#0284c7;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0284c7;"></span></div> |
| Red | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#b91c1c;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#be123c;"></span></div> |
| Black | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#0f172a;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |


You can also simplify the themes by adding `"-simple"` to the theme name in the `gingersnap.json` file.

```json
// site settings
{
    "theme": "pink-simple"
}
```

When you specify a simple theme, Gingersnap will use the primary color of the root theme, and will use black as the secondary color.

| | |
| ----------- | ----------- |
| Purple | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#4f46e5;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |
| Green | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#0f766e;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |
| Pink | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#db2777;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |
| Blue | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#0284c7;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |
| Red | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#b91c1c;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |
| Black | <div style="display:flex;align-items:center;"><span style="border-radius:15px;width:20px;height:20px;background:#0f172a;"></span><span style="margin-left:10px;border-radius:15px;width:20px;height:20px;background:#0f172a;"></span></div> |


#### Media Within Posts

An image example
<img width="800" height="450" src="/media/food.webp" alt="Organic Food"/>

```html
<img width="800" height="450" src="/media/food.webp" alt="Organic Food"/>
```


<br />


A video example
<video controls src="/media/mario-kart.mp4" alt="Mario Kart"></video>

```html
<video controls src="/media/mario-kart.mp4" alt="Mario Kart"></video>
```


<br />


An audio example
<audio controls src="/media/drums.mp3" alt="Drums"></audio>

```html
<audio controls src="/media/drums.mp3" alt="Drums"></audio>
```


<br />


Have fun building sites with Gingersnap! ❤️
