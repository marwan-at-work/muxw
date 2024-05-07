package muxw

import "testing"

func TestIsRemainderPattern(t *testing.T) {
	for _, tc := range [...]struct {
		name string
		path string
		want bool
	}{
		{
			name: "valid",
			path: "/{hello...}",
			want: true,
		},
		{
			name: "multi_segment",
			path: "/one/{two...}",
			want: true,
		},
		{
			name: "slash",
			path: "/",
			want: false,
		},
		{
			name: "empty",
			path: "",
			want: false,
		},
		{
			name: "wildcard",
			path: "/{hello}",
			want: false,
		},
		{
			name: "wildcard",
			path: "/{hello}",
			want: false,
		},
		{
			name: "multi_segment_false",
			path: "/one/{two}",
			want: false,
		},
		{
			name: "bad_pattern",
			path: "/one/...}",
			want: false,
		},
		{
			name: "bad_wildcard_pattern",
			path: "/one/{...}",
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := isRemainderPattern(tc.path)
			if got != tc.want {
				t.Fatalf("expected isRemainderPattern to return %t when given %s but got %t", tc.want, tc.path, got)
			}
		})
	}
}
