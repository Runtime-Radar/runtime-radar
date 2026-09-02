package service

import (
	"path"
	"sort"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListOpts struct {
	nodes, pods, containers []string

	// Pagination
	PageSize uint32
	PageNum  uint32

	// Sort params
	SortBy   string // namespace+name(default), name, created_at
	SortDesc bool
}

func (l *ListOpts) isPodMatched(pod *v1.Pod) (bool, error) {
	podMatched, err := isMatched(l.pods, pod.Name)
	log.Debug().Err(err).Msgf("pod matched: %s", pod.Name)
	if err != nil || !podMatched {
		return false, err
	}

	nodeMatched, err := isMatched(l.nodes, pod.Spec.NodeName)
	log.Debug().Err(err).Msgf("node matched: %s", pod.Spec.NodeName)
	if err != nil || !nodeMatched {
		return false, err
	}

	if len(l.containers) == 0 {
		return true, nil
	}

	for _, container := range pod.Spec.Containers {
		containerMatched, err := isMatched(l.containers, container.Name)
		log.Debug().Err(err).Msgf("container matched: %s", container.Name)
		if err != nil {
			return false, err
		}
		if containerMatched {
			return true, nil
		}
	}

	return false, nil
}

func (l *ListOpts) isNodeMatched(node *v1.Node) (bool, error) {
	nodeMatched, err := isMatched(l.nodes, node.Name)
	log.Debug().Err(err).Msgf("node matched: %s", node.Name)

	return nodeMatched, err
}

func withSort[T metav1.Object](list []T, opts *ListOpts) []T {
	sort.Slice(list, func(i, j int) bool {
		x, y := i, j
		if opts.SortDesc {
			x, y = j, i
		}

		switch opts.SortBy {
		case "name":
			return list[x].GetName() < list[y].GetName()

		case "created_at":
			return list[x].GetCreationTimestamp().Time.After(list[y].GetCreationTimestamp().Time)

		default:
			if list[x].GetNamespace() == list[y].GetNamespace() {
				return list[x].GetName() < list[y].GetName()
			}
			return list[x].GetNamespace() < list[y].GetNamespace()
		}
	})

	return list
}

func withPagination[T metav1.Object](list []T, opts *ListOpts) []T {
	if opts.PageSize == 0 {
		return list
	}

	pageNum := opts.PageNum
	if opts.PageNum > 0 && opts.PageSize > 0 {
		pageNum = pageNum - 1
	}

	filtered := make([]T, 0, len(list))
	first := int(pageNum * opts.PageSize)
	if first <= len(list) {
		filtered = list[first:]

		if int(opts.PageSize) < len(filtered) {
			filtered = filtered[:opts.PageSize]
		}
	}

	return filtered
}

func isMatched(patterns []string, value string) (bool, error) {
	if len(patterns) == 0 {
		return true, nil
	}

	for _, p := range patterns {
		matched, err := path.Match(p, value)
		if err != nil || matched {
			return matched, err
		}
	}

	return false, nil
}
