//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	httpapi "github.com/mhgffqwoer/pr-service/internal/adapters/http"
	"github.com/mhgffqwoer/pr-service/internal/adapters/postgres"
	"github.com/mhgffqwoer/pr-service/internal/domain/models"
	"github.com/mhgffqwoer/pr-service/internal/domain/services"
)

type PRServiceTestSuite struct {
	suite.Suite
	db      *sqlx.DB
	handler *httpapi.Handlers
	service *services.Service
}

func (s *PRServiceTestSuite) SetupSuite() {
	dsn := "postgres://testuser:testpass@localhost:5432/pr_service_test?sslmode=disable"

	// Повторные попытки подключения к базе данных
	var db *sqlx.DB
	var err error
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := range maxRetries {
		db, err = sqlx.Connect("postgres", dsn)
		if err == nil {
			break
		}
		if i == maxRetries-1 {
			s.T().Fatalf("Failed to connect to test database after %d attempts: %v", maxRetries, err)
		}
		time.Sleep(retryInterval)
	}

	s.db = db

	prRepo := postgres.NewPullRequestRepository(db)
	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)

	s.service = services.NewService(prRepo, userRepo, teamRepo)
	s.handler = httpapi.NewHandlers(s.service)
}

func (s *PRServiceTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE pull_requests, users, teams CASCADE")
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PRServiceTestSuite))
}

