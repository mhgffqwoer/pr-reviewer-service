package services

import (
	"errors"
	"testing"
	"time"

	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories for testing
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

// Mock interfaces for testing
type MockPullRequestRepositoryInterface interface {
	PullRequestRepository
	mock.TestingT
}

type MockUserRepositoryInterface interface {
	UserRepository
	mock.TestingT
}

type MockTeamRepositoryInterface interface {
	TeamRepository
	mock.TestingT
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
				prRepo.On("Merge", "pr-1").Return(nil)
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
			expectedError: nil,
		},
		{
			name: "PR not found",
			prID: "pr-1",
			mockSetup: func(prRepo *MockPullRequestRepository) {
				prRepo.On("Merge", "pr-1").Return(ErrNotFound)
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
