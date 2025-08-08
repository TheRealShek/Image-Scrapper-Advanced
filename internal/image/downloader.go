// Package image provides image download logic
package image

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"img-scraper/internal/common"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// DownloadFile downloads a file from the given URL to the specified directory, using a random name with an increasing number suffix.
func DownloadFile(url, outDir string, idx int) error {
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
		_, _ = rand.Read(rnd)
		rndStr := hex.EncodeToString(rnd)
		fname := fmt.Sprintf("file_%s_%03d%s", rndStr, idx, ext)
		fpath := filepath.Join(outDir, fname)
		f, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.Write(data)
		return err
	}
	// HTTP(S) download with custom User-Agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", common.RandomUserAgent())
	client := common.NewClient()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	rnd := make([]byte, 4)
	_, _ = rand.Read(rnd)
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
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	// Check file size and delete if < 50 KB
	info, err := f.Stat()
	if err == nil && info.Size() < 50*1024 {
		f.Close()
		os.Remove(fpath)
		return fmt.Errorf("file %s deleted (size < 50KB)", fname)
	}
	return nil
}

// DownloadFilesConcurrently downloads up to 10 files at a time.
func DownloadFilesConcurrently(urls []string, outDir string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	var successCount int64
	var mu sync.Mutex
	for i, u := range urls {
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
