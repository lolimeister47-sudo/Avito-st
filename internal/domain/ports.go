package domain

import "context"

type TeamRepository interface {
	CreateTeam(ctx context.Context, team Team) error
	GetTeam(ctx context.Context, name TeamName) (*Team, error)
}

type UserRepository interface {
	UpsertUser(ctx context.Context, user User) error
	SetIsActive(ctx context.Context, userID UserID, isActive bool) (*User, error)
	GetByID(ctx context.Context, userID UserID) (*User, error)
	ListActiveByTeamExcept(ctx context.Context, team TeamName, exclude []UserID) ([]User, error)
}

type PRRepository interface {
	WithTx(ctx context.Context, fn func(tx PRTx) error) error

	GetByID(ctx context.Context, id PullRequestID) (*PullRequest, error)
	Create(ctx context.Context, pr PullRequest) error
	Update(ctx context.Context, pr PullRequest) error
	ListByReviewer(ctx context.Context, reviewerID UserID) ([]PullRequest, error)
}

type PRTx interface {
	GetByIDForUpdate(ctx context.Context, id PullRequestID) (*PullRequest, error)
	Create(ctx context.Context, pr PullRequest) error
	Update(ctx context.Context, pr PullRequest) error
}
