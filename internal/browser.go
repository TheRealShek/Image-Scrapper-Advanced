package internal

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// RenderPage uses chromedp to render a page and return the HTML after JS execution.
func RenderPage(url string, timeout time.Duration) (string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()
	var html string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	if err != nil {
		return "", err
	}
	return html, nil
}
