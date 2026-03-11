package postgres

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/mhgffqwoer/pr-reviewer-service/internal/domain/models"
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

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Exists(teamName string) bool {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM teams 
			WHERE name = $1
		)
	`
	err := r.db.Get(&exists, query, teamName)
	if err != nil {
		return false
	}
	return exists
}

func (r *TeamRepository) GetByName(teamName string) (*models.Team, error) {
	team := &models.Team{}

	getTeamQuery := `
		SELECT name 
		FROM teams 
		WHERE name = $1
	`
	err := r.db.Get(&team.TeamName, getTeamQuery, teamName)
	if err != nil {
		return nil, err
	}

	getMembersQuery := `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`
	err = r.db.Select(&team.Members, getMembersQuery, teamName)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *TeamRepository) Save(team *models.Team) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	upsertTeam := `
        INSERT INTO teams (name)
        VALUES ($1)
        ON CONFLICT (name) DO NOTHING
    `
	_, err = tx.Exec(upsertTeam, team.TeamName)
	if err != nil {
		return err
	}

	insertUser := `
        INSERT INTO users (user_id, username, team_name, is_active)
        VALUES (:user_id, :username, :team_name, :is_active)
        ON CONFLICT (user_id) DO UPDATE SET
            username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active
    `
	for _, member := range team.Members {
		userArgs := map[string]any{
			"user_id":   member.UserID,
			"username":  member.Username,
			"team_name": team.TeamName,
			"is_active": member.IsActive,
		}
		_, err = tx.NamedExec(insertUser, userArgs)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

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
