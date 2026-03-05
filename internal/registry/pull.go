package registry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/schollz/progressbar/v3"
)

type ProgressUpdate struct {
	Progress int64
	Total    int64
}

type PullOptions struct {
	OutputDir    string
	ShowProgress bool
	ProgressChan chan<- ProgressUpdate
}

type safeProgressWriter struct {
	bar          *progressbar.ProgressBar
	progressChan chan<- ProgressUpdate
	total        int64
	current      int64
}

func (w *safeProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.current += int64(n)

	if w.bar != nil {
		if w.current <= w.total {
			_ = w.bar.Add(n)
		} else {
			remaining := w.total - (w.current - int64(n))
			if remaining > 0 {
				_ = w.bar.Add(int(remaining))
			}
		}
	}

	if w.progressChan != nil {
		select {
		case w.progressChan <- ProgressUpdate{Progress: w.current, Total: w.total}:
		default:
		}
	}

	return n, nil
}

func SaveImageToTar(img v1.Image, ref *ImageReference, opts *PullOptions) (string, error) {
	if opts == nil {
		opts = &PullOptions{}
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	filename := fmt.Sprintf("%s_%s.tar", ref.Name, ref.Tag)
	filename = sanitizeFilename(filename)
	outputPath := filepath.Join(outputDir, filename)

	file, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create tar file: %w", err)
	}
	defer file.Close()

	var writer io.Writer = file
	var bar *progressbar.ProgressBar
	var totalSize int64

	if opts.ShowProgress || opts.ProgressChan != nil {
		imgSize, _ := img.Size()
		layerSize, _ := getImageSize(img)
		totalSize = imgSize + layerSize

		if opts.ShowProgress {
			bar = progressbar.DefaultBytes(totalSize, "Writing")
		}

		writer = io.MultiWriter(file, &safeProgressWriter{
			bar:          bar,
			progressChan: opts.ProgressChan,
			total:        totalSize,
			current:      0,
		})
	}

	tag, err := parseTag(ref.String())
	if err != nil {
		return "", fmt.Errorf("parse tag: %w", err)
	}

	if err := tarball.Write(tag, img, writer); err != nil {
		return "", fmt.Errorf("write tarball: %w", err)
	}

	if bar != nil {
		_ = bar.Finish()
	}

	return outputPath, nil
}

func sanitizeFilename(name string) string {
	replacements := map[string]string{
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}
	for old, new := range replacements {
		name = strings.ReplaceAll(name, old, new)
	}
	return name
}

func getImageSize(img v1.Image) (int64, error) {
	layers, err := img.Layers()
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			return 0, err
		}
		totalSize += size
	}
	return totalSize, nil
}

func parseTag(ref string) (name.Tag, error) {
	return name.NewTag(ref)
}

func (c *Client) PullAndSave(refStr, arch, outputDir string, progressChan chan<- ProgressUpdate) error {
	ref, err := ParseImageRef(refStr, arch, "")
	if err != nil {
		return err
	}

	img, err := c.PullImage(ref)
	if err != nil {
		return err
	}

	_, err = SaveImageToTar(img, ref, &PullOptions{
		OutputDir:    outputDir,
		ShowProgress: false,
		ProgressChan: progressChan,
	})
	return err
}
