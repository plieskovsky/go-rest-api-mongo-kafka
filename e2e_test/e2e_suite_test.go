package e2e_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"user-service/e2e_test/test_helpers"
	"user-service/internal/model"
)

type E2ETestSuite struct {
	suite.Suite
	testStart time.Time
	testUser  model.User
}

func Test_RunE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (suite *E2ETestSuite) BeforeTest(_, _ string) {
	err := test_helpers.CleanupMongoUsersCollection()
	suite.Assert().NoError(err, "mongo users collection cleanup")

	// UTC & truncate to millis because that is the Mongo timezone & precision, so it can be used in assertions & test data
	suite.testStart = time.Now().UTC().Truncate(time.Millisecond)

	suite.testUser = model.User{
		ID:        uuid.New(),
		FirstName: "Andrey",
		LastName:  "Amadeus",
		Nickname:  "andrey1",
		Password:  "andreyPWD",
		Email:     "andrey@gmail.com",
		Country:   "FR",
		CreatedAt: suite.testStart,
		UpdatedAt: suite.testStart,
	}
}

func (suite *E2ETestSuite) SetupSuite() {
	err := test_helpers.SetupMongoConnection()
	suite.Require().NoError(err, "mongo connection setup")

	err = test_helpers.SetupKafkaConsumer()
	suite.Require().NoError(err, "kafka consumer connection setup")
}

func (suite *E2ETestSuite) TearDownSuite() {
	err := test_helpers.CloseMongoConnection()
	suite.Assert().NoError(err, "mongo connection close ")

	err = test_helpers.CloseKafkaConsumer()
	suite.Assert().NoError(err, "kafka consumer close ")
}

func (suite *E2ETestSuite) GetTestStart() time.Time {
	return suite.testStart
}

func (suite *E2ETestSuite) GetTestUser() model.User {
	return suite.testUser
}
