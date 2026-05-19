package database

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidOrder = errors.New("invalid order")

var (
	allowedOrderFields = []string{
		"count",
		"severity",
		"verdict",
		"created_at",
		"detector_id",
		// vulnerability field
		"cve",
		// component fields
		"name",
		"event_type",
		// events fields
		"registered_at",
		"source",
		"type",
		"events.created_at",
	}

	allowedOrderDirections = []string{
		"", // default order
		"asc",
		"desc",
		"asc nulls first",
		"desc nulls first",
		"asc nulls last",
		"desc nulls last",
	}

	sortsMapping map[string]struct{}
)

func init() {
	sortsMapping = make(map[string]struct{}, len(allowedOrderFields)*len(allowedOrderDirections))
	for _, dir := range allowedOrderDirections {
		for _, sort := range allowedOrderFields {
			order := sort
			if dir != "" {
				order = fmt.Sprintf("%s %s", sort, dir)
			}

			sortsMapping[order] = struct{}{}
		}
	}

}

func MapOrderToSQL(order []string) ([]string, error) {
	res := make([]string, 0, len(order))
	for _, o := range order {
		value := strings.ToLower(o)

		_, ok := sortsMapping[value]
		if !ok {
			return nil, fmt.Errorf("%w: unsupported order value: %s", ErrInvalidOrder, value)
		}

		res = append(res, value)
	}

	return res, nil
}

func orderToSlice(order any) ([]string, error) {
	ss, ok := order.(string)
	if !ok {
		return nil, fmt.Errorf("unsupported order type: %T", order)
	}

	s := strings.Split(ss, ",")
	res := make([]string, 0, len(s))

	for _, o := range s {
		res = append(res, strings.TrimSpace(o))
	}

	return res, nil
}

func SanitizeOrder(order any) (any, error) {
	s, err := SanitizeOrderAsSlice(order)
	if err != nil {
		return nil, err
	}

	return strings.Join(s, ","), nil
}

func SanitizeOrderAsSlice(order any) ([]string, error) {
	if order == nil || order == "" {
		return nil, nil
	}

	s, err := orderToSlice(order)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidOrder, err)
	}

	mapped, err := MapOrderToSQL(s)
	if err != nil {
		return nil, err
	}

	return mapped, nil
}
