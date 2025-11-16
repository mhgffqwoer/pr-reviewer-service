package repositories

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/mhgffqwoer/pr-service/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Exists(userID string) bool {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM users 
			WHERE user_id = $1
		)
	`
	err := r.db.Get(&exists, query, userID)
	if err != nil {
		return false
	}
	return exists
}

func (r *UserRepository) GetByID(userID string) (*models.User, error) {
	user := &models.User{}

	getUserQuery := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`
	err := r.db.Get(user, getUserQuery, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Save(user *models.User) error {
	upsertQuery := `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username,
		    team_name = EXCLUDED.team_name,
		    is_active = EXCLUDED.is_active
	`
	_, err := r.db.Exec(upsertQuery, user.UserID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *UserRepository) GetReview(userID string) ([]*models.PullRequestShort, error) {
	var prs []*models.PullRequestShort
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status
		FROM pull_requests
		WHERE assigned_reviewers @> $1::jsonb
	`
	userArray := []string{userID}
	userJSON, err := json.Marshal(userArray)
	if err != nil {
		return nil, err
	}

	err = r.db.Select(&prs, query, userJSON)
	if err != nil {
		return nil, err
	}

	return prs, nil
}