func (s *PRServiceTestSuite) TestCreatePR_Success() {

	s.db.Exec("INSERT INTO teams (name) VALUES ('Backend')")
	s.db.Exec(`
		INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES 
		('user_1', 'Alice', 'Backend', true),
		('user_2', 'Bob', 'Backend', true),
		('user_3', 'Charlie', 'Backend', true)`)

	reqBody, _ := json.Marshal(map[string]string{
		"pull_request_id":   "PR-123",
		"pull_request_name": "feat: add auth",
		"author_id":         "user_1",
	})

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	s.handler.CreatePR(rr, req)

	assert.Equal(s.T(), http.StatusCreated, rr.Code)

	var response struct {
		PR models.PullRequest `json:"pull_request"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "PR-123", response.PR.PullRequestID)
	assert.Len(s.T(), response.PR.AssignedReviewers, 2)
}

func (s *PRServiceTestSuite) TestReassignPR() {

	s.db.Exec("INSERT INTO teams (name) VALUES ('DevOps')")
	s.db.Exec(`INSERT INTO users (user_id, username, team_name, is_active) VALUES 
		('u1', 'Author', 'DevOps', true), 
		('u2', 'Reviewer1', 'DevOps', true), 
		('u3', 'Reviewer2', 'DevOps', true),
		('u4', 'Reviewer3', 'DevOps', true)`)

	pr, err := s.service.PullRequestService.Create("PR-REASSIGN-ID", "feat: test reassign", "u1")
	assert.NoError(s.T(), err)
	assert.Len(s.T(), pr.AssignedReviewers, 2, "Должно быть назначено 2 ревьюера")

	oldReviewer := pr.AssignedReviewers[0]

	reqBody, _ := json.Marshal(map[string]string{
		"pull_request_id": "PR-REASSIGN-ID",
		"old_reviewer_id": oldReviewer,
	})

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	s.handler.ReassignPR(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)

	updatedPR, err := s.service.PullRequestService.GetByID("PR-REASSIGN-ID")
	assert.NoError(s.T(), err)

	assert.Len(s.T(), updatedPR.AssignedReviewers, 2, "Количество ревьюеров должно остаться равным 2")
	assert.NotContains(s.T(), updatedPR.AssignedReviewers, oldReviewer, "Старый ревьюер должен исчезнуть из списка")

	for _, rID := range updatedPR.AssignedReviewers {
		assert.NotEqual(s.T(), "u1", rID, "Автор не может стать ревьюером")
		assert.Contains(s.T(), []string{"u2", "u3", "u4"}, rID, "Ревьюер должен быть из состава команды")
	}
}

func (s *PRServiceTestSuite) TestTeamFlow() {
	reqBody, _ := json.Marshal(models.Team{
		TeamName: "Frontend",
		Members: []models.TeamMember{
			{UserID: "u1", Username: "User 1", IsActive: true},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	s.handler.CreateTeam(rr, req)
	assert.Equal(s.T(), http.StatusCreated, rr.Code)

	req = httptest.NewRequest(http.MethodGet, "/team/get?team_name=Frontend", nil)
	rr = httptest.NewRecorder()
	s.handler.GetTeam(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	var team models.Team
	json.Unmarshal(rr.Body.Bytes(), &team)
	assert.Equal(s.T(), "Frontend", team.TeamName)
}

func (s *PRServiceTestSuite) TestSetUserActive() {
	s.db.Exec("INSERT INTO teams (name) VALUES ('Alpha')")
	s.db.Exec("INSERT INTO users (user_id, username, team_name, is_active) VALUES ('active_user', 'Nick', 'Alpha', true)")

	reqBody, _ := json.Marshal(map[string]any{"user_id": "active_user", "is_active": false})
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	s.handler.SetUserActive(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)

	var isActive bool
	s.db.Get(&isActive, "SELECT is_active FROM users WHERE user_id = 'active_user'")
	assert.False(s.T(), isActive)
}

func (s *PRServiceTestSuite) TestMergeAndReviewHistory() {

	s.db.MustExec("INSERT INTO teams (name) VALUES ('Beta')")
	s.db.MustExec("INSERT INTO users (user_id, username, team_name, is_active) VALUES ('auth1', 'Author', 'Beta', true)")
	s.db.MustExec("INSERT INTO users (user_id, username, team_name, is_active) VALUES ('rev1', 'Reviewer', 'Beta', true)")
	s.db.MustExec(`
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers) 
		VALUES ('PR-M', 'Fix Critical Bug', 'auth1', 'OPEN', '["rev1"]')
	`)
	reqBody, _ := json.Marshal(map[string]string{
		"pull_request_id": "PR-M",
	})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	s.handler.MergePR(rr, req)
	assert.Equal(s.T(), http.StatusOK, rr.Code)

	req = httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=rev1", nil)
	rr = httptest.NewRecorder()
	s.handler.GetReview(rr, req)

	assert.Equal(s.T(), http.StatusOK, rr.Code)
	var res struct {
		PRs []models.PullRequestShort `json:"pull_requests"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &res)
	assert.NoError(s.T(), err)

	assert.NotEmpty(s.T(), res.PRs, "Список PR не должен быть пустым")
	if len(res.PRs) > 0 {
		assert.Equal(s.T(), "PR-M", res.PRs[0].PullRequestID)
		assert.Equal(s.T(), "MERGED", string(res.PRs[0].Status))
	}
}

func (s *PRServiceTestSuite) TestCreatePR_Conflict() {
	s.db.MustExec("INSERT INTO teams (name) VALUES ('T1')")
	s.db.MustExec("INSERT INTO users (user_id, username, team_name, is_active) VALUES ('u1', 'U1', 'T1', true)")
	s.db.MustExec(`
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, assigned_reviewers) 
        VALUES ('EXISTING', 'Original Name', 'u1', 'OPEN', '[]')
    `)

	reqBody, _ := json.Marshal(map[string]string{
		"pull_request_id":   "EXISTING",
		"pull_request_name": "New Name",
		"author_id":         "u1",
	})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	s.handler.CreatePR(rr, req)

	assert.Equal(s.T(), http.StatusConflict, rr.Code)
}

func (s *PRServiceTestSuite) TestReassign_MergedPR() {
	s.db.Exec("INSERT INTO pull_requests (pull_request_id, author_id, status, assigned_reviewers) VALUES ('M1', 'u1', 'MERGED', '[\"r1\"]')")

	reqBody, _ := json.Marshal(map[string]string{"pull_request_id": "M1", "old_reviewer_id": "r1"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	s.handler.ReassignPR(rr, req)

	assert.NotEqual(s.T(), http.StatusOK, rr.Code)
}
