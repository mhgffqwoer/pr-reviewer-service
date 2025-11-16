package services

import (
	"errors"

	"github.com/mhgffqwoer/pr-service/internal/models"
)

var (
	ErrTeamExists = errors.New("TEAM_EXISTS")
)

type TeamRepository interface {
	GetByName(teamName string) (*models.Team, error)
	Save(team *models.Team) error
	Exists(teamName string) bool
}

type TeamService struct {
	repo TeamRepository
}

func NewTeamService(repo TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) AddTeam(team *models.Team) (*models.Team, error) {
	if s.repo.Exists(team.TeamName) {
		return nil, ErrTeamExists
	}

	_ = s.repo.Save(team)
	return team, nil
}

func (s *TeamService) GetTeam(teamName string) (*models.Team, error) {
	team, err := s.repo.GetByName(teamName)
	if err != nil {
		return nil, ErrNotFound
	}
	return team, nil
}
