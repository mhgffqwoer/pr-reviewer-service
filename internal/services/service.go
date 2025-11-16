package services

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