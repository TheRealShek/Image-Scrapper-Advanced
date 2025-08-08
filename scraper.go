package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Scrape fetches the page and returns a list of image and/or video URLs found.
func Scrape(pageURL string, images, videos bool) ([]string, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("GET error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	base, _ := url.Parse(pageURL)
	found := map[string]struct{}{}
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "data:") {
			return
		}
		u, err := url.Parse(raw)
		if err != nil {
			return
		}
		res := base.ResolveReference(u).String()
		found[res] = struct{}{}
	}
	if images {
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			if v, ok := s.Attr("src"); ok {
				add(v)
			}
			for _, attr := range []string{"data-src", "data-lazy", "data-original"} {
				if v, ok := s.Attr(attr); ok {
					add(v)
				}
			}
			if v, ok := s.Attr("srcset"); ok {
				for _, part := range strings.Split(v, ",") {
					p := strings.Fields(strings.TrimSpace(part))
					if len(p) > 0 {
						add(p[0])
					}
				}
			}
		})
		doc.Find("source").Each(func(i int, s *goquery.Selection) {
			if v, ok := s.Attr("srcset"); ok {
				for _, part := range strings.Split(v, ",") {
					p := strings.Fields(strings.TrimSpace(part))
					if len(p) > 0 {
						add(p[0])
					}
				}
			}
		})
	}
	if videos {
		doc.Find("video").Each(func(i int, s *goquery.Selection) {
			if v, ok := s.Attr("src"); ok {
				add(v)
			}
		})
		doc.Find("source").Each(func(i int, s *goquery.Selection) {
			if v, ok := s.Attr("type"); ok && strings.HasPrefix(v, "video/") {
				if src, ok := s.Attr("src"); ok {
					add(src)
				}
			}
		})
	}
	var results []string
	for u := range found {
		results = append(results, u)
	}
	return results, nil
}
