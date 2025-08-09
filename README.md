
# Image & Video Scraper (Go)

>A modern, modular Go application to scrape images and videos from any web page, with a beautiful web UI and robust edge-case handling.

---

## GitHub

- **Repository:** [Image-Scrapper-Advanced](https://github.com/TheRealShek/Image-Scrapper-Advanced)
- **Contributions:** Pull requests are welcome! Please open an issue first to discuss major changes.
- **Issues:** If you find a bug or have a feature request, open an issue on the [GitHub Issues page](https://github.com/TheRealShek/Image-Scrapper-Advanced/issues).


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
See [`WORKFLOW.md`](./WORKFLOW.md) for detailed file and module descriptions.

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
