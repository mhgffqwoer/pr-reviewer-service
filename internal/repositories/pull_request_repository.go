package repositories

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/mhgffqwoer/pr-service/internal/models"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) Exists(prID string) bool {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM pull_requests 
			WHERE pull_request_id = $1
		)
	`
	err := r.db.Get(&exists, query, prID)
	if err != nil {
		return false
	}
	return exists
}

func (r *PullRequestRepository) Save(pr *models.PullRequest) error {
	reviewersJSON, err := json.Marshal(pr.AssignedReviewers)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers, created_at)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, NOW()))
		ON CONFLICT (pull_request_id) DO UPDATE SET
			pull_request_name = EXCLUDED.pull_request_name,
			status = EXCLUDED.status,
			assigned_reviewers = EXCLUDED.assigned_reviewers,
			created_at = EXCLUDED.created_at
	`
	_, err = r.db.Exec(query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status.String(), reviewersJSON, pr.CreatedAt)
	return err
}

func (r *PullRequestRepository) GetByID(prID string) (*models.PullRequest, error) {
	pr := &models.PullRequest{}

	query := `
		SELECT 
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			assigned_reviewers,
			created_at,
			merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var reviewersJSON []byte
	err := r.db.QueryRowx(query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&reviewersJSON,
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(reviewersJSON) > 0 {
		if err := json.Unmarshal(reviewersJSON, &pr.AssignedReviewers); err != nil {
			return nil, err
		}
	}

	return pr, nil
}

func (r *PullRequestRepository) Merge(prID string) error {
	query := `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = COALESCE(merged_at, NOW())
		WHERE pull_request_id = $1
	`
	_, err := r.db.Exec(query, prID)
	return err
}