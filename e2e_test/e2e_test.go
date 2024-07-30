package e2e_test

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"user-service/e2e_test/test_helpers"
	"user-service/internal/model"
)

// E2E tests that cover the Happy path of each REST endpoint + unhappy paths of some endpoints.
// In a real world project would cover each endpoint with both happy & unhappy + would add tests
// for some generic REST failures. Not writing them now as I believe the existing ones should be
// sufficient to showcase the way to write and structure them.

func (suite *E2ETestSuite) Test_CreateUser_HappyPath() {
	require := suite.Require()
	assert := suite.Assert()

	resp, responseCode := test_helpers.CallCreateUserEndpoint(suite.T(), suite.GetTestUser())
	require.Equal(http.StatusCreated, responseCode)

	var gotUser model.User
	err := json.Unmarshal(resp, &gotUser)
	require.NoError(err, "failed to unmarshal response body")

	// validate http response
	assert.Equal(suite.GetTestUser().FirstName, gotUser.FirstName)
	assert.Equal(suite.GetTestUser().LastName, gotUser.LastName)
	assert.Equal(suite.GetTestUser().Nickname, gotUser.Nickname)
	assert.Equal(suite.GetTestUser().Email, gotUser.Email)
	assert.Equal(suite.GetTestUser().Password, gotUser.Password)
	assert.Equal(suite.GetTestUser().Country, gotUser.Country)
	assert.NotEqual(suite.GetTestUser().ID, gotUser.ID)
	assert.NotEmpty(gotUser.ID)
	assert.True(gotUser.CreatedAt.After(suite.GetTestStart()))
	assert.True(gotUser.UpdatedAt.After(suite.GetTestStart()))

	// validate db user
	dbUser := test_helpers.GetUserFromDB(suite.T(), gotUser.ID)
	assert.Equal(gotUser, dbUser)

	// validate kafka event
	event := test_helpers.GetKafkaCreateOrUpdateEvent(suite.T())
	assert.EqualValues(model.USER_CREATED, event.Action)
	assert.Equal(gotUser, event.UserData)
}

func (suite *E2ETestSuite) Test_CreateUser_Invalid_Payload() {
	require := suite.Require()
	assert := suite.Assert()

	invalidUser := suite.GetTestUser()
	invalidUser.FirstName = ""

	resp, responseCode := test_helpers.CallCreateUserEndpoint(suite.T(), invalidUser)
	require.Equal(http.StatusBadRequest, responseCode)

	var errResp test_helpers.ErrResponse
	err := json.Unmarshal(resp, &errResp)
	require.NoError(err, "failed to unmarshal response body")
	assert.Equal("first name is required", errResp.Error)

	// validate db
	test_helpers.AssertUsersDBCollectionIsEmpty(suite.T())

	// validate kafka event
	test_helpers.AssertNoUserEventPublishedToKafka(suite.T())
}

func (suite *E2ETestSuite) Test_UpdateUser_HappyPath() {
	require := suite.Require()
	assert := suite.Assert()
	origUser := suite.GetTestUser()

	test_helpers.CreateUserInDB(suite.T(), origUser)

	updateUser := origUser
	updateUser.FirstName = "difFirst"
	updateUser.LastName = "difLast"
	updateUser.Nickname = "difNick"
	updateUser.Country = "difCount"
	updateUser.Email = "difEmail@gmail.com"
	updateUser.Password = "difPassword"

	resp, responseCode := test_helpers.CallUpdateUserEndpoint(suite.T(), updateUser)
	require.Equal(http.StatusNoContent, responseCode)
	require.Empty(resp)

	// validate db user
	gotDBUser := test_helpers.GetUserFromDB(suite.T(), origUser.ID)

	// validate http response
	assert.Equal(updateUser.FirstName, gotDBUser.FirstName)
	assert.Equal(updateUser.LastName, gotDBUser.LastName)
	assert.Equal(updateUser.Nickname, gotDBUser.Nickname)
	assert.Equal(updateUser.Email, gotDBUser.Email)
	assert.Equal(updateUser.Password, gotDBUser.Password)
	assert.Equal(updateUser.Country, gotDBUser.Country)
	assert.Equal(updateUser.ID, gotDBUser.ID)
	assert.Equal(origUser.CreatedAt, gotDBUser.CreatedAt)
	assert.True(gotDBUser.UpdatedAt.After(origUser.UpdatedAt))

	// validate kafka event
	event := test_helpers.GetKafkaCreateOrUpdateEvent(suite.T())
	assert.EqualValues(model.USER_UPDATED, event.Action)
	assert.Equal(gotDBUser, event.UserData)
}

