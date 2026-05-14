package docker

import (
	"fmt"
	"strings"

	"github.com/distribution/reference"
)

const (
	dockerAddr        = "docker.io"
	dockerLibraryAddr = "docker.io/library"
)

// Reference represents a parsed reference to docker image.
// Image contains image name with tag and/or digest and Registry contains registry's domain.
type Reference struct {
	Image    string
	Registry string
}

func ParseReference(ref string) (*Reference, error) {
	imageRef, err := reference.ParseDockerRef(ref)
	if err != nil {
		return nil, fmt.Errorf("can't parse image name: %w", err)
	}

	// Skip default docker registry
	registry := reference.Domain(imageRef)
	if registry == dockerAddr {
		registry = ""
	}

	image := cutImageRegistry(imageRef.Name(), registry)
	tag, digest := parseTagDigest(ref)

	if tag == "" && digest == "" {
		tag = "latest"
	}
	if tag != "" {
		image = image + ":" + tag
	}
	if digest != "" {
		image = image + "@" + digest
	}

	r := &Reference{image, registry}

	return r, nil
}

func cutImageRegistry(name, registry string) string {
	prefixes := []string{dockerLibraryAddr, dockerAddr, registry}

	for _, prefix := range prefixes {
		if p := prefix + "/"; strings.HasPrefix(name, p) {
			return strings.TrimPrefix(name, p)
		}
	}

	return name
}

// parseTagDigest parses a Docker image reference and returns its tag and digest.
// It returns both tag and digest if both are present in ref.
// If only tag or only digest is present, tag or digest is returned respectively.
// If none are present, empty strings are returned.
func parseTagDigest(ref string) (tag string, digest string) {
	if strings.Contains(ref, "@") { // we've got digest
		parts := strings.SplitN(ref, "@", 2)

		repo := parts[0]
		digest = parts[1]

		n := strings.LastIndex(repo, ":")
		if n < 0 {
			return "", digest
		}

		if t := repo[n+1:]; !strings.Contains(t, "/") {
			tag = t
		}

		return tag, digest
	}

	// No digest present, check for tag only
	n := strings.LastIndex(ref, ":")
	if n < 0 {
		return "", ""
	}

	if t := ref[n+1:]; !strings.Contains(t, "/") {
		tag = t
	}

	return tag, ""
}

// ImagePath returns the image reference with tag and digest stripped.
func (r Reference) ImagePath() string {
	path := r.Image
	if idx := strings.Index(path, "@"); idx != -1 {
		path = path[:idx]
	}
	if idx := strings.LastIndex(path, ":"); idx != -1 {
		if !strings.Contains(path[idx+1:], "/") {
			path = path[:idx]
		}
	}
	return path
}

// ImageWithTag returns the image reference with tag but strips digest.
func (r Reference) ImageWithTag() string {
	path := r.Image
	if idx := strings.Index(path, "@"); idx != -1 {
		path = path[:idx]
	}
	return path
}

// FullImageRef reconstructs the full OCI reference including registry.
// For Docker Hub images (empty registry), the image is returned as-is.
func (r Reference) FullImageRef() string {
	if r.Registry == "" {
		return r.Image
	}
	return r.Registry + "/" + r.Image
}
