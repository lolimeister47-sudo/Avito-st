package usecase

import (
	"context"
	"math/rand"
	"time"

	"prservice/internal/domain"
)

type PRService struct {
	prs   domain.PRRepository
	users domain.UserRepository
}

func NewPRService(prs domain.PRRepository, users domain.UserRepository) *PRService {
	return &PRService{prs: prs, users: users}
}

// Создание PR + автоназначение ревьюверов
func (s *PRService) CreatePR(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	existing, err := s.prs.GetByID(ctx, pr.ID)
	if err == nil && existing != nil {
		return nil, domain.ErrPRExists
	}

	author, err := s.users.GetByID(ctx, pr.AuthorID)
	if err != nil || author == nil {
		return nil, domain.ErrNotFound
	}

	now := time.Now().UTC()
	pr.Status = domain.PRStatusOpen
	pr.CreatedAt = &now

	var result *domain.PullRequest

	err = s.prs.WithTx(ctx, func(tx domain.PRTx) error {
		if err := tx.Create(ctx, pr); err != nil {
			return err
		}

		candidates, err := s.users.ListActiveByTeamExcept(ctx, author.TeamName, []domain.UserID{author.ID})
		if err != nil {
			return err
		}

		// перемешиваем
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		maxRev := 2
		if len(candidates) < maxRev {
			maxRev = len(candidates)
		}

		reviewers := make([]domain.UserID, maxRev)
		for i := 0; i < maxRev; i++ {
			reviewers[i] = candidates[i].ID
		}
		pr.AssignedReviewers = reviewers

		if err := tx.Update(ctx, pr); err != nil {
			return err
		}

		loaded, err := tx.GetByIDForUpdate(ctx, pr.ID)
		if err != nil {
			return err
		}
		result = loaded
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Merge (идемпотентный)
func (s *PRService) Merge(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error) {
	var result *domain.PullRequest

	err := s.prs.WithTx(ctx, func(tx domain.PRTx) error {
		pr, err := tx.GetByIDForUpdate(ctx, id)
		if err != nil || pr == nil {
			return domain.ErrNotFound
		}

		if pr.Status == domain.PRStatusMerged {
			result = pr
			return nil
		}

		now := time.Now().UTC()
		pr.Status = domain.PRStatusMerged
		pr.MergedAt = &now

		if err := tx.Update(ctx, *pr); err != nil {
			return err
		}
		result = pr
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// Переназначение ревьювера
func (s *PRService) ReassignReviewer(
	ctx context.Context,
	prID domain.PullRequestID,
	oldReviewer domain.UserID,
) (*domain.PullRequest, domain.UserID, error) {
	var result *domain.PullRequest
	var newReviewerID domain.UserID

	err := s.prs.WithTx(ctx, func(tx domain.PRTx) error {
		pr, err := tx.GetByIDForUpdate(ctx, prID)
		if err != nil || pr == nil {
			return domain.ErrNotFound
		}
		if pr.Status == domain.PRStatusMerged {
			return domain.ErrPRMerged
		}

		idx := -1
		for i, r := range pr.AssignedReviewers {
			if r == oldReviewer {
				idx = i
				break
			}
		}
		if idx == -1 {
			return domain.ErrNotAssigned
		}

		oldUser, err := s.users.GetByID(ctx, oldReviewer)
		if err != nil || oldUser == nil {
			return domain.ErrNotFound
		}

		exclude := append([]domain.UserID{oldReviewer}, pr.AssignedReviewers...)
		candidates, err := s.users.ListActiveByTeamExcept(ctx, oldUser.TeamName, exclude)
		if err != nil {
			return err
		}
		if len(candidates) == 0 {
			return domain.ErrNoCandidate
		}

		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		newUser := candidates[0]

		pr.AssignedReviewers[idx] = newUser.ID

		if err := tx.Update(ctx, *pr); err != nil {
			return err
		}
		result = pr
		newReviewerID = newUser.ID
		return nil
	})

	if err != nil {
		return nil, "", err
	}
	return result, newReviewerID, nil
}

func (s *PRService) ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequest, error) {
	return s.prs.ListByReviewer(ctx, reviewerID)
}
