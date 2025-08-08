package image

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractImageURLs parses HTML and returns all image URLs found, resolved to absolute URLs.
func ExtractImageURLs(html string, baseURL string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	base, _ := url.Parse(baseURL)
	found := map[string]struct{}{}
	var urls []string

	// 1. Extract from <a href=...> tags with image extensions (full-res images)
	imageExts := []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".bmp", ".tiff"}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			href = strings.TrimSpace(href)
			if href == "" {
				return
			}
			lower := strings.ToLower(href)
			isImage := false
			for _, ext := range imageExts {
				if strings.HasSuffix(lower, ext) {
					isImage = true
					break
				}
			}
			if isImage {
				// resolve relative URLs
				if !strings.HasPrefix(href, "http") && !strings.HasPrefix(href, "data:") {
					u, err := url.Parse(href)
					if err == nil {
						href = base.ResolveReference(u).String()
					}
				}
				if _, exists := found[href]; !exists {
					found[href] = struct{}{}
					urls = append(urls, href)
				}
			}
		}
	})

	// 2. Extract from <img> tags (thumbnails, fallback), but skip if a similar <a href=...> image exists
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		attrs := []string{"src", "data-src", "data-lazy", "data-original"}
		for _, attr := range attrs {
			if v, ok := s.Attr(attr); ok {
				v = strings.TrimSpace(v)
				if v == "" {
					continue
				}
				// resolve relative URLs
				if !strings.HasPrefix(v, "http") && !strings.HasPrefix(v, "data:") {
					u, err := url.Parse(v)
					if err == nil {
						v = base.ResolveReference(u).String()
					}
				}
				// skip if a higher-quality image (from <a href=...>) with the same basename exists
				imgBase := strings.ToLower(filepath.Base(v))
				skip := false
				for u := range found {
					if strings.Contains(strings.ToLower(filepath.Base(u)), strings.TrimSuffix(imgBase, filepath.Ext(imgBase))) {
						skip = true
						break
					}
				}
				if skip {
					continue
				}
				if _, exists := found[v]; !exists {
					found[v] = struct{}{}
					urls = append(urls, v)
				}
			}
		}
		if v, ok := s.Attr("srcset"); ok {
			for _, part := range strings.Split(v, ",") {
				p := strings.Fields(strings.TrimSpace(part))
				if len(p) > 0 {
					vv := p[0]
					vv = strings.TrimSpace(vv)
					if vv == "" {
						continue
					}
					if !strings.HasPrefix(vv, "http") && !strings.HasPrefix(vv, "data:") {
						u, err := url.Parse(vv)
						if err == nil {
							vv = base.ResolveReference(u).String()
						}
					}
					imgBase := strings.ToLower(filepath.Base(vv))
					skip := false
					for u := range found {
						if strings.Contains(strings.ToLower(filepath.Base(u)), strings.TrimSuffix(imgBase, filepath.Ext(imgBase))) {
							skip = true
							break
						}
					}
					if skip {
						continue
					}
					if _, exists := found[vv]; !exists {
						found[vv] = struct{}{}
						urls = append(urls, vv)
					}
				}
			}
		}
	})
	return urls, nil
}
