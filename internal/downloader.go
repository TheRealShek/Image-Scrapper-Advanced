package internal

import (
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/chromedp/chromedp"
	"io"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DownloadImagesAdvancedBatch downloads images concurrently using AdvancedDownloadFile, with per-domain rate limiting, cookie reuse, and stats.
func DownloadImagesAdvancedBatch(imgURLs []string, pageURL, outDir string) {
	type result struct {
		url     string
		method  string // basic/cookies/browser
		errType string
		err     error
	}
	var (
		wg             sync.WaitGroup
		mu             sync.Mutex
		results        []result
		domainLastTime = make(map[string]time.Time)
		domainDelay    = 1200 * time.Millisecond // per-domain delay
		domainMu       = make(map[string]*sync.Mutex)
	)
	getDomain := func(rawurl string) string {
		u, _ := url.Parse(rawurl)
		return u.Host
	}
	// Group images by domain for batch cookie reuse
	type imgTask struct {
		url    string
		domain string
		idx    int
	}
	tasks := make([]imgTask, 0, len(imgURLs))
	for i, u := range imgURLs {
		tasks = append(tasks, imgTask{url: u, domain: getDomain(u), idx: i + 1})
	}
	// Worker pool
	maxWorkers := 5
	jobs := make(chan imgTask)
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				d := task.domain
				// Per-domain mutex for rate limiting
				mu.Lock()
				if domainMu[d] == nil {
					domainMu[d] = &sync.Mutex{}
				}
				domMu := domainMu[d]
				mu.Unlock()
				domMu.Lock()
				last := domainLastTime[d]
				wait := domainDelay - time.Since(last)
				if wait > 0 {
					time.Sleep(wait)
				}
				domainLastTime[d] = time.Now()
				domMu.Unlock()
				// Try download, track escalation method
				method := "basic"
				err := AdvancedDownloadFileWithStats(task.url, pageURL, outDir, task.idx, &method)
				mu.Lock()
				r := result{url: task.url, method: method}
				if err != nil {
					r.err = err
					if strings.Contains(err.Error(), "403") {
						r.errType = "403 Forbidden"
					} else if strings.Contains(err.Error(), "timeout") {
						r.errType = "Timeout"
					} else if strings.Contains(err.Error(), "SSL") || strings.Contains(err.Error(), "tls") {
						r.errType = "SSL/TLS"
					} else {
						r.errType = "Other"
					}
				}
				results = append(results, r)
				mu.Unlock()
			}
		}()
	}
	for _, t := range tasks {
		jobs <- t
	}
	close(jobs)
	wg.Wait()
	// Stats
	stats := map[string]int{"basic": 0, "cookies": 0, "browser": 0}
	errStats := map[string]int{}
	for _, r := range results {
		stats[r.method]++
		if r.errType != "" {
			errStats[r.errType]++
		}
	}
	fmt.Printf("\nDownload summary: Success: %d, Errors: %d\n", stats["basic"]+stats["cookies"]+stats["browser"]-len(errStats), len(errStats))
	fmt.Printf("By method: basic=%d, cookies=%d, browser=%d\n", stats["basic"], stats["cookies"], stats["browser"])
	fmt.Println("Error breakdown:")
	for k, v := range errStats {
		fmt.Printf("  %s: %d\n", k, v)
	}
}

// AdvancedDownloadFileWithStats wraps AdvancedDownloadFile and sets method to escalation used.
func AdvancedDownloadFileWithStats(imgURL, pageURL, outDir string, idx int, method *string) error {
	err := AdvancedDownloadFile(imgURL, pageURL, outDir, idx)
	if err == nil {
		*method = "basic"
		return nil
	}
	if strings.Contains(err.Error(), "cookies") {
		*method = "cookies"
	} else if strings.Contains(err.Error(), "chromedp") || strings.Contains(err.Error(), "browser") {
		*method = "browser"
	}
	return err
}

