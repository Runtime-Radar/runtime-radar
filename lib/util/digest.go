package util

import "github.com/opencontainers/go-digest"

func IsValidDigest(d string) bool {
	dgst, err := digest.Parse(d)
	if err != nil {
		return false // Format is invalid
	}
	// Check if the algorithm is specifically sha256
	return dgst.Algorithm() == digest.SHA256
}
