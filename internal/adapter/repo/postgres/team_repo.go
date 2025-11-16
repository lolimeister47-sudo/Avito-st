package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"prservice/internal/domain"
)

type TeamRepo struct {
	db *DB
}

func NewTeamRepo(db *DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, team domain.Team) error {
	_, err := r.db.pool.Exec(ctx,
		`INSERT INTO teams (team_name) VALUES ($1)`,
		string(team.Name),
	)
	return err
}

func (r *TeamRepo) GetTeam(ctx context.Context, name domain.TeamName) (*domain.Team, error) {
	// проверяем, есть ли команда
	var teamName string
	err := r.db.pool.QueryRow(ctx,
		`SELECT team_name FROM teams WHERE team_name = $1`,
		string(name),
	).Scan(&teamName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// подгружаем участников
	rows, err := r.db.pool.Query(ctx,
		`SELECT user_id, username, is_active
		   FROM users
		  WHERE team_name = $1
		  ORDER BY user_id`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var (
			id       string
			username string
			active   bool
		)
		if err := rows.Scan(&id, &username, &active); err != nil {
			return nil, err
		}
		members = append(members, domain.TeamMember{
			UserID:   domain.UserID(id),
			Username: username,
			IsActive: active,
		})
	}

	return &domain.Team{
		Name:    domain.TeamName(teamName),
		Members: members,
	}, nil
}
