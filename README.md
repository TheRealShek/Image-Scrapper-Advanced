# Image & Video Scraper

A modular Go application to scrape images and videos from a web page.

## File Structure

- `main.go` — Entry point. Handles CLI flags, output directory, and calls the scraper and downloader.
- `scraper.go` — Contains the `Scrape` function to extract image and video URLs from a web page.
- `downloader.go` — Contains the `Download` function to download files from URLs to disk.

## Usage


```
go run . -url <page_url> [-out <output_dir>] [-type image|video|all]
```
- `-url` (required): The web page to scrape.
- `-out`: Output directory (default: `./videos`).
- `-type`: What to scrape: `image`, `video`, or `all` (default: `all`).

## Flow

1. **main.go** parses flags and creates the output directory.
2. Renders the page with JavaScript using a headless browser.
3. Extracts image and/or video URLs from the rendered HTML.
4. Downloads each found image or video to the output directory.

## Requirements
- Go 1.18+
- [github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery)

## Example

```
go run . -url https://example.com -type all
```

This will download all images and videos found on the page to the `./videos` directory (or your specified output directory).
