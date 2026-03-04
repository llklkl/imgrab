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

type PullOptions struct {
	OutputDir    string
	ShowProgress bool
}

type safeProgressWriter struct {
	bar *progressbar.ProgressBar
}

func (w *safeProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	_ = w.bar.Add(n)
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

	if opts.ShowProgress {
		size, err := getImageSize(img)
		if err != nil {
			return "", fmt.Errorf("get image size: %w", err)
		}

		bar = progressbar.DefaultBytes(size, "Writing")
		writer = io.MultiWriter(file, &safeProgressWriter{bar: bar})
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

func (c *Client) PullAndSave(refStr, outputDir string) error {
	ref, err := ParseImageRef(refStr, "", "")
	if err != nil {
		return err
	}

	img, err := c.PullImage(ref)
	if err != nil {
		return err
	}

	_, err = SaveImageToTar(img, ref, &PullOptions{
		OutputDir:    outputDir,
		ShowProgress: true,
	})
	return err
}
