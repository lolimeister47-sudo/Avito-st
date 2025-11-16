package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"prservice/internal/domain"
)

type PRRepo struct {
	db *DB
}

func NewPRRepo(db *DB) *PRRepo {
	return &PRRepo{db: db}
}

// ==================== domain.PRRepository ====================

func (r *PRRepo) WithTx(ctx context.Context, fn func(tx domain.PRTx) error) error {
	conn, err := r.db.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	wrapped := &prTx{tx: tx}

	if err := fn(wrapped); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (r *PRRepo) GetByID(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error) {
	row := r.db.pool.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		   FROM pull_requests
		  WHERE pull_request_id = $1`,
		string(id),
	)

	pr, err := scanPR(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// подгружаем ревьюверов
	reviewers, err := r.loadReviewers(ctx, id)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (r *PRRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	_, err := r.db.pool.Exec(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		string(pr.ID),
		pr.Name,
		string(pr.AuthorID),
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}
	return r.saveReviewers(ctx, pr.ID, pr.AssignedReviewers)
}

func (r *PRRepo) Update(ctx context.Context, pr domain.PullRequest) error {
	_, err := r.db.pool.Exec(ctx,
		`UPDATE pull_requests
		    SET pull_request_name = $2,
		        author_id = $3,
		        status = $4,
		        created_at = $5,
		        merged_at = $6
		  WHERE pull_request_id = $1`,
		string(pr.ID),
		pr.Name,
		string(pr.AuthorID),
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}
	return r.saveReviewers(ctx, pr.ID, pr.AssignedReviewers)
}

func (r *PRRepo) ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequest, error) {
	rows, err := r.db.pool.Query(ctx,
		`SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		   FROM pull_requests pr
		   JOIN pull_request_reviewers r ON r.pull_request_id = pr.pull_request_id
		  WHERE r.reviewer_id = $1
		  ORDER BY pr.created_at DESC NULLS LAST`,
		string(reviewerID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		// для /users/getReview AssignedReviewers не требуется
		res = append(res, pr)
	}
	return res, nil
}

// ==================== domain.PRTx ====================

type prTx struct {
	tx pgx.Tx
}

func (t *prTx) GetByIDForUpdate(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error) {
	row := t.tx.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		   FROM pull_requests
		  WHERE pull_request_id = $1
		  FOR UPDATE`,
		string(id),
	)

	pr, err := scanPR(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	reviewers, err := loadReviewersTx(ctx, t.tx, id)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (t *prTx) Create(ctx context.Context, pr domain.PullRequest) error {
	_, err := t.tx.Exec(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		string(pr.ID),
		pr.Name,
		string(pr.AuthorID),
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}
	return saveReviewersTx(ctx, t.tx, pr.ID, pr.AssignedReviewers)
}

func (t *prTx) Update(ctx context.Context, pr domain.PullRequest) error {
	_, err := t.tx.Exec(ctx,
		`UPDATE pull_requests
		    SET pull_request_name = $2,
		        author_id = $3,
		        status = $4,
		        created_at = $5,
		        merged_at = $6
		  WHERE pull_request_id = $1`,
		string(pr.ID),
		pr.Name,
		string(pr.AuthorID),
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}
	return saveReviewersTx(ctx, t.tx, pr.ID, pr.AssignedReviewers)
}

// domain.PRRepository methods reused

func (t *prTx) WithTx(ctx context.Context, fn func(tx domain.PRTx) error) error {
	// транзакция уже открыта — вложенные транзакции не нужны
	return fn(t)
}

func (t *prTx) GetByID(ctx context.Context, id domain.PullRequestID) (*domain.PullRequest, error) {
	return t.GetByIDForUpdate(ctx, id)
}

func (t *prTx) ListByReviewer(ctx context.Context, reviewerID domain.UserID) ([]domain.PullRequest, error) {
	rows, err := t.tx.Query(ctx,
		`SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		   FROM pull_requests pr
		   JOIN pull_request_reviewers r ON r.pull_request_id = pr.pull_request_id
		  WHERE r.reviewer_id = $1`,
		string(reviewerID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		res = append(res, pr)
	}
	return res, nil
}

// ==================== helpers ====================

func scanPR(row pgx.Row) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	if err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PRRepo) loadReviewers(ctx context.Context, id domain.PullRequestID) ([]domain.UserID, error) {
	rows, err := r.db.pool.Query(ctx,
		`SELECT reviewer_id
		   FROM pull_request_reviewers
		  WHERE pull_request_id = $1`,
		string(id),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []domain.UserID
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, domain.UserID(rid))
	}
	return reviewers, nil
}

func (r *PRRepo) saveReviewers(ctx context.Context, id domain.PullRequestID, reviewers []domain.UserID) error {
	// сначала очистим
	if _, err := r.db.pool.Exec(ctx,
		`DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`, string(id)); err != nil {
		return err
	}
	for _, rid := range reviewers {
		if _, err := r.db.pool.Exec(ctx,
			`INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			 VALUES ($1, $2)`,
			string(id),
			string(rid),
		); err != nil {
			return err
		}
	}
	return nil
}

func loadReviewersTx(ctx context.Context, tx pgx.Tx, id domain.PullRequestID) ([]domain.UserID, error) {
	rows, err := tx.Query(ctx,
		`SELECT reviewer_id
		   FROM pull_request_reviewers
		  WHERE pull_request_id = $1`,
		string(id),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []domain.UserID
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, domain.UserID(rid))
	}
	return reviewers, nil
}

func saveReviewersTx(ctx context.Context, tx pgx.Tx, id domain.PullRequestID, reviewers []domain.UserID) error {
	if _, err := tx.Exec(ctx,
		`DELETE FROM pull_request_reviewers WHERE pull_request_id = $1`, string(id)); err != nil {
		return err
	}
	for _, rid := range reviewers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			 VALUES ($1, $2)`,
			string(id),
			string(rid),
		); err != nil {
			return err
		}
	}
	return nil
}
