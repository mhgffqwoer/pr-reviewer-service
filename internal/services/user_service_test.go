package services

import (
	"errors"
	"testing"

	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository for testing
type MockUserRepositoryForUserTest struct {
	mock.Mock
}

func (m *MockUserRepositoryForUserTest) GetByID(userID string) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForUserTest) Exists(userID string) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

func (m *MockUserRepositoryForUserTest) Save(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserTest) GetReview(userID string) ([]*models.PullRequestShort, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PullRequestShort), args.Error(1)
}

func TestUserService_SetActive(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		isActive      bool
		mockSetup     func(repo *MockUserRepositoryForUserTest)
		expectedError error
		expectedUser  *models.User
	}{
		{
			name:     "successful activation",
			userID:   "user1",
			isActive: true,
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: false,
				}, nil)
				repo.On("Save", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: nil,
			expectedUser: &models.User{
				UserID:   "user1",
				Username: "Alice",
				TeamName: "team1",
				IsActive: true,
			},
		},
		{
			name:     "successful deactivation",
			userID:   "user1",
			isActive: false,
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: true,
				}, nil)
				repo.On("Save", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: nil,
			expectedUser: &models.User{
				UserID:   "user1",
				Username: "Alice",
				TeamName: "team1",
				IsActive: false,
			},
		},
		{
			name:     "user not found",
			userID:   "nonexistent",
			isActive: true,
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "nonexistent").Return(false)
			},
			expectedError: ErrNotFound,
			expectedUser:  nil,
		},
		{
			name:     "get user error",
			userID:   "user1",
			isActive: true,
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetByID", "user1").Return((*models.User)(nil), errors.New("database error"))
			},
			expectedError: ErrNotFound,
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockUserRepositoryForUserTest{}
			tt.mockSetup(repo)

			service := NewUserService(repo)
			result, err := service.SetActive(tt.userID, tt.isActive)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedUser.UserID, result.UserID)
				assert.Equal(t, tt.expectedUser.Username, result.Username)
				assert.Equal(t, tt.expectedUser.TeamName, result.TeamName)
				assert.Equal(t, tt.expectedUser.IsActive, result.IsActive)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetReviewe(t *testing.T) {
	tests := []struct {
		name            string
		userID          string
		mockSetup       func(repo *MockUserRepositoryForUserTest)
		expectedError   error
		expectedReviews []*models.PullRequestShort
	}{
		{
			name:   "successful get reviews",
			userID: "user1",
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetReview", "user1").Return([]*models.PullRequestShort{
					{
						PullRequestID:   "pr-1",
						PullRequestName: "Test PR 1",
						AuthorID:        "user2",
					},
					{
						PullRequestID:   "pr-2",
						PullRequestName: "Test PR 2",
						AuthorID:        "user3",
					},
				}, nil)
			},
			expectedError: nil,
			expectedReviews: []*models.PullRequestShort{
				{
					PullRequestID:   "pr-1",
					PullRequestName: "Test PR 1",
					AuthorID:        "user2",
				},
				{
					PullRequestID:   "pr-2",
					PullRequestName: "Test PR 2",
					AuthorID:        "user3",
				},
			},
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "nonexistent").Return(false)
			},
			expectedError:   ErrNotFound,
			expectedReviews: nil,
		},
		{
			name:   "get reviews error",
			userID: "user1",
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetReview", "user1").Return([]*models.PullRequestShort(nil), errors.New("database error"))
			},
			expectedError:   nil,
			expectedReviews: nil,
		},
		{
			name:   "empty reviews list",
			userID: "user1",
			mockSetup: func(repo *MockUserRepositoryForUserTest) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetReview", "user1").Return([]*models.PullRequestShort{}, nil)
			},
			expectedError:   nil,
			expectedReviews: []*models.PullRequestShort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockUserRepositoryForUserTest{}
			tt.mockSetup(repo)

			service := NewUserService(repo)
			result, err := service.GetReviewe(tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedReviews, result)
			}

			repo.AssertExpectations(t)
		})
	}
}