// AdvancedDownloadFile downloads a file with realistic browser headers, SSL/TLS config, anti-hotlink bypass, and chromedp fallback.
// If 403 or HTTPS error, it will escalate to browser simulation and use cookies from the hosting page.
func AdvancedDownloadFile(imgURL, pageURL, outDir string, idx int) error {
	// 1. Visit the hosting page to get cookies (anti-hotlink bypass)
	jar, _ := http.DefaultClient.Jar, http.DefaultClient.Jar // use default jar if available
	client := &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:   nil, // use system certs
			DisableKeepAlives: false,
			ForceAttemptHTTP2: true,
		},
		Jar: jar,
	}
	// Visit the page to get cookies
	pageReq, _ := http.NewRequest("GET", pageURL, nil)
	pageReq.Header.Set("User-Agent", RandomUserAgent())
	pageReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	pageReq.Header.Set("Accept-Language", "en-US,en;q=0.9")
	pageReq.Header.Set("Connection", "keep-alive")
	_, _ = client.Do(pageReq) // ignore error, just for cookies

	// 2. Prepare realistic browser headers for image request
	req, err := http.NewRequest("GET", imgURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", RandomUserAgent())
	req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", pageURL)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "image")
	req.Header.Set("Sec-Fetch-Mode", "no-cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	// 3. Try download with retries and error handling
	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := client.Do(req)
		if err != nil {
			// HTTPS error: try with InsecureSkipVerify (not recommended for prod)
			if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") && attempt == 0 {
				client.Transport = &http.Transport{
					TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
					DisableKeepAlives: false,
					ForceAttemptHTTP2: true,
				}
				continue
			}
			lastErr = err
			time.Sleep(time.Duration(500+100*attempt) * time.Millisecond)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == 403 && attempt == maxRetries-1 {
			// Escalate: use chromedp to fetch image with cookies
			return downloadWithChromedp(imgURL, outDir, idx)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("bad status: %s", resp.Status)
			time.Sleep(time.Duration(500+100*attempt) * time.Millisecond)
			continue
		}
		// Save file
		ext := filepath.Ext(imgURL)
		if len(ext) == 0 || len(ext) > 10 {
			ext = ".jpg"
		}
		rnd := make([]byte, 4)
		_, _ = crand.Read(rnd)
		rndStr := hex.EncodeToString(rnd)
		fname := fmt.Sprintf("file_%s_%03d%s", rndStr, idx, ext)
		fpath := filepath.Join(outDir, fname)
		f, err := os.Create(fpath)
		if err != nil {
			lastErr = err
			continue
		}
		defer f.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		info, err := f.Stat()
		if err == nil && info.Size() < 10*1024 {
			f.Close()
			os.Remove(fpath)
			lastErr = fmt.Errorf("file %s deleted (size < 10KB)", fname)
			continue
		}
		return nil // success
	}
	return fmt.Errorf("advanced download failed for %s: %v", imgURL, lastErr)
}

// downloadWithChromedp fetches an image using a headless browser and saves it.
func downloadWithChromedp(imgURL, outDir string, idx int) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(imgURL),
		chromedp.WaitVisible("img,body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Download the image bytes
			req, err := http.NewRequest("GET", imgURL, nil)
			if err != nil {
				return err
			}
			req.Header.Set("User-Agent", RandomUserAgent())
			req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			buf, err = io.ReadAll(resp.Body)
			return err
		}),
	)
	if err != nil {
		return err
	}
	ext := filepath.Ext(imgURL)
	if len(ext) == 0 || len(ext) > 10 {
		ext = ".jpg"
	}
	rnd := make([]byte, 4)
	_, _ = crand.Read(rnd)
	rndStr := hex.EncodeToString(rnd)
	fname := fmt.Sprintf("file_%s_%03d%s", rndStr, idx, ext)
	fpath := filepath.Join(outDir, fname)
	return os.WriteFile(fpath, buf, 0644)
}

