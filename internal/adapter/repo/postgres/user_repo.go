package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"prservice/internal/domain"
)

type UserRepo struct {
	db *DB
}

func NewUserRepo(db *DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) UpsertUser(ctx context.Context, u domain.User) error {
	_, err := r.db.pool.Exec(ctx,
		`INSERT INTO users (user_id, username, team_name, is_active)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) DO UPDATE SET
		   username = EXCLUDED.username,
		   team_name = EXCLUDED.team_name,
		   is_active = EXCLUDED.is_active`,
		string(u.ID),
		u.Username,
		string(u.TeamName),
		u.IsActive,
	)
	return err
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID domain.UserID, isActive bool) (*domain.User, error) {
	var u domain.User
	err := r.db.pool.QueryRow(ctx,
		`UPDATE users
		    SET is_active = $2
		  WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active`,
		string(userID),
		isActive,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID domain.UserID) (*domain.User, error) {
	var u domain.User
	err := r.db.pool.QueryRow(ctx,
		`SELECT user_id, username, team_name, is_active
		   FROM users
		  WHERE user_id = $1`,
		string(userID),
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// ListActiveByTeamExcept — активные участники команды, кроме списка exclude
func (r *UserRepo) ListActiveByTeamExcept(
	ctx context.Context,
	team domain.TeamName,
	exclude []domain.UserID,
) ([]domain.User, error) {
	base := `SELECT user_id, username, team_name, is_active
	           FROM users
	          WHERE team_name = $1
	            AND is_active = TRUE`
	args := []any{string(team)}
	if len(exclude) > 0 {
		var placeholders []string
		for i, id := range exclude {
			args = append(args, string(id))
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		}
		base += " AND user_id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}

	rows, err := r.db.pool.Query(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}
