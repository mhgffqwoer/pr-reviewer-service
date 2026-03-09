package services

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/mhgffqwoer/pr-service/internal/logger"
	"github.com/mhgffqwoer/pr-service/internal/models"
)

var (
	ErrPRExists = errors.New("PR_EXISTS")
	ErrNotFound = errors.New("NOT_FOUND")
	ErrPRMerged = errors.New("PR_MERGED")
)

type PullRequestRepository interface {
	Exists(prID string) bool
	GetByID(prID string) (*models.PullRequest, error)
	Save(pr *models.PullRequest) error
	Merge(prID string) error
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
		for i := 0; i < num; i++ {
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
	if err := s.prRepo.Merge(prID); err != nil {
		return nil, err
	}

	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, ErrNotFound
	}

	return pr, nil
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
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