// DownloadFile downloads a file from the given URL to the specified directory, using a random name with an increasing number suffix.
// Now with retry, user agent rotation, and rate limiting.
func DownloadFile(url, outDir string, idx int) error {
	const (
		maxRetries  = 5
		minDelay    = 500 * time.Millisecond
		maxDelay    = 10 * time.Second
		minFileSize = 50 * 1024 // 50KB
	)
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Rate limiting: sleep before each attempt (jittered)
		sleepDuration := time.Duration(int64(minDelay) + int64(attempt)*int64(minDelay))
		if sleepDuration > maxDelay {
			sleepDuration = maxDelay
		}
		if attempt > 0 {
			time.Sleep(sleepDuration + time.Duration(mrand.Intn(300))*time.Millisecond) // add jitter
		}
		if strings.HasPrefix(url, "data:image/") {
			// Handle data URL
			re := regexp.MustCompile(`^data:image/(\w+);base64,(.*)$`)
			matches := re.FindStringSubmatch(url)
			if len(matches) != 3 {
				return fmt.Errorf("invalid data url format")
			}
			ext := "." + matches[1]
			data, err := base64.StdEncoding.DecodeString(matches[2])
			if err != nil {
				return fmt.Errorf("base64 decode error: %w", err)
			}
			rnd := make([]byte, 4)
			_, _ = crand.Read(rnd)
			rndStr := hex.EncodeToString(rnd)
			fname := fmt.Sprintf("file_%s_%03d%s", rndStr, idx, ext)
			fpath := filepath.Join(outDir, fname)
			f, err := os.Create(fpath)
			if err != nil {
				lastErr = err
				continue
			}
			defer f.Close()
			_, err = f.Write(data)
			if err != nil {
				lastErr = err
				continue
			}
			return nil
		}
		// HTTP(S) download with custom User-Agent
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", RandomUserAgent())
		client := NewClient()
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		// Retry on 429, 403, 5xx, and timeouts
		if resp.StatusCode == 429 || resp.StatusCode == 403 || (resp.StatusCode >= 500 && resp.StatusCode < 600) {
			lastErr = fmt.Errorf("bad status: %s", resp.Status)
			// Optionally, parse Retry-After header for 429
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if sec, err := strconv.Atoi(retryAfter); err == nil {
					time.Sleep(time.Duration(sec) * time.Second)
				}
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("bad status: %s", resp.Status)
			continue
		}
		rnd := make([]byte, 4)
		_, _ = crand.Read(rnd)
		rndStr := hex.EncodeToString(rnd)
		ext := filepath.Ext(url)
		// Map of common content types to file extensions
		contentTypeToExt := map[string]string{
			"image/jpeg":      ".jpg",
			"image/png":       ".png",
			"image/gif":       ".gif",
			"image/webp":      ".webp",
			"image/bmp":       ".bmp",
			"image/tiff":      ".tiff",
			"video/mp4":       ".mp4",
			"video/webm":      ".webm",
			"video/ogg":       ".ogv",
			"video/quicktime": ".mov",
		}
		if len(ext) > 10 || len(ext) == 0 || ext == ".bin" {
			ctype := resp.Header.Get("Content-Type")
			if ctype != "" {
				if semi := strings.Index(ctype, ";"); semi != -1 {
					ctype = ctype[:semi]
				}
				if newExt, ok := contentTypeToExt[ctype]; ok {
					ext = newExt
				} else {
					ext = ".bin"
				}
			} else {
				ext = ".bin"
			}
		}
		fname := fmt.Sprintf("file_%s_%03d%s", rndStr, idx, ext)
		fpath := filepath.Join(outDir, fname)
		f, err := os.Create(fpath)
		if err != nil {
			lastErr = err
			continue
		}
		defer f.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			lastErr = err
			continue
		}
		// Check file size and delete if < 50 KB
		info, err := f.Stat()
		if err == nil && info.Size() < int64(minFileSize) {
			f.Close()
			os.Remove(fpath)
			lastErr = fmt.Errorf("file %s deleted (size < 50KB)", fname)
			continue
		}
		return nil // success
	}
	return fmt.Errorf("download failed for %s after %d attempts: %v", url, maxRetries, lastErr)
}

// DownloadFilesConcurrently downloads up to 5 files at a time, with global rate limiting.
func DownloadFilesConcurrently(urls []string, outDir string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // reduce concurrency for less blocking
	var successCount int64
	var mu sync.Mutex
	rateLimit := time.Tick(350 * time.Millisecond) // global rate limit
	for i, u := range urls {
		<-rateLimit // throttle globally
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func(url string, idx int) {
			defer wg.Done()
			defer func() { <-sem }() // release
			if err := DownloadFile(url, outDir, idx); err != nil {
				fmt.Println("Download error:", err)
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(u, i+1)
	}
	wg.Wait()
	fmt.Printf("Successfully downloaded %d file(s).\n", successCount)
}
