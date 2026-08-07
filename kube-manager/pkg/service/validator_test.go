package service

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodGeneric_validateFilters(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
		err  error
	}{
		{
			name: "happy path",
			args: []string{"pod-* ", "  ", " k8s.node-123", " * ", "*-node"},
			want: []string{"pod-*", "k8s.node-123", "*", "*-node"},
		},
		{
			name: "bad pattern, star in the center",
			args: []string{"pod-*-123"},
			want: []string{},
			err:  ErrInvalidFilter("pod-*-123"),
		},
		{
			name: "bad pattern, starts with dash",
			args: []string{"-pod-123"},
			want: []string{},
			err:  ErrInvalidFilter("-pod-123"),
		},
		{
			name: "bad pattern, some stars",
			args: []string{"pod-***"},
			want: []string{},
			err:  ErrInvalidFilter("pod-***"),
		},
		{
			name: "bad pattern, ends with unexpected symbols",
			args: []string{"pod-123-[!@"},
			want: []string{},
			err:  ErrInvalidFilter("pod-123-[!@"),
		},
		{
			name: "bad pattern, contains unexpected symbol",
			args: []string{"pod-#123"},
			want: []string{},
			err:  ErrInvalidFilter("pod-#123"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateFilters(tt.args)
			if tt.err != nil {
				require.Error(t, err)
				assert.Equal(t, tt.err, err)

				return
			}
			require.NoError(t, err)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateFilters() got = %v, want %v", got, tt.want)
			}
		})
	}
}
