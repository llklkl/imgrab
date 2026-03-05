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
	if t, ok := ref.(name.Tag); ok {
		tag = t.TagStr()
	} else if d, ok := ref.(name.Digest); ok {
		tag = d.DigestStr()
	} else {
		tag = "latest"
	}

	repo := ref.Context()
	registry := repo.Registry.Name()
	repository := repo.Name()

	nameStr := repository
	if idx := strings.LastIndex(nameStr, "/"); idx != -1 {
		nameStr = nameStr[idx+1:]
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
		Name:       nameStr,
		Tag:        tag,
		Arch:       arch,
		OS:         os,
	}, nil
}

func (ir *ImageReference) String() string {
	if ir.Tag == "" {
		return ir.Repository
	}
	return fmt.Sprintf("%s:%s", ir.Repository, ir.Tag)
}

func (c *Client) PullImage(ref *ImageReference) (v1.Image, error) {
	imgRef, err := name.ParseReference(ref.String())
	if err != nil {
		return nil, fmt.Errorf("parse reference: %w", err)
	}

	desc, err := remote.Get(imgRef, remote.WithAuth(c.auth))
	if err != nil {
		return nil, fmt.Errorf("get descriptor: %w", err)
	}

	if desc.MediaType.IsIndex() {
		idx, err := desc.ImageIndex()
		if err != nil {
			return nil, fmt.Errorf("get image index: %w", err)
		}
		idxManifest, err := idx.IndexManifest()
		if err != nil {
			return nil, fmt.Errorf("get index manifest: %w", err)
		}

		targetPlatform := fmt.Sprintf("%s/%s", ref.OS, ref.Arch)
		for _, manifest := range idxManifest.Manifests {
			if manifest.Platform != nil {
				platform := fmt.Sprintf("%s/%s", manifest.Platform.OS, manifest.Platform.Architecture)
				if platform == targetPlatform {
					img, err := idx.Image(manifest.Digest)
					if err != nil {
						return nil, fmt.Errorf("get image for platform %s: %w", targetPlatform, err)
					}
					return img, nil
				}
			}
		}
		return nil, fmt.Errorf("no image found for platform %s", targetPlatform)
	}

	img, err := desc.Image()
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}
	return img, nil
}