func (suite *E2ETestSuite) Test_UpdateUser_NonExistent() {
	require := suite.Require()
	assert := suite.Assert()

	updateUser := suite.GetTestUser()
	updateUser.FirstName = "difFirst"
	updateUser.LastName = "difLast"
	updateUser.Nickname = "difNick"
	updateUser.Country = "difCount"
	updateUser.Email = "difEmail@gmail.com"
	updateUser.Password = "difPassword"

	resp, responseCode := test_helpers.CallUpdateUserEndpoint(suite.T(), updateUser)
	require.Equal(http.StatusNotFound, responseCode)

	var errResp test_helpers.ErrResponse
	err := json.Unmarshal(resp, &errResp)
	require.NoError(err, "failed to unmarshal response body")
	assert.Equal("user not found", errResp.Error)

	// validate db
	test_helpers.AssertUsersDBCollectionIsEmpty(suite.T())

	// validate kafka event
	test_helpers.AssertNoUserEventPublishedToKafka(suite.T())
}

func (suite *E2ETestSuite) Test_DeleteUser_HappyPath() {
	require := suite.Require()
	assert := suite.Assert()
	origUser := suite.GetTestUser()

	test_helpers.CreateUserInDB(suite.T(), origUser)

	resp, responseCode := test_helpers.CallDeleteUserEndpoint(suite.T(), origUser.ID)
	require.Equal(http.StatusNoContent, responseCode)
	require.Empty(resp)

	// validate db
	test_helpers.AssertUsersDBCollectionIsEmpty(suite.T())

	// validate kafka event
	event := test_helpers.GetKafkaDeletedEvent(suite.T())
	assert.EqualValues(model.USER_DELETED, event.Action)
	assert.Equal(origUser.ID.String(), event.UserData.ID)
}

func (suite *E2ETestSuite) Test_GetUser_HappyPath() {
	require := suite.Require()
	assert := suite.Assert()
	origUser := suite.GetTestUser()

	test_helpers.CreateUserInDB(suite.T(), origUser)

	resp, responseCode := test_helpers.CallGetUserEndpoint(suite.T(), origUser.ID)
	require.Equal(http.StatusOK, responseCode)

	var gotUser model.User
	err := json.Unmarshal(resp, &gotUser)
	require.NoError(err, "failed to unmarshal response body")
	assert.Equal(origUser, gotUser)

	// validate kafka event
	test_helpers.AssertNoUserEventPublishedToKafka(suite.T())
}

func (suite *E2ETestSuite) Test_GetUsers_HappyPath() {
	require := suite.Require()
	assert := suite.Assert()

	user1 := suite.GetTestUser()
	user1.ID = uuid.New()
	user1.Nickname = "anna"
	user1.Country = "UK"

	user2 := suite.GetTestUser()
	user2.ID = uuid.New()
	user2.Nickname = "beta"
	user2.Country = "CZ"

	user3 := suite.GetTestUser()
	user3.ID = uuid.New()
	user3.Nickname = "felipe"
	user3.Country = "CZ"

	user4 := suite.GetTestUser()
	user4.ID = uuid.New()
	user4.Nickname = "kendra"
	user4.Country = "CZ"

	user5 := suite.GetTestUser()
	user5.ID = uuid.New()
	user5.Nickname = "xena"
	user5.Country = "CZ"

	test_helpers.CreateUsersInDB(suite.T(), user1, user2, user3, user4, user5)

	resp, responseCode := test_helpers.CallPath(suite.T(), http.MethodGet, "/v1/users?country=CZ&sortBy=nickname.asc&page=1&pageSize=2")
	require.Equal(http.StatusOK, responseCode)

	var gotUsers []model.User
	err := json.Unmarshal(resp, &gotUsers)
	require.NoError(err, "failed to unmarshal response body")
	assert.Equal([]model.User{user4, user5}, gotUsers)

	// validate kafka event
	test_helpers.AssertNoUserEventPublishedToKafka(suite.T())
}

func (suite *E2ETestSuite) Test_CallNonExistingPath() {
	require := suite.Require()

	resp, responseCode := test_helpers.CallPath(suite.T(), http.MethodGet, "/non/existing/path")
	require.Equal(http.StatusNotFound, responseCode)
	require.Equal("404 page not found", string(resp))

	// validate db
	test_helpers.AssertUsersDBCollectionIsEmpty(suite.T())

	// validate kafka event
	test_helpers.AssertNoUserEventPublishedToKafka(suite.T())
}
