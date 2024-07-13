package controller

import (
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"user-service/internal/model"
)

var supportedSortFields = map[string]struct{}{
	"last_name":  {},
	"first_name": {},
	"nickname":   {},
	"password":   {},
	"email":      {},
	"country":    {},
	"created_at": {},
	"updated_at": {},
}

const (
	userIDPathParam = "userID"
	defaultPageSize = 20
	defaultPage     = 0
)

func parseGetUsersParams(c *gin.Context) (*model.GetUsersParams, error) {
	pageSize := defaultPageSize
	page := defaultPage
	sort := model.Sort{
		Field: "last_name",
		Type:  "asc",
	}

	if got, ok := c.GetQuery("pageSize"); ok {
		parsed, err := strconv.Atoi(got)
		if err != nil {
			return nil, errors.New("pageSize query parameter has to be a number")
		}
		if parsed < 0 {
			return nil, errors.New("pageSize query parameter has to be a positive number")
		}
		pageSize = parsed
	}

	if got, ok := c.GetQuery("page"); ok {
		parsed, err := strconv.Atoi(got)
		if err != nil {
			return nil, errors.New("page query parameter has to be a number")
		}
		if parsed < 0 {
			return nil, errors.New("page query parameter has to be a positive number")
		}
		page = parsed
	}

	if got, ok := c.GetQuery("sortBy"); ok {
		parsed, err := parseSortBy(got)
		if err != nil {
			return nil, err
		}
		sort = *parsed
	}

	return &model.GetUsersParams{
		PageSize:     pageSize,
		Page:         page,
		Sort:         sort,
		FilterFields: parseFilterFields(c),
	}, nil
}

func parseSortBy(sortBy string) (*model.Sort, error) {
	sortBy = strings.ToLower(sortBy)
	parts := strings.Split(sortBy, ".")

	if len(parts) != 2 {
		return nil, errors.New("invalid sortBy query parameter format")
	}

	if _, ok := supportedSortFields[parts[0]]; !ok {
		return nil, errors.New("unsupported sorting field")
	}

	if parts[1] != "asc" && parts[1] != "desc" {
		return nil, errors.New("invalid sorting type")
	}

	return &model.Sort{
		Field: parts[0],
		Type:  parts[1],
	}, nil
}

func parseFilterFields(c *gin.Context) model.FilterFields {
	filter := model.FilterFields{}

	if v, ok := c.GetQuery("first_name"); ok {
		filter.FirstName = v
	}
	if v, ok := c.GetQuery("last_name"); ok {
		filter.LastName = v
	}
	if v, ok := c.GetQuery("nickname"); ok {
		filter.Nickname = v
	}
	if v, ok := c.GetQuery("email"); ok {
		filter.Email = v
	}
	if v, ok := c.GetQuery("country"); ok {
		filter.Country = v
	}

	return filter
}
