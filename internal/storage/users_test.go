package storage

import (
	"context"
	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
	"user-service/internal/model"
)

// Unit tests that cover the functionality of fetching the Users list, as that one is the most complex one from all storage functions.
// In a real project I would cover also all the remaining functions. The tests would look very similar to the below ones,
// therefore not writing them, as I believe the existing ones should be enough to showcase the way I would write & structure them.

func (suite *MongoTestSuite) Test_GetUsers() {
	storage := NewMongoUsersStorage(suite.db)

	userAnna := model.User{ID: uuid.New(), FirstName: "anna", LastName: "alakava", Nickname: "diff", Password: "apwd", Email: "ann@gmail.com", Country: "Austria", CreatedAt: suite.testStart, UpdatedAt: suite.testStart}
	userBeta := model.User{ID: uuid.New(), FirstName: "beta", LastName: "brumkaa", Nickname: "same", Password: "bpwd", Email: "bet@gmail.com", Country: "Austria", CreatedAt: suite.testStart, UpdatedAt: suite.testStart}
	userDenn := model.User{ID: uuid.New(), FirstName: "denn", LastName: "dobrare", Nickname: "same", Password: "cpwd", Email: "den@gmail.com", Country: "Austria", CreatedAt: suite.testStart, UpdatedAt: suite.testStart}
	userEmel := model.User{ID: uuid.New(), FirstName: "emel", LastName: "estaril", Nickname: "same", Password: "dpwd", Email: "eme@gmail.com", Country: "Egypttt", CreatedAt: suite.testStart, UpdatedAt: suite.testStart}
	userFero := model.User{ID: uuid.New(), FirstName: "fero", LastName: "farinha", Nickname: "same", Password: "fpwd", Email: "fer@gmail.com", Country: "Egypttt", CreatedAt: suite.testStart, UpdatedAt: suite.testStart}
	suite.createTestUsers(userAnna, userBeta, userDenn, userEmel, userFero)

	tests := []struct {
		name    string
		params  model.GetUsersParams
		want    []model.User
		wantErr bool
	}{
		{
			name:    "empty params",
			params:  model.GetUsersParams{},
			wantErr: true,
		},
		{
			name: "sorting by existing field - asc",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
			},
			want: []model.User{userAnna, userBeta, userDenn, userEmel, userFero},
		},
		{
			name: "sorting by existing field - desc",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "desc",
				},
			},
			want: []model.User{userFero, userEmel, userDenn, userBeta, userAnna},
		},
		{
			name: "sorting by non existing field",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "non_existent",
					Type:  "asc",
				},
			},
			want: []model.User{userAnna, userBeta, userDenn, userEmel, userFero},
		},
		{
			name: "filter by first name - existing single DB document",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				FilterFields: model.FilterFields{
					FirstName: "denn",
				},
			},
			want: []model.User{userDenn},
		},
		{
			name: "filter by country - existing multiple DB documents",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				FilterFields: model.FilterFields{
					Country: "Austria",
				},
			},
			want: []model.User{userAnna, userBeta, userDenn},
		},
		{
			name: "filter by nickname - non existing DB document",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				FilterFields: model.FilterFields{
					Nickname: "nonExisting",
				},
			},
			want: nil,
		},
		{
			name: "multiple filter fields",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				FilterFields: model.FilterFields{
					Country:  "Austria",
					Nickname: "same",
				},
			},
			want: []model.User{userBeta, userDenn},
		},
		{
			name: "pagination - 0 page of size 2",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				Page:     0,
				PageSize: 2,
			},
			want: []model.User{userAnna, userBeta},
		},
		{
			name: "pagination - 1st page of size 2",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				Page:     1,
				PageSize: 2,
			},
			want: []model.User{userDenn, userEmel},
		},
		{
			name: "pagination - 2nd page of size 2",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				Page:     2,
				PageSize: 2,
			},
			want: []model.User{userFero},
		},
		{
			name: "pagination - 0 page of negative size",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "first_name",
					Type:  "asc",
				},
				Page:     0,
				PageSize: -5,
			},
			wantErr: true,
		},
		{
			name: "filter & sort & pagination",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "nickname desc",
					Type:  "asc",
				},
				Page:     0,
				PageSize: 2,
				FilterFields: model.FilterFields{
					Country: "Austria",
				},
			},
			want: []model.User{userAnna, userBeta},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			got, err := storage.GetUsers(ctx, tt.params)

			suite.Require().Equal(tt.wantErr, err != nil)
			suite.Assert().Equal(tt.want, got)
		})
	}
}

