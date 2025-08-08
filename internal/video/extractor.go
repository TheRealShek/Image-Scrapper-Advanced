package video

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractVideoURLs parses HTML and returns all video URLs and metadata found.
func ExtractVideoURLs(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	var urls []string
	doc.Find("video, source").Each(func(i int, s *goquery.Selection) {
		if src, ok := s.Attr("src"); ok {
			urls = append(urls, src)
		}
	})
	return urls, nil
}
