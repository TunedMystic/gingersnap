package app

import "sort"

// Cutoff values for different post lists.
const limitFeatured = 3
const limitRelated = 6
const limitLatest = 9
const limitSection = 6
const limitLatestPostDetail = 4

// ------------------------------------------------------------------
//
//
// Type: store
//
//
// ------------------------------------------------------------------

// store is responsible for storing Posts and Categories
// in an organized way, making easy to access.
// .
type store struct {
	posts           []*post
	pages           []*post
	postsLatest     []*post
	postsLatestSm   []*post
	postsFeatured   []*post
	postsBySlug     map[string]*post
	postsByCategory map[category][]*post

	categories       []category
	categoriesBySlug map[string]category

	sections map[string]section
}

func newStore() *store {
	return &store{}
}

func (s *store) InitPosts(postsBySlug map[string]*post) {

	// Note: the `postsBySlug` map contains
	//       blog posts AND standalone posts (pages).

	s.postsBySlug = postsBySlug

	postsLen := len(s.postsBySlug)

	s.posts = make([]*post, 0, postsLen)
	s.pages = make([]*post, 0, postsLen)
	s.postsLatest = make([]*post, 0, postsLen)
	s.postsLatestSm = make([]*post, 0, postsLen)
	s.postsFeatured = make([]*post, 0, postsLen)
	s.postsByCategory = make(map[category][]*post, postsLen)

	// [1/5] Separate the posts into blog posts and standalone posts (pages).
	for slug := range s.postsBySlug {
		p := s.postsBySlug[slug]

		if p.IsBlog {
			s.posts = append(s.posts, p)
		} else {
			s.pages = append(s.pages, p)
		}
	}

	// [2/5] Sort the posts by pubdate timestamp.
	sort.SliceStable(s.posts, func(i, j int) bool {
		return s.posts[i].PubdateTS > s.posts[j].PubdateTS
	})

	// [3/5] Prepare the posts by category.
	for i := range s.posts {
		p := s.posts[i]

		// skip if the post has no category (standalone post)
		if p.Category.IsEmpty() {
			continue
		}

		cat := p.Category

		// Create slice if it does not exist.
		if len(s.postsByCategory[cat]) == 0 {
			s.postsByCategory[cat] = make([]*post, 0, postsLen)
		}

		// Add post to the category slice.
		s.postsByCategory[cat] = append(s.postsByCategory[cat], p)

		// Update the index of the post, relative to the category it is in.
		p.idxCategory = len(s.postsByCategory[cat]) - 1
	}

	// [4/5] Prepare the latest posts.
	s.postsLatest = s.posts[:min(limitLatest, len(s.posts))]
	s.postsLatestSm = s.postsLatest[:min(limitLatestPostDetail, len(s.postsLatest))]

	// [5/5] Prepare the featured posts.
	for i := range s.posts {
		p := s.posts[i]

		if p.IsFeatured {
			s.postsFeatured = append(s.postsFeatured, p)
		}
	}

	s.postsFeatured = s.postsFeatured[:min(limitFeatured, len(s.postsFeatured))]
}

func (s *store) InitCategories(categoriesBySlug map[string]category) {

	s.categoriesBySlug = categoriesBySlug

	s.categories = make([]category, 0, len(s.categoriesBySlug))

	for slug := range s.categoriesBySlug {
		cat := s.categoriesBySlug[slug]
		s.categories = append(s.categories, cat)
	}
}

func (s *store) InitSections() {
	s.sections = make(map[string]section, len(s.postsByCategory)+2)

	// Create sections for each "category-grouping" of posts.
	for cat := range s.postsByCategory {
		s.sections[cat.Slug] = section{
			Category: cat,
			Posts:    s.postsByCategory[cat],
		}
	}

	// Create section for the "Latest Posts" pseudo-category.
	s.sections[sectionLatest] = section{
		Category: category{
			Slug:  "",
			Title: "Latest Posts",
		},
		Posts: s.postsLatest,
	}

	// Create section for the "Featured Posts" pseudo-category.
	s.sections[sectionFeatured] = section{
		Category: category{
			Slug:  "",
			Title: "Featured Posts",
		},
		Posts: s.postsFeatured,
	}

	// Create section for the "All Posts" pseudo-category.
	s.sections[sectionAll] = section{
		Category: category{
			Slug:  "",
			Title: "",
		},
		Posts: s.posts,
	}
}

func (s *store) RelatedPosts(p *post) []*post {
	// If the post is a standalone post, then return nil.
	if p.IsPage {
		return nil
	}

	// Retrieve the posts for the category.
	cPosts, ok := s.postsByCategory[p.Category]

	// If there are no posts for the category,
	// then return nil.
	if !ok {
		return nil
	}

	// If there are not enough posts in the category to gather,
	// then return nil. This way, the number of category posts
	// must cross the `limitRelated` threshold before they start
	// being recommended.
	if len(cPosts) <= limitRelated {
		return nil
	}

	// Starting from the index of the current post,
	// gather the next x number of posts for the category.
	// Use modulo calculation to ensure the selection wraps
	// around the category posts.
	related := make([]*post, 0, limitRelated)
	for i := 0; i < limitRelated; i++ {
		related = append(related, cPosts[(p.idxCategory+i+1)%len(cPosts)])
	}

	return related
}