func (suite *MongoTestSuite) Test_GetUsersDBCallContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	storage := NewMongoUsersStorage(suite.db)
	params := model.GetUsersParams{
		Sort: model.Sort{
			Field: "first_name",
			Type:  "asc",
		},
	}

	got, err := storage.GetUsers(ctx, params)

	suite.Assert().Error(err)
	suite.Assert().Empty(got)
}

func Test_createGetUsersFilter(t *testing.T) {
	tests := []struct {
		name         string
		filterFields model.FilterFields
		want         bson.M
	}{
		{
			name:         "empty",
			filterFields: model.FilterFields{},
			want:         bson.M{},
		},
		{
			name: "first name",
			filterFields: model.FilterFields{
				FirstName: "value",
			},
			want: bson.M{"first_name": "value"},
		},
		{
			name: "last name",
			filterFields: model.FilterFields{
				LastName: "value",
			},
			want: bson.M{"last_name": "value"},
		},
		{
			name: "nickname",
			filterFields: model.FilterFields{
				Nickname: "value",
			},
			want: bson.M{"nickname": "value"},
		},
		{
			name: "email",
			filterFields: model.FilterFields{
				Email: "value",
			},
			want: bson.M{"email": "value"},
		},
		{
			name: "country",
			filterFields: model.FilterFields{
				Country: "value",
			},
			want: bson.M{"country": "value"},
		},
		{
			name: "combination of two",
			filterFields: model.FilterFields{
				Country:  "value",
				Nickname: "value2",
			},
			want: bson.M{"country": "value", "nickname": "value2"},
		},
		{
			name: "combination of all",
			filterFields: model.FilterFields{
				FirstName: "value1",
				LastName:  "value2",
				Nickname:  "value3",
				Email:     "value4",
				Country:   "value5",
			},
			want: bson.M{
				"first_name": "value1",
				"last_name":  "value2",
				"nickname":   "value3",
				"email":      "value4",
				"country":    "value5"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := model.GetUsersParams{
				FilterFields: tt.filterFields,
			}

			got := createGetUsersFilter(p)

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_createGetUsersOpts(t *testing.T) {
	tests := []struct {
		name          string
		params        model.GetUsersParams
		want          *options.FindOptions
		wantErr       bool
		wantErrString string
	}{
		{
			name:          "empty params",
			params:        model.GetUsersParams{},
			wantErr:       true,
			wantErrString: "sort field is required",
		},
		{
			name: "only sort field - default asc sort type",
			params: model.GetUsersParams{
				Sort: model.Sort{Field: "sort_field"},
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", 1}}).
				SetLimit(0).
				SetSkip(0),
		},
		{
			name: "sort field & asc sort type",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "sort_field",
					Type:  "asc",
				},
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", 1}}).
				SetLimit(0).
				SetSkip(0),
		},
		{
			name: "sort field & desc sort type",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "sort_field",
					Type:  "desc",
				},
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", -1}}).
				SetLimit(0).
				SetSkip(0),
		},
		{
			name: "sort field & unknown sort type - defaults to asc",
			params: model.GetUsersParams{
				Sort: model.Sort{
					Field: "sort_field",
					Type:  "unknown",
				},
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", 1}}).
				SetLimit(0).
				SetSkip(0),
		},
		{
			name: "negative page",
			params: model.GetUsersParams{
				Sort: model.Sort{Field: "sort_field"},
				Page: -1,
			},
			wantErr:       true,
			wantErrString: "page cannot be negative number",
		},
		{
			name: "negative page size",
			params: model.GetUsersParams{
				Sort:     model.Sort{Field: "sort_field"},
				PageSize: -1,
			},
			wantErr:       true,
			wantErrString: "page size cannot be negative number",
		},
		{
			name: "page set",
			params: model.GetUsersParams{
				Sort: model.Sort{Field: "sort_field"},
				Page: 5,
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", 1}}).
				SetLimit(0).
				SetSkip(0),
		},
		{
			name: "page size set",
			params: model.GetUsersParams{
				Sort:     model.Sort{Field: "sort_field"},
				PageSize: 5,
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", 1}}).
				SetLimit(5).
				SetSkip(0),
		},
		{
			name: "page & page size set",
			params: model.GetUsersParams{
				Sort:     model.Sort{Field: "sort_field"},
				Page:     2,
				PageSize: 5,
			},
			want: options.Find().
				SetSort(bson.D{{"sort_field", 1}}).
				SetLimit(5).
				SetSkip(10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createGetUsersOpts(tt.params)

			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErrString != "" {
				assert.Equal(t, tt.wantErrString, err.Error())
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
