package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"img-scraper/internal"
)

var formTmpl = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Image/Video Scraper</title>
	<style>
		body {
			background: linear-gradient(120deg, #f8fafc 0%, #e0e7ef 100%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			font-family: 'Segoe UI', Arial, sans-serif;
		}
		.card {
			background: #fff;
			border-radius: 16px;
			box-shadow: 0 4px 24px rgba(0,0,0,0.08);
			padding: 2.5rem 2.5rem 2rem 2.5rem;
			max-width: 420px;
			width: 100%;
			text-align: center;
		}
		h2 {
			margin-bottom: 1.5rem;
			color: #2d3748;
			font-weight: 600;
		}
		label {
			display: block;
			margin-bottom: 1.2rem;
			color: #4a5568;
			font-size: 1rem;
			text-align: left;
		}
		input[type="text"], select {
			width: 100%;
			padding: 0.7rem 1rem;
			margin-top: 0.3rem;
			border: 1px solid #cbd5e1;
			border-radius: 8px;
			font-size: 1rem;
			background: #f9fafb;
			transition: border 0.2s;
		}
		input[type="text"]:focus, select:focus {
			border: 1.5px solid #3182ce;
			outline: none;
		}
		input[type="submit"] {
			background: linear-gradient(90deg, #3182ce 0%, #4fd1c5 100%);
			color: #fff;
			border: none;
			border-radius: 8px;
			padding: 0.8rem 2.2rem;
			font-size: 1.1rem;
			font-weight: 600;
			cursor: pointer;
			margin-top: 0.5rem;
			box-shadow: 0 2px 8px rgba(49,130,206,0.08);
			transition: background 0.2s;
		}
		input[type="submit"]:hover {
			background: linear-gradient(90deg, #2563eb 0%, #38b2ac 100%);
		}
		.footer {
			margin-top: 2.5rem;
			color: #a0aec0;
			font-size: 0.95rem;
		}
		.result {
			margin-top: 1.5rem;
			text-align: left;
		}
		.result p {
			margin: 0.5rem 0;
		}
	</style>
</head>
<body>
	<div class="card">
		<h2>Image/Video Scraper</h2>
		<form method="POST" action="/scrape">
			<label>URL:
				<input type="text" name="url" placeholder="https://example.com" required>
			</label>
			<label>Type:
				<select name="type">
					<option value="image">Image</option>
					<option value="video">Video</option>
					<option value="all">All</option>
				</select>
			</label>
			<input type="submit" value="Scrape">
		</form>
		<div class="footer">&copy; 2025 Image/Video Scraper</div>
	</div>
</body>
</html>`

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprintf(w, "<html><body>%s</body></html>", formTmpl)
			return
		}
		// POST: handle scrape
		if err := r.ParseForm(); err != nil {
			// Show error below the card
			fmt.Fprintf(w, "<html><body><div style='display:flex;flex-direction:column;align-items:center;'>%s<div class='result'><p style='color:red'>Invalid form</p></div></div></body></html>", formTmpl)
			return
		}
		url := r.FormValue("url")
		typeOpt := r.FormValue("type")
		if url == "" {
			fmt.Fprintf(w, "<html><body><div style='display:flex;flex-direction:column;align-items:center;'>%s<div class='result'><p style='color:red'>URL required</p></div></div></body></html>", formTmpl)
			return
		}
		os.MkdirAll("Downloaded", 0755)
		html, err := internal.RenderPage(url, 50*time.Second)
		if err != nil {
			fmt.Fprintf(w, "<html><body><div style='display:flex;flex-direction:column;align-items:center;'>%s<div class='result'><p>Page render error: %v</p></div></div></body></html>", formTmpl, err)
			return
		}
		var imageURLs, videoURLs []string
		var result string
		if typeOpt == "image" || typeOpt == "all" {
			imageURLs, err = internal.ExtractImageURLs(html, url)
			if err != nil {
				result += fmt.Sprintf("<p>Image extraction error: %v</p>", err)
			} else if len(imageURLs) > 0 {
				result += fmt.Sprintf("<p>Found %d image files</p>", len(imageURLs))
				internal.DownloadFilesConcurrently(imageURLs, "Downloaded")
				result += "<p>Downloaded images to Downloaded/</p>"
			} else if typeOpt == "image" {
				result += "<p>No image URLs found.</p>"
			}
		}
		if typeOpt == "video" || typeOpt == "all" {
			videoURLs, err = internal.ExtractVideoURLs(html)
			if err != nil {
				result += fmt.Sprintf("<p>Video extraction error: %v</p>", err)
			} else if len(videoURLs) > 0 {
				result += fmt.Sprintf("<p>Found %d video files</p>", len(videoURLs))
				internal.DownloadFilesConcurrently(videoURLs, "Downloaded")
				result += "<p>Downloaded videos to Downloaded/</p>"
			} else if typeOpt == "video" {
				result += "<p>No video URLs found.</p>"
			}
		}
		// Show result below the card
		fmt.Fprintf(w, "<html><body><div style='display:flex;flex-direction:column;align-items:center;'>%s<div class='result'>%s</div></div></body></html>", formTmpl, result)
	})

	fmt.Println("Web UI running at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
