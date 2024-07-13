package model

// GetUsersParams represent parameters to fetch users list.
type GetUsersParams struct {
	PageSize     int
	Page         int
	Sort         Sort
	FilterFields FilterFields
}

type Sort struct {
	Field string
	Type  string
}

type FilterFields struct {
	FirstName string
	LastName  string
	Nickname  string
	Email     string
	Country   string
}
