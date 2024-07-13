package metrics

import (
	"github.com/go-playground/assert/v2"
	"testing"
)

func Test_removeDynamicPathParams(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "path without single sub path",
			path: "/metrics",
			want: "/metrics",
		},
		{
			name: "path with 2 sub paths",
			path: "/v1/users",
			want: "/v1/users",
		},
		{
			name: "path with 3 sub paths - last is dynamic",
			path: "/v1/users/668e57eb1ae7770b85ca64ad",
			want: "/v1/users",
		},
		{
			name: "path with 5 sub paths - last is dynamic",
			path: "/v1/users/668e57eb1ae7770b85ca64ad/something/nice",
			want: "/v1/users",
		},
		{
			name: "with query params",
			path: "/v1/users/another?pageSize=2&page=0",
			want: "/v1/users",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDynamicPathParams(tt.path)

			assert.Equal(t, tt.want, got)
		})
	}
}
