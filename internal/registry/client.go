package registry

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type Client struct {
	auth authn.Authenticator
}

type ImageReference struct {
	Registry   string
	Repository string
	Name       string
	Tag        string
	Arch       string
	OS         string
}

func NewClient() *Client {
	return &Client{
		auth: authn.Anonymous,
	}
}

func (c *Client) WithAuth(auth authn.Authenticator) *Client {
	c.auth = auth
	return c
}

func ParseImageRef(refStr, arch, os string) (*ImageReference, error) {
	ref, err := name.ParseReference(refStr)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	var tag string
	switch r := ref.(type) {
	case name.Tag:
		tag = r.TagStr()
	case name.Digest:
		tag = r.DigestStr()
	default:
		tag = "latest"
	}

	fullRef := ref.Name()
	parts := strings.SplitN(fullRef, "/", 3)

	var registry, repository, name string
	if len(parts) >= 3 && (strings.Contains(parts[0], ".") || strings.Contains(parts[0], ":")) {
		registry = parts[0]
		repository = parts[1] + "/" + parts[2]
		name = parts[2]
	} else if len(parts) == 2 {
		registry = "index.docker.io"
		repository = parts[0] + "/" + parts[1]
		name = parts[1]
	} else {
		registry = "index.docker.io"
		repository = "library/" + parts[0]
		name = parts[0]
	}

	if idx := strings.Index(name, ":"); idx != -1 {
		name = name[:idx]
	}

	if arch == "" {
		arch = runtime.GOARCH
	}
	if os == "" {
		os = runtime.GOOS
	}

	return &ImageReference{
		Registry:   registry,
		Repository: repository,
		Name:       name,
		Tag:        tag,
		Arch:       arch,
		OS:         os,
	}, nil
}

func (ir *ImageReference) String() string {
	if ir.Tag == "" {
		return fmt.Sprintf("%s/%s", ir.Registry, ir.Repository)
	}
	return fmt.Sprintf("%s/%s:%s", ir.Registry, ir.Repository, ir.Tag)
}

func (c *Client) PullImage(ref *ImageReference) (v1.Image, error) {
	imgRef, err := name.ParseReference(ref.String())
	if err != nil {
		return nil, fmt.Errorf("parse reference: %w", err)
	}

	img, err := remote.Image(imgRef, remote.WithAuth(c.auth))
	if err != nil {
		return nil, fmt.Errorf("pull image: %w", err)
	}

	return img, nil
}
