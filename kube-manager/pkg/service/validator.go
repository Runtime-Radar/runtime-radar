package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/runtime-radar/runtime-radar/kube-manager/api"
)

var (
	entityReg = regexp.MustCompile("^[a-z0-9*]([-a-z0-9.]*[a-z0-9*])?$")

	allowedSortFields = map[string]struct{}{"": {}, "created_at": {}, "name": {}}

	ErrInvalidSort   = errors.New("invalid sort parameter")
	ErrInvalidFilter = func(filter string) error {
		return fmt.Errorf("invalid filter parameter: %s", filter)
	}
)

func validateParams(namespaces, nodes, pods, containers []string, sort *api.Sort) (*ListOpts, []string, error) {
	opts := &ListOpts{}
	namespaces, err := validateFilters(namespaces)
	if err != nil {
		return nil, nil, err
	}

	if opts.nodes, err = validateFilters(nodes); err != nil {
		return nil, nil, err
	}
	if opts.pods, err = validateFilters(pods); err != nil {
		return nil, nil, err
	}
	if opts.containers, err = validateFilters(containers); err != nil {
		return nil, nil, err
	}

	if sort == nil {
		return opts, namespaces, nil
	}

	if _, ok := allowedSortFields[sort.GetField()]; !ok {
		return nil, nil, ErrInvalidSort
	}
	opts.SortBy = sort.GetField()

	if strings.ToLower(sort.GetKey()) == "desc" {
		opts.SortDesc = true
	}

	return opts, namespaces, nil
}

func validateFilters(req []string) ([]string, error) {
	var tmpls []string
	for _, t := range req {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}

		if !entityReg.Match([]byte(t)) {
			return []string{}, ErrInvalidFilter(t)
		}

		tmpls = append(tmpls, t)
	}

	return tmpls, nil
}
