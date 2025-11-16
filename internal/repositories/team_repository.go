package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/mhgffqwoer/pr-service/internal/models"
)

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
		userArgs := map[string]interface{}{
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
