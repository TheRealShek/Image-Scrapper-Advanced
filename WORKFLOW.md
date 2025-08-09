
# Project Workflow & File Overview

This document explains the purpose and function of each file in the project.

## Top-level Files

- **main.go**: Entry point. Runs the web server and serves the web UI for scraping.
- **downloader.go**: Provides a simple function to download files from URLs (used in main package).
- **scraper.go**: Contains logic to scrape image/video URLs from a web page using goquery.
- **go.mod / go.sum**: Go module files for dependency management.
- **README.md**: Project overview, setup, and usage instructions.
- **WORKFLOW.md**: This file. Explains the workflow and file responsibilities.

## Downloaded/
- **Downloaded/**: Directory where all downloaded files are saved.

## internal/
This folder contains core modules for advanced scraping and downloading.

- **antiban.go**: Handles random User-Agent selection and HTTP client creation to avoid bans.
- **browser.go**: Uses chromedp to render JavaScript-heavy pages and extract HTML after JS execution.
- **downloader.go**: Advanced file downloader. Handles both normal URLs and data URLs, saves files with unique names.
- **extractor.go**: Extracts video URLs from HTML using goquery.
- **image_extractor.go**: Extracts image URLs from HTML, including from <a> and <img> tags, resolving relative URLs.
- **scheduler.go**: Provides a simple scheduler to run tasks at intervals (like a cron job).
- **session.go**: Stub for session/cookie management, authentication, and CAPTCHA handling.

---

**How it works:**
1. The user submits a URL via the web UI (main.go).
2. The scraper module (scraper.go) fetches and parses the page for image/video links.
3. For JavaScript-heavy sites, browser.go renders the page to extract dynamic content.
4. Extracted URLs are passed to the downloader (downloader.go or internal/downloader.go) to save files.
5. The antiban module randomizes requests to avoid detection.
6. The scheduler can automate scraping tasks if needed.
7. Session management is available for advanced use cases.
