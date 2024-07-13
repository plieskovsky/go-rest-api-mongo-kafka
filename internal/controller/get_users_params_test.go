package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"net/http"
	url2 "net/url"
	"testing"
	"user-service/internal/model"
)

func Test_parseSortBy(t *testing.T) {
	tests := []struct {
		name    string
		sortBy  string
		want    *model.Sort
		wantErr bool
	}{
		{
			name:    "created and asc type",
			sortBy:  "created_at.asc",
			want:    &model.Sort{Field: "created_at", Type: "asc"},
			wantErr: false,
		},
		{
			name:    "updated and desc type",
			sortBy:  "updated_at.desc",
			want:    &model.Sort{Field: "updated_at", Type: "desc"},
			wantErr: false,
		},
		{
			name:    "last name and desc type",
			sortBy:  "last_name.desc",
			want:    &model.Sort{Field: "last_name", Type: "desc"},
			wantErr: false,
		},
		{
			name:    "first name and desc type",
			sortBy:  "first_name.desc",
			want:    &model.Sort{Field: "first_name", Type: "desc"},
			wantErr: false,
		},
		{
			name:    "nickname and desc type",
			sortBy:  "nickname.desc",
			want:    &model.Sort{Field: "nickname", Type: "desc"},
			wantErr: false,
		},
		{
			name:    "password and desc type",
			sortBy:  "password.desc",
			want:    &model.Sort{Field: "password", Type: "desc"},
			wantErr: false,
		},
		{
			name:    "email and desc type",
			sortBy:  "email.desc",
			want:    &model.Sort{Field: "email", Type: "desc"},
			wantErr: false,
		},
		{
			name:    "unsupported field and desc type",
			sortBy:  "unknown.desc",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			sortBy:  "email.unknown",
			wantErr: true,
		},
		{
			name:    "incorrect format",
			sortBy:  "email.desc.another",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSortBy(tt.sortBy)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseFilterFields(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  model.FilterFields
	}{
		{
			name:  "first name",
			query: "first_name=John",
			want: model.FilterFields{
				FirstName: "John",
			},
		},
		{
			name:  "last name",
			query: "last_name=Wick",
			want: model.FilterFields{
				LastName: "Wick",
			},
		},
		{
			name:  "nickname",
			query: "nickname=johnywicky",
			want: model.FilterFields{
				Nickname: "johnywicky",
			},
		},
		{
			name:  "email",
			query: "email=john.wick@example.com",
			want: model.FilterFields{
				Email: "john.wick@example.com",
			},
		},
		{
			name:  "country",
			query: "country=UK",
			want: model.FilterFields{
				Country: "UK",
			},
		},
		{
			name:  "unknown",
			query: "unknown=idk",
			want:  model.FilterFields{},
		},
		{
			name:  "all present",
			query: "first_name=John&last_name=Wick&nickname=johnywicky&email=john.wick@example.com&country=UK",
			want: model.FilterFields{
				FirstName: "John",
				LastName:  "Wick",
				Nickname:  "johnywicky",
				Email:     "john.wick@example.com",
				Country:   "UK",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := gin.Context{
				Request: &http.Request{
					URL: &url2.URL{
						RawQuery: tt.query,
					},
				},
			}

			got := parseFilterFields(&ctx)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseGetUsersParams(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    *model.GetUsersParams
		wantErr bool
	}{
		{
			name:  "empty query - defaults",
			query: "",
			want: &model.GetUsersParams{
				PageSize: 20,
				Page:     0,
				Sort: model.Sort{
					Field: "last_name",
					Type:  "asc",
				},
			},
			wantErr: false,
		},
		{
			name:  "unknown query - defaults",
			query: "unknown=idk",
			want: &model.GetUsersParams{
				PageSize: 20,
				Page:     0,
				Sort: model.Sort{
					Field: "last_name",
					Type:  "asc",
				},
			},
			wantErr: false,
		},
		{
			name:  "page size",
			query: "pageSize=13",
			want: &model.GetUsersParams{
				PageSize: 13,
				Page:     0,
				Sort: model.Sort{
					Field: "last_name",
					Type:  "asc",
				},
			},
			wantErr: false,
		},
		{
			name:  "page",
			query: "page=7",
			want: &model.GetUsersParams{
				PageSize: 20,
				Page:     7,
				Sort: model.Sort{
					Field: "last_name",
					Type:  "asc",
				},
			},
			wantErr: false,
		},
		{
			name:  "sorting",
			query: "sortBy=first_name.desc",
			want: &model.GetUsersParams{
				PageSize: 20,
				Page:     0,
				Sort: model.Sort{
					Field: "first_name",
					Type:  "desc",
				},
			},
			wantErr: false,
		},
		{
			name:  "filters",
			query: "nickname=punisher&email=test@bubu.com",
			want: &model.GetUsersParams{
				PageSize: 20,
				Page:     0,
				Sort: model.Sort{
					Field: "last_name",
					Type:  "asc",
				},
				FilterFields: model.FilterFields{
					Nickname: "punisher",
					Email:    "test@bubu.com",
				},
			},
			wantErr: false,
		},
		{
			name:  "all fields combined",
			query: "pageSize=13&page=4&sortBy=first_name.desc&nickname=punisher&email=test@bubu.com",
			want: &model.GetUsersParams{
				PageSize: 13,
				Page:     4,
				Sort: model.Sort{
					Field: "first_name",
					Type:  "desc",
				},
				FilterFields: model.FilterFields{
					Nickname: "punisher",
					Email:    "test@bubu.com",
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid page",
			query:   "page=notNumber",
			wantErr: true,
		},
		{
			name:    "invalid page size",
			query:   "pageSize=notNumber",
			wantErr: true,
		},
		{
			name:    "invalid sort by",
			query:   "sortBy=invalid_format",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := gin.Context{
				Request: &http.Request{
					URL: &url2.URL{
						RawQuery: tt.query,
					},
				},
			}

			got, err := parseGetUsersParams(&ctx)

			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
