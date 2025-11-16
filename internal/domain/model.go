package domain

import "time"

type TeamName string
type UserID string
type PullRequestID string

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type TeamMember struct {
	UserID   UserID
	Username string
	IsActive bool
}

type Team struct {
	Name    TeamName
	Members []TeamMember
}

type User struct {
	ID       UserID
	Username string
	TeamName TeamName
	IsActive bool
}

type PullRequest struct {
	ID                PullRequestID
	Name              string
	AuthorID          UserID
	Status            PRStatus
	AssignedReviewers []UserID
	CreatedAt         *time.Time
	MergedAt          *time.Time
}
