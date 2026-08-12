// Protocol Buffers - Google's data interchange format
// Copyright 2008 Google Inc.  All rights reserved.
// Copyright 2022 Teppei Fukuda.  All rights reserved.
// https://developers.google.com/protocol-buffers/

// Package fieldmaskpb contains a WASM-friendly copy of
// google.protobuf.FieldMask. The helpers in this file mirror upstream's
// reflection-free operations (Append, Union, Intersect, Normalize, IsValid).
// Reflection-dependent helpers (such as the upstream form of New/Append/IsValid
// that take a proto.Message and validate paths against its descriptor) are
// intentionally omitted because protoreflect cannot be used when plugins are
// compiled to wasip1.
package fieldmaskpb

import "sort"

// Append adds the given paths to the FieldMask in order.
//
// No validation is performed. Call IsValid afterwards if the input may be
// untrusted.
func (x *FieldMask) Append(paths ...string) {
	x.Paths = append(x.Paths, paths...)
}

// Union returns a new FieldMask whose paths are the union of the input masks'
// paths. The result is normalized (see Normalize).
func Union(masks ...*FieldMask) *FieldMask {
	out := &FieldMask{}
	for _, m := range masks {
		if m == nil {
			continue
		}
		out.Paths = append(out.Paths, m.Paths...)
	}
	out.Normalize()
	return out
}

// Intersect returns a new FieldMask containing only paths that are covered by
// every input mask. A path p is covered by a mask if the mask contains p or
// an ancestor of p at a path-segment boundary. The result is normalized.
//
// Calling Intersect with no masks returns an empty FieldMask.
func Intersect(masks ...*FieldMask) *FieldMask {
	if len(masks) == 0 {
		return &FieldMask{}
	}
	out := &FieldMask{Paths: append([]string(nil), masks[0].GetPaths()...)}
	out.Normalize()

	for _, m := range masks[1:] {
		next := append([]string(nil), m.GetPaths()...)
		other := &FieldMask{Paths: next}
		other.Normalize()

		var merged []string
		for _, p := range out.Paths {
			if covers(other.Paths, p) {
				merged = append(merged, p)
			}
		}
		for _, p := range other.Paths {
			if covers(out.Paths, p) {
				merged = append(merged, p)
			}
		}
		out.Paths = merged
		out.Normalize()
		if len(out.Paths) == 0 {
			return out
		}
	}
	return out
}

// Normalize sorts paths lexicographically and removes redundancy: duplicate
// paths are collapsed and any path covered by a parent path is dropped (if
// "foo" is in the mask, "foo.bar" is removed because the parent already
// designates everything beneath it).
func (x *FieldMask) Normalize() {
	if x == nil || len(x.Paths) == 0 {
		return
	}
	sort.Strings(x.Paths)
	out := x.Paths[:0]
	var last string
	for i, p := range x.Paths {
		if i > 0 && p == last {
			continue
		}
		if last != "" && isDescendant(last, p) {
			continue
		}
		out = append(out, p)
		last = p
	}
	x.Paths = out
}

// IsValid reports whether every path in the mask is syntactically well-formed:
// non-empty, dot-separated, with each segment matching [a-z][a-z0-9_]* (the
// proto field naming convention).
//
// NOTE: This differs from google.golang.org/protobuf/types/known/fieldmaskpb.
// Upstream's IsValid takes a proto.Message and additionally verifies each path
// resolves to an actual field on that message via the message descriptor. That
// descriptor-based check requires protoreflect, which is unavailable when
// plugins are compiled to WASM (wasip1) — the same reason go-plugin ships its
// own copies of the well-known types. Perform descriptor-based validation on
// the host side if you need it.
func (x *FieldMask) IsValid() bool {
	if x == nil {
		return false
	}
	for _, p := range x.Paths {
		if !isValidPath(p) {
			return false
		}
	}
	return true
}

// isDescendant reports whether child is a strict descendant of parent, where
// child has the form "<parent>.<rest>".
func isDescendant(parent, child string) bool {
	if len(child) <= len(parent) {
		return false
	}
	if child[:len(parent)] != parent {
		return false
	}
	return child[len(parent)] == '.'
}

// covers reports whether any path in mask is equal to p or a parent of p.
func covers(mask []string, p string) bool {
	for _, m := range mask {
		if m == p || isDescendant(m, p) {
			return true
		}
	}
	return false
}

// isValidPath reports whether p is a syntactically valid FieldMask path:
// dot-separated, each segment matching [a-z][a-z0-9_]*.
func isValidPath(p string) bool {
	if p == "" {
		return false
	}
	segStart := 0
	for i := 0; i <= len(p); i++ {
		if i == len(p) || p[i] == '.' {
			if !isValidSegment(p[segStart:i]) {
				return false
			}
			segStart = i + 1
		}
	}
	return true
}

func isValidSegment(s string) bool {
	if s == "" {
		return false
	}
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
		case c >= '0' && c <= '9':
		case c == '_':
		default:
			return false
		}
	}
	return true
}
