package jwt

import (
	"reflect"
	"testing"
	"time"
)

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		time          JSONTime
		want          string
		expectedError bool
	}{
		{
			name:          "happy path",
			time:          JSONTime(time.Date(2025, time.July, 17, 12, 0, 0, 0, time.UTC)),
			want:          "1752753600",
			expectedError: false,
		},
		{
			name:          "JSONTime is empty",
			time:          JSONTime(time.Time{}),
			want:          "0",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.time.MarshalJSON()
			if tt.expectedError {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}
			if tt.want != string(got) {
				t.Fatalf("Expected %v, got %v", tt.want, string(got))
			}
		})
	}
}

func TestUnMarshalJSON(t *testing.T) {
	tests := []struct {
		name          string
		time          string
		want          JSONTime
		expectedError bool
	}{
		{
			name:          "happy path",
			time:          "1752753600",
			want:          JSONTime(time.Date(2025, time.July, 17, 15, 0, 0, 0, time.Local)),
			expectedError: false,
		},
		{
			name:          "JSONTime is float",
			time:          "1652824443.7082417",
			want:          JSONTime(time.Date(2022, time.May, 18, 00, 54, 03, 0, time.Local)),
			expectedError: false,
		},
		{
			name: "JSONTime is zero",
			time: "0",
			want: JSONTime(time.Date(1970, time.January, 1, 3, 0, 0, 0, time.Local)),

			expectedError: false,
		},
		{
			name:          "JSONTime is string",
			time:          "\"1652824443.7082417\"",
			want:          JSONTime(time.Time{}),
			expectedError: true,
		},
		{
			name:          "JSONTime is empty",
			time:          "",
			want:          JSONTime(time.Time{}),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got JSONTime
			err := got.UnmarshalJSON([]byte(tt.time))
			if tt.expectedError {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				return
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Fatalf("Expected %v, got %v", tt.want, got)
			}
		})
	}
}
