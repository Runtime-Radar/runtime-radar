package database

import (
	"errors"
	"testing"
)

func TestSanitizeOrder(t *testing.T) {
	tests := []struct {
		name          string
		order         any
		want          any
		expectedError error
	}{
		{
			name:          "happy path",
			order:         " expires_At asc NULLS last , nAmE DeSC   ",
			want:          "expires_at asc nulls last,name desc",
			expectedError: nil,
		},
		{
			name:          "only field", // direction is optional
			order:         " name   ",
			want:          "name",
			expectedError: nil,
		},
		{
			name:          "default order",
			order:         "created_at desc",
			want:          "created_at desc",
			expectedError: nil,
		},
		{
			name:          "nil",
			order:         nil,
			want:          nil,
			expectedError: nil,
		},
		{
			name:          "empty",
			order:         "",
			want:          nil,
			expectedError: nil,
		},
		{
			name:          "unsupported order type",
			order:         123,
			want:          nil,
			expectedError: ErrInvalidOrder,
		},
		{
			name:          "invalid order",
			order:         "unknown DeSC",
			want:          nil,
			expectedError: ErrInvalidOrder,
		},
		{
			name:          "sql injection",
			order:         "(select 1 from pg_sleep(3))",
			want:          nil,
			expectedError: ErrInvalidOrder,
		},
		{
			name:          "sql injection appended to allowed field",
			order:         "created_at desc, (select 1 from pg_sleep(3))",
			want:          nil,
			expectedError: ErrInvalidOrder,
		},
		{
			name:          "hidden column",
			order:         "hash",
			want:          nil,
			expectedError: ErrInvalidOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeOrder(tt.order)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("Expected error to be %v, got %v", tt.expectedError, err)
			}
			if tt.want != got {
				t.Fatalf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}
