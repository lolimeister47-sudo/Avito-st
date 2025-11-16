package usecase

import (
	"context"

	"prservice/internal/domain"
)

type TeamService struct {
	teams domain.TeamRepository
	users domain.UserRepository
}

func NewTeamService(teams domain.TeamRepository, users domain.UserRepository) *TeamService {
	return &TeamService{teams: teams, users: users}
}

func (s *TeamService) AddTeam(ctx context.Context, team domain.Team) (*domain.Team, error) {
	existing, err := s.teams.GetTeam(ctx, team.Name)
	if err == nil && existing != nil {
		return nil, domain.ErrTeamExists
	}

	if err := s.teams.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	for _, m := range team.Members {
		u := domain.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: team.Name,
			IsActive: m.IsActive,
		}
		if err := s.users.UpsertUser(ctx, u); err != nil {
			return nil, err
		}
	}

	res, err := s.teams.GetTeam(ctx, team.Name)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, domain.ErrNotFound
	}
	return res, nil
}

func (s *TeamService) GetTeam(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	team, err := s.teams.GetTeam(ctx, name)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, domain.ErrNotFound
	}
	return team, nil
}
