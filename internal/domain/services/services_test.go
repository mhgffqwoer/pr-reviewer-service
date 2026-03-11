//go:build unit

package services

import (
	"errors"
	"testing"
	"time"

	"github.com/mhgffqwoer/pr-reviewer-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) Exists(prID string) bool {
	args := m.Called(prID)
	return args.Bool(0)
}

func (m *MockPullRequestRepository) GetByID(prID string) (*models.PullRequest, error) {
	args := m.Called(prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) Save(pr *models.PullRequest) error {
	args := m.Called(pr)
	return args.Error(0)
}

func (m *MockPullRequestRepository) Merge(prID string) error {
	args := m.Called(prID)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(userID string) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Exists(userID string) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

func (m *MockUserRepository) Save(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetReview(userID string) ([]*models.PullRequestShort, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PullRequestShort), args.Error(1)
}

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) GetByName(teamName string) (*models.Team, error) {
	args := m.Called(teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) Exists(teamName string) bool {
	args := m.Called(teamName)
	return args.Bool(0)
}

func (m *MockTeamRepository) Save(team *models.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func TestPullRequestService_Create(t *testing.T) {
	tests := []struct {
		name              string
		prID              string
		prName            string
		authorID          string
		mockSetup         func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository)
		expectedError     error
		expectedReviewers int
	}{
		{
			name:     "successful creation with 2 reviewers",
			prID:     "pr-1",
			prName:   "Test PR",
			authorID: "user1",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("Exists", "pr-1").Return(false)
				userRepo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: true,
				}, nil)
				teamRepo.On("GetByName", "team1").Return(&models.Team{
					TeamName: "team1",
					Members: []models.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
						{UserID: "user2", Username: "Bob", IsActive: true},
						{UserID: "user3", Username: "Charlie", IsActive: true},
						{UserID: "user4", Username: "David", IsActive: true},
					},
				}, nil)
				prRepo.On("Save", mock.AnythingOfType("*models.PullRequest")).Return(nil)
			},
			expectedError:     nil,
			expectedReviewers: 2,
		},
		{
			name:     "successful creation with 1 reviewer",
			prID:     "pr-2",
			prName:   "Test PR 2",
			authorID: "user1",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("Exists", "pr-2").Return(false)
				userRepo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: true,
				}, nil)
				teamRepo.On("GetByName", "team1").Return(&models.Team{
					TeamName: "team1",
					Members: []models.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
						{UserID: "user2", Username: "Bob", IsActive: true},
					},
				}, nil)
				prRepo.On("Save", mock.AnythingOfType("*models.PullRequest")).Return(nil)
			},
			expectedError:     nil,
			expectedReviewers: 1,
		},
		{
			name:     "successful creation with 0 reviewers",
			prID:     "pr-3",
			prName:   "Test PR 3",
			authorID: "user1",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("Exists", "pr-3").Return(false)
				userRepo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: true,
				}, nil)
				teamRepo.On("GetByName", "team1").Return(&models.Team{
					TeamName: "team1",
					Members: []models.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
					},
				}, nil)
				prRepo.On("Save", mock.AnythingOfType("*models.PullRequest")).Return(nil)
			},
			expectedError:     nil,
			expectedReviewers: 0,
		},
		{
			name:     "PR already exists",
			prID:     "pr-1",
			prName:   "Test PR",
			authorID: "user1",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("Exists", "pr-1").Return(true)
			},
			expectedError:     ErrPRExists,
			expectedReviewers: 0,
		},
		{
			name:     "author not found",
			prID:     "pr-1",
			prName:   "Test PR",
			authorID: "nonexistent",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("Exists", "pr-1").Return(false)
				userRepo.On("GetByID", "nonexistent").Return((*models.User)(nil), errors.New("user not found"))
			},
			expectedError:     ErrNotFound,
			expectedReviewers: 0,
		},
		{
			name:     "team not found",
			prID:     "pr-1",
			prName:   "Test PR",
			authorID: "user1",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("Exists", "pr-1").Return(false)
				userRepo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "nonexistent",
					IsActive: true,
				}, nil)
				teamRepo.On("GetByName", "nonexistent").Return((*models.Team)(nil), errors.New("team not found"))
			},
			expectedError:     ErrNotFound,
			expectedReviewers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &MockPullRequestRepository{}
			userRepo := &MockUserRepository{}
			teamRepo := &MockTeamRepository{}

			tt.mockSetup(prRepo, userRepo, teamRepo)

			service := NewPullRequestService(prRepo, userRepo, teamRepo)
			result, err := service.Create(tt.prID, tt.prName, tt.authorID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.prID, result.PullRequestID)
				assert.Equal(t, tt.prName, result.PullRequestName)
				assert.Equal(t, tt.authorID, result.AuthorID)
				assert.Equal(t, models.StatusOpen, result.Status)
				assert.Len(t, result.AssignedReviewers, tt.expectedReviewers)
				assert.NotNil(t, result.CreatedAt)
				assert.Nil(t, result.MergedAt)
			}

			prRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestPullRequestService_Merge(t *testing.T) {
	tests := []struct {
		name          string
		prID          string
		mockSetup     func(prRepo *MockPullRequestRepository)
		expectedError error
	}{
		{
			name: "successful merge",
			prID: "pr-1",
			mockSetup: func(prRepo *MockPullRequestRepository) {
				prRepo.On("GetByID", "pr-1").Return(&models.PullRequest{
					PullRequestID:     "pr-1",
					PullRequestName:   "Test PR",
					AuthorID:          "user1",
					Status:            models.StatusOpen,
					AssignedReviewers: []string{"user2", "user3"},
					CreatedAt:         &time.Time{},
					MergedAt:          nil,
				}, nil)
				prRepo.On("Merge", "pr-1").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "PR not found",
			prID: "pr-1",
			mockSetup: func(prRepo *MockPullRequestRepository) {
				prRepo.On("GetByID", "pr-1").Return((*models.PullRequest)(nil), ErrNotFound)
			},
			expectedError: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &MockPullRequestRepository{}
			userRepo := &MockUserRepository{}
			teamRepo := &MockTeamRepository{}

			tt.mockSetup(prRepo)

			service := NewPullRequestService(prRepo, userRepo, teamRepo)
			result, err := service.Merge(tt.prID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, models.StatusMerged, result.Status)
			}

			prRepo.AssertExpectations(t)
		})
	}
}

func TestPullRequestService_Reassign(t *testing.T) {
	tests := []struct {
		name                string
		prID                string
		oldReviewerID       string
		mockSetup           func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository)
		expectedError       error
		expectedNewReviewer string
	}{
		{
			name:          "successful reassignment",
			prID:          "pr-1",
			oldReviewerID: "user2",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("GetByID", "pr-1").Return(&models.PullRequest{
					PullRequestID:     "pr-1",
					PullRequestName:   "Test PR",
					AuthorID:          "user1",
					Status:            models.StatusOpen,
					AssignedReviewers: []string{"user2", "user3"},
					CreatedAt:         &time.Time{},
					MergedAt:          nil,
				}, nil)
				userRepo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: true,
				}, nil)
				teamRepo.On("GetByName", "team1").Return(&models.Team{
					TeamName: "team1",
					Members: []models.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
						{UserID: "user2", Username: "Bob", IsActive: true},
						{UserID: "user3", Username: "Charlie", IsActive: true},
						{UserID: "user4", Username: "David", IsActive: true},
					},
				}, nil)
				prRepo.On("Save", mock.AnythingOfType("*models.PullRequest")).Return(nil)
			},
			expectedError:       nil,
			expectedNewReviewer: "user4",
		},
		{
			name:          "PR not found",
			prID:          "pr-1",
			oldReviewerID: "user2",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("GetByID", "pr-1").Return((*models.PullRequest)(nil), errors.New("PR not found"))
			},
			expectedError: ErrNotFound,
		},
		{
			name:          "PR already merged",
			prID:          "pr-1",
			oldReviewerID: "user2",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("GetByID", "pr-1").Return(&models.PullRequest{
					PullRequestID:     "pr-1",
					PullRequestName:   "Test PR",
					AuthorID:          "user1",
					Status:            models.StatusMerged,
					AssignedReviewers: []string{"user2", "user3"},
					CreatedAt:         &time.Time{},
					MergedAt:          &time.Time{},
				}, nil)
			},
			expectedError: ErrPRMerged,
		},
		{
			name:          "reviewer not assigned",
			prID:          "pr-1",
			oldReviewerID: "user5",
			mockSetup: func(prRepo *MockPullRequestRepository, userRepo *MockUserRepository, teamRepo *MockTeamRepository) {
				prRepo.On("GetByID", "pr-1").Return(&models.PullRequest{
					PullRequestID:     "pr-1",
					PullRequestName:   "Test PR",
					AuthorID:          "user1",
					Status:            models.StatusOpen,
					AssignedReviewers: []string{"user2", "user3"},
					CreatedAt:         &time.Time{},
					MergedAt:          nil,
				}, nil)
				userRepo.On("GetByID", "user1").Return(&models.User{
					UserID:   "user1",
					Username: "Alice",
					TeamName: "team1",
					IsActive: true,
				}, nil)
				teamRepo.On("GetByName", "team1").Return(&models.Team{
					TeamName: "team1",
					Members: []models.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
						{UserID: "user2", Username: "Bob", IsActive: true},
						{UserID: "user3", Username: "Charlie", IsActive: true},
						{UserID: "user4", Username: "David", IsActive: true},
					},
				}, nil)
			},
			expectedError: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := &MockPullRequestRepository{}
			userRepo := &MockUserRepository{}
			teamRepo := &MockTeamRepository{}

			tt.mockSetup(prRepo, userRepo, teamRepo)

			service := NewPullRequestService(prRepo, userRepo, teamRepo)
			result, newReviewer, err := service.Reassign(tt.prID, tt.oldReviewerID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedNewReviewer, newReviewer)
			}

			prRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			teamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamService_AddTeam(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		mockSetup     func(repo *MockTeamRepository)
		expectedError error
		expectedTeam  *models.Team
	}{
		{
			name:     "successful team addition",
			teamName: "team1",
			mockSetup: func(repo *MockTeamRepository) {
				repo.On("Exists", "team1").Return(false)
				repo.On("Save", mock.AnythingOfType("*models.Team")).Return(nil)
			},
			expectedError: nil,
			expectedTeam: &models.Team{
				TeamName: "team1",
				Members:  nil,
			},
		},
		{
			name:     "team already exists",
			teamName: "team1",
			mockSetup: func(repo *MockTeamRepository) {
				repo.On("Exists", "team1").Return(true)
			},
			expectedError: ErrTeamExists,
			expectedTeam:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockTeamRepository{}
			tt.mockSetup(repo)

			service := NewTeamService(repo)
			result, err := service.AddTeam(&models.Team{TeamName: tt.teamName})

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTeam.TeamName, result.TeamName)
				assert.Equal(t, tt.expectedTeam.Members, result.Members)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestTeamService_GetTeam(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		mockSetup     func(repo *MockTeamRepository)
		expectedError error
		expectedTeam  *models.Team
	}{
		{
			name:     "successful get team",
			teamName: "team1",
			mockSetup: func(repo *MockTeamRepository) {
				repo.On("GetByName", "team1").Return(&models.Team{
					TeamName: "team1",
					Members: []models.TeamMember{
						{UserID: "user1", Username: "Alice", IsActive: true},
						{UserID: "user2", Username: "Bob", IsActive: true},
					},
				}, nil)
			},
			expectedError: nil,
			expectedTeam: &models.Team{
				TeamName: "team1",
				Members: []models.TeamMember{
					{UserID: "user1", Username: "Alice", IsActive: true},
					{UserID: "user2", Username: "Bob", IsActive: true},
				},
			},
		},
		{
			name:     "team not found",
			teamName: "nonexistent",
			mockSetup: func(repo *MockTeamRepository) {
				repo.On("GetByName", "nonexistent").Return((*models.Team)(nil), errors.New("team not found"))
			},
			expectedError: ErrNotFound,
			expectedTeam:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockTeamRepository{}
			tt.mockSetup(repo)

			service := NewTeamService(repo)
			result, err := service.GetTeam(tt.teamName)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedTeam.TeamName, result.TeamName)
				assert.Equal(t, tt.expectedTeam.Members, result.Members)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestUserService_SetActive(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		isActive      bool
		mockSetup     func(repo *MockUserRepository)
		expectedError error
		expectedUser  *models.User
	}{
		{
			name:     "successful activation",
			userID:   "user1",
			isActive: true,
			mockSetup: func(repo *MockUserRepository) {
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
			mockSetup: func(repo *MockUserRepository) {
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
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Exists", "nonexistent").Return(false)
			},
			expectedError: ErrNotFound,
			expectedUser:  nil,
		},
		{
			name:     "get user error",
			userID:   "user1",
			isActive: true,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetByID", "user1").Return((*models.User)(nil), errors.New("database error"))
			},
			expectedError: ErrNotFound,
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockUserRepository{}
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
		mockSetup       func(repo *MockUserRepository)
		expectedError   error
		expectedReviews []*models.PullRequestShort
	}{
		{
			name:   "successful get reviews",
			userID: "user1",
			mockSetup: func(repo *MockUserRepository) {
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
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Exists", "nonexistent").Return(false)
			},
			expectedError:   ErrNotFound,
			expectedReviews: nil,
		},
		{
			name:   "get reviews error",
			userID: "user1",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetReview", "user1").Return([]*models.PullRequestShort(nil), errors.New("database error"))
			},
			expectedError:   nil,
			expectedReviews: nil,
		},
		{
			name:   "empty reviews list",
			userID: "user1",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Exists", "user1").Return(true)
				repo.On("GetReview", "user1").Return([]*models.PullRequestShort{}, nil)
			},
			expectedError:   nil,
			expectedReviews: []*models.PullRequestShort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockUserRepository{}
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
