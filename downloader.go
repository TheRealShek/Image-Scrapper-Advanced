package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Download fetches a file from the given URL and saves it to the output directory.
func Download(rawurl, outdir string) error {
	resp, err := http.Get(rawurl)
	if err != nil {
		return fmt.Errorf("download error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("bad status for %s: %s", rawurl, resp.Status)
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return fmt.Errorf("parse url err: %w", err)
	}
	fname := path.Base(u.Path)
	if fname == "" || fname == "." || strings.Contains(fname, "/") {
		fname = "file_" + strings.ReplaceAll(strings.ReplaceAll(u.Host+u.Path, "/", "_"), "?", "_")
	}
	fname = filepath.Join(outdir, fname)
	if _, err := os.Stat(fname); err == nil {
		fname = fname + "_1"
	}
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("create file err: %w", err)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("write file err: %w", err)
	}
	fmt.Println("Saved:", fname)
	return nil
}
