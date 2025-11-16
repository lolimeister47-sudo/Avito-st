package domain

import "errors"

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrPRExists     = errors.New("pr already exists")
	ErrPRMerged     = errors.New("pr is merged")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this pr")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
	ErrNotFound     = errors.New("resource not found")
)
