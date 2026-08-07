package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_isMatched(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		value    string
		want     bool
	}{
		{
			name:     "equal",
			patterns: []string{"p", "pod-321", "pod-*"},
			value:    "pod-321",
			want:     true,
		},
		{
			name:     "ends with",
			patterns: []string{"pod-321", "pod-*"},
			value:    "pod-123dfs",
			want:     true,
		},
		{
			name:     "starts with",
			patterns: []string{"*-pod"},
			value:    "123-pod",
			want:     true,
		},
		{
			name:     "contains",
			patterns: []string{"*-pod-*"},
			value:    "123-pod-x",
			want:     true,
		},
		{
			name:     "all",
			patterns: []string{"*"},
			value:    "pod-x",
			want:     true,
		},
		{
			name:     "empty patterns",
			patterns: []string{},
			value:    "pod-x-123",
			want:     true,
		},
		{
			name:     "empty",
			patterns: []string{""},
			value:    "pod-x-123",
			want:     false,
		},
		{
			name:     "not equal",
			patterns: []string{"pod-x-12", "pod-321", "d-*"},
			value:    "pod-x-123",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isMatched(tt.patterns, tt.value)
			require.NoError(t, err)
			if got != tt.want {
				t.Errorf("isMatched() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPodMatched(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod-321",
		},
		Spec: v1.PodSpec{
			NodeName: "node-321",
			Containers: []v1.Container{
				{Name: "container-321"},
			},
		},
	}

	tests := []struct {
		name string
		opts *ListOpts
		want bool
	}{
		{
			name: "all patterns matched",
			opts: &ListOpts{
				nodes:      []string{"node-*"},
				pods:       []string{"pod-*"},
				containers: []string{"*"},
			},
			want: true,
		},
		{
			name: "container pattern not matched",
			opts: &ListOpts{
				nodes:      []string{"node-*"},
				pods:       []string{"pod-*"},
				containers: []string{"not-matched"},
			},
			want: false,
		},
		{
			name: "pod pattern not matched",
			opts: &ListOpts{
				nodes:      []string{"node-*"},
				pods:       []string{"not-matched"},
				containers: []string{"container-*"},
			},
			want: false,
		},
		{
			name: "one pattern empty",
			opts: &ListOpts{
				nodes:      []string{},
				pods:       []string{"*"},
				containers: []string{"container-*"},
			},
			want: true,
		},
		{
			name: "all patterns empty",
			opts: &ListOpts{
				nodes:      []string{},
				pods:       []string{},
				containers: []string{},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.opts.isPodMatched(pod)
			require.NoError(t, err)
			if got != tt.want {
				t.Errorf("isPodMatched() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isNodeMatched(t *testing.T) {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-321-abc",
		},
	}

	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{
			name:     "pattern matched",
			patterns: []string{"node-*"},
			want:     true,
		},
		{
			name:     "pattern not matched",
			patterns: []string{"not-matched"},
			want:     false,
		},
		{
			name:     "empty pattern",
			patterns: []string{},
			want:     true,
		},
		{
			name:     "any",
			patterns: []string{"*"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &ListOpts{
				nodes: tt.patterns,
			}
			got, err := l.isNodeMatched(node)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_WithSort(t *testing.T) {
	list := testData()

	type sortCases struct {
		name     string
		opts     ListOpts
		expected []string
	}
	sc := []sortCases{
		{
			name:     "default",
			opts:     ListOpts{},
			expected: []string{"pod-2", "pod-3", "pod-6", "pod-1", "pod-4", "pod-5"},
		},
		{
			name: "by name",
			opts: ListOpts{
				SortBy: "name",
			},
			expected: []string{"pod-1", "pod-2", "pod-3", "pod-4", "pod-5", "pod-6"},
		},
		{
			name: "by created_at",
			opts: ListOpts{
				SortBy: "created_at",
			},
			expected: []string{"pod-6", "pod-5", "pod-2", "pod-3", "pod-4", "pod-1"},
		},
		{
			name: "by name desc",
			opts: ListOpts{
				SortBy:   "name",
				SortDesc: true,
			},
			expected: []string{"pod-6", "pod-5", "pod-4", "pod-3", "pod-2", "pod-1"},
		},
		{
			name: "by created_at desc",
			opts: ListOpts{
				SortBy:   "created_at",
				SortDesc: true,
			},
			expected: []string{"pod-1", "pod-4", "pod-3", "pod-2", "pod-5", "pod-6"},
		},
	}

	for _, tt := range sc {
		t.Run("sort "+tt.name, func(t *testing.T) {
			sorted := withSort[*v1.Pod](list, &tt.opts)
			assert.Len(t, sorted, len(tt.expected))
			names := make([]string, len(sorted))
			for i := range sorted {
				names[i] = sorted[i].Name
			}

			assert.Equal(t, tt.expected, names)
		})
	}
}

func Test_WithPagination(t *testing.T) {
	list := withSort[*v1.Pod](testData(), &ListOpts{})

	type sortCases struct {
		name     string
		opts     ListOpts
		expected []string
	}
	sc := []sortCases{
		{
			name:     "by default",
			opts:     ListOpts{},
			expected: []string{"pod-2", "pod-3", "pod-6", "pod-1", "pod-4", "pod-5"},
		},
		{
			name: "first page, with empty page_num",
			opts: ListOpts{
				PageSize: 2,
				PageNum:  0,
			},
			expected: []string{"pod-2", "pod-3"},
		},
		{
			name: "in the middle of the list",
			opts: ListOpts{
				PageSize: 2,
				PageNum:  2,
			},
			expected: []string{"pod-6", "pod-1"},
		},
		{
			name: "last empty page",
			opts: ListOpts{
				PageSize: 3,
				PageNum:  3,
			},
			expected: []string{},
		},
		{
			name: "last page",
			opts: ListOpts{
				PageSize: 4,
				PageNum:  3,
			},
			expected: []string{},
		},
	}

	for _, tt := range sc {
		t.Run("filter "+tt.name, func(t *testing.T) {
			filtered := withPagination[*v1.Pod](list, &tt.opts)

			assert.Len(t, filtered, len(tt.expected))
			names := make([]string, len(filtered))
			for i := range filtered {
				names[i] = filtered[i].Name
			}

			assert.Equal(t, tt.expected, names)
		})
	}
}

func testData() []*v1.Pod {
	tm := []time.Time{}
	for i := 0; i < 10; i++ {
		tm = append(tm, time.Now().Add(time.Duration(i)*time.Minute))
	}

	list := []*v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pod-1",
				Namespace:         "ns-2",
				CreationTimestamp: metav1.Time{Time: tm[0]},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pod-2",
				Namespace:         "ns-1",
				CreationTimestamp: metav1.Time{Time: tm[3]},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pod-3",
				Namespace:         "ns-1",
				CreationTimestamp: metav1.Time{Time: tm[2]},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pod-4",
				Namespace:         "ns-2",
				CreationTimestamp: metav1.Time{Time: tm[1]},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pod-5",
				Namespace:         "ns-2",
				CreationTimestamp: metav1.Time{Time: tm[5]},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "pod-6",
				Namespace:         "ns-1",
				CreationTimestamp: metav1.Time{Time: tm[6]},
			},
		},
	}

	return list
}
