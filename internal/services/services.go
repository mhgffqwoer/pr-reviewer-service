package services

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/mhgffqwoer/pr-service/internal/logger"
	"github.com/mhgffqwoer/pr-service/internal/models"
)

var (
	ErrPRExists   = errors.New("PR_EXISTS")
	ErrNotFound   = errors.New("NOT_FOUND")
	ErrTeamExists = errors.New("TEAM_EXISTS")
	ErrPRMerged   = errors.New("PR_MERGED")
)

type Service struct {
	TeamService        *TeamService
	UserService        *UserService
	PullRequestService *PullRequestService
}

func NewService(prRepo PullRequestRepository, userRepo UserRepository, teamRepo TeamRepository) *Service {
	return &Service{
		TeamService:        NewTeamService(teamRepo),
		UserService:        NewUserService(userRepo),
		PullRequestService: NewPullRequestService(prRepo, userRepo, teamRepo),
	}
}

type PullRequestRepository interface {
	Exists(prID string) bool
	GetByID(prID string) (*models.PullRequest, error)
	Save(pr *models.PullRequest) error
	Merge(prID string) error
}

type TeamRepository interface {
	GetByName(teamName string) (*models.Team, error)
	Save(team *models.Team) error
	Exists(teamName string) bool
}

type UserRepository interface {
	GetByID(userID string) (*models.User, error)
	Exists(userID string) bool
	Save(user *models.User) error
	GetReview(userID string) ([]*models.PullRequestShort, error)
}

type PullRequestService struct {
	prRepo   PullRequestRepository
	userRepo UserRepository
	teamRepo TeamRepository
}

func NewPullRequestService(repo PullRequestRepository, userRepo UserRepository, teamRepo TeamRepository) *PullRequestService {
	return &PullRequestService{
		prRepo:   repo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *PullRequestService) Create(prID, prName, authorID string) (*models.PullRequest, error) {
	logger.Get().Debugw(fmt.Sprintf("DEBUG: s = %p, prRepo = %p, userRepo = %p, teamRepo = %p\n", s, s.prRepo, s.userRepo, s.teamRepo))
	if s.prRepo.Exists(prID) {
		return nil, fmt.Errorf("%w: PR %s already exists", ErrPRExists, prID)
	}

	author, err := s.userRepo.GetByID(authorID)
	if err != nil {
		return nil, fmt.Errorf("%w: resource not found", ErrNotFound)
	}
	if author == nil {
		return nil, fmt.Errorf("%w: resource not found", ErrNotFound)
	}

	team, err := s.teamRepo.GetByName(author.TeamName)
	if err != nil {
		return nil, fmt.Errorf("%w: resource not found", ErrNotFound)
	}

	var candidates []string
	for _, member := range team.Members {
		if member.UserID != authorID && member.IsActive {
			candidates = append(candidates, member.UserID)
		}
	}

	num := min(len(candidates), 2)
	var reviewers []string
	if num > 0 {
		reviewers = make([]string, num)
		perm := rand.Perm(len(candidates))
		for i := range num {
			reviewers[i] = candidates[perm[i]]
		}
	}

	now := time.Now()
	pr := &models.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            models.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         &now,
	}

	_ = s.prRepo.Save(pr)
	return pr, nil
}

func (s *PullRequestService) Merge(prID string) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == models.StatusMerged {
		return nil, ErrPRMerged
	}

	err = s.prRepo.Merge(prID)
	if err != nil {
		return nil, err
	}

	// Update the status to merged
	pr.Status = models.StatusMerged
	now := time.Now()
	pr.MergedAt = &now

	return pr, nil
}

func (s *PullRequestService) GetByID(prID string) (*models.PullRequest, error) {
	return s.prRepo.GetByID(prID)
}

func (s *PullRequestService) Reassign(prID, oldReviewerID string) (*models.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, "", ErrNotFound
	}

	if pr.Status != models.StatusOpen {
		return nil, "", ErrPRMerged
	}

	author, err := s.userRepo.GetByID(pr.AuthorID)
	if err != nil {
		return nil, "", ErrNotFound
	}
	team, err := s.teamRepo.GetByName(author.TeamName)
	if err != nil {
		return nil, "", ErrNotFound
	}

	found := false
	newReviewers := make([]string, 0, len(pr.AssignedReviewers))
	for _, r := range pr.AssignedReviewers {
		if r == oldReviewerID {
			found = true
			continue
		}
		newReviewers = append(newReviewers, r)
	}
	if !found {
		return nil, "", ErrNotFound
	}

	candidates := make([]string, 0)
	for _, member := range team.Members {
		if member.UserID == author.UserID {
			continue
		}
		if !member.IsActive {
			continue
		}
		if contains(newReviewers, member.UserID) {
			continue
		}
		if member.UserID == oldReviewerID {
			continue
		}
		candidates = append(candidates, member.UserID)
	}
	newReviewer := ""
	if len(candidates) > 0 {
		perm := rand.Perm(len(candidates))
		newReviewer = candidates[perm[0]]
		newReviewers = append(newReviewers, newReviewer)
	}

	pr.AssignedReviewers = newReviewers
	_ = s.prRepo.Save(pr)
	return pr, newReviewer, nil
}

func contains(list []string, v string) bool {
	return slices.Contains(list, v)
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

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) SetActive(userID string, isActive bool) (*models.User, error) {
	if !s.repo.Exists(userID) {
		return nil, ErrNotFound
	}

	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, ErrNotFound
	}
	user.IsActive = isActive
	_ = s.repo.Save(user)

	return user, nil
}

func (s *UserService) GetReviewe(userID string) ([]*models.PullRequestShort, error) {
	if !s.repo.Exists(userID) {
		return nil, ErrNotFound
	}

	reviews, _ := s.repo.GetReview(userID)
	return reviews, nil
}
