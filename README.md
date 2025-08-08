
# Image & Video Scraper (Go)

>A modern, modular Go application to scrape images and videos from any web page, with a beautiful web UI and robust edge-case handling.

---

## Features
- **Web UI**: Paste a URL, select scrape type (image/video/all), and download with a click.
- **Dynamic Content**: Uses a headless browser to render JavaScript-heavy sites.
- **Smart Extraction**: Grabs full-res images and videos, avoids thumbnails, and handles data/relative URLs.
- **Concurrent Downloads**: Fast, multi-threaded downloading.
- **Auto File Type Detection**: Files are saved with the correct extension (jpg, png, mp4, etc.).
- **Auto-Cleanup**: Skips and deletes files smaller than 50KB (configurable).

---

## Getting Started

### 1. Clone & Install
```sh
git clone https://github.com/TheRealShek/Image-Scrapper-Advanced.git
cd Image-Scrapper-Advanced
go mod tidy
```

### 2. Create the Download Folder
> **Important:** The `Downloaded` folder is ignored by git (see `.gitignore`). You must create it yourself:
```sh
mkdir Downloaded
```

### 3. Run the Web UI
```sh
go run main.go
```
Then open [http://localhost:8080/](http://localhost:8080/) in your browser.

---

## Usage

### Web UI
1. Paste the target URL.
2. Select what to scrape: **Image**, **Video**, or **All**.
3. Click **Scrape**. Downloads will appear in the `Downloaded/` folder.

### CLI (if enabled)
```sh
go run . -url <page_url> [-out <output_dir>] [-type image|video|all]
```

---

## File Structure
- `main.go` — Entry point, web server, and UI logic.
- `internal/` — Modular logic for extraction, downloading, anti-ban, etc.
- `.gitignore` — Excludes `Downloaded/` from git.

---

## Requirements
- Go 1.18+
- [github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery)
- [chromedp](https://github.com/chromedp/chromedp) (for headless browser rendering)

---

## Example

**Web UI:**
> Paste `https://example.com` and select `All` to download all images and videos to `Downloaded/`.

**CLI:**
```sh
go run . -url https://example.com -type all
```

---
