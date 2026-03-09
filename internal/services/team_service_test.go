package services

import (
	"errors"
	"testing"

	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTeamRepository for testing
type MockTeamRepositoryForTeamTest struct {
	mock.Mock
}

func (m *MockTeamRepositoryForTeamTest) GetByName(teamName string) (*models.Team, error) {
	args := m.Called(teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepositoryForTeamTest) Exists(teamName string) bool {
	args := m.Called(teamName)
	return args.Bool(0)
}

func (m *MockTeamRepositoryForTeamTest) Save(team *models.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func TestTeamService_AddTeam(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		mockSetup     func(repo *MockTeamRepositoryForTeamTest)
		expectedError error
		expectedTeam  *models.Team
	}{
		{
			name:     "successful team addition",
			teamName: "team1",
			mockSetup: func(repo *MockTeamRepositoryForTeamTest) {
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
			mockSetup: func(repo *MockTeamRepositoryForTeamTest) {
				repo.On("Exists", "team1").Return(true)
			},
			expectedError: ErrTeamExists,
			expectedTeam:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockTeamRepositoryForTeamTest{}
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
		mockSetup     func(repo *MockTeamRepositoryForTeamTest)
		expectedError error
		expectedTeam  *models.Team
	}{
		{
			name:     "successful get team",
			teamName: "team1",
			mockSetup: func(repo *MockTeamRepositoryForTeamTest) {
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
			mockSetup: func(repo *MockTeamRepositoryForTeamTest) {
				repo.On("GetByName", "nonexistent").Return((*models.Team)(nil), errors.New("team not found"))
			},
			expectedError: ErrNotFound,
			expectedTeam:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockTeamRepositoryForTeamTest{}
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
