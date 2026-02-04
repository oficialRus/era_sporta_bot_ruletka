package repository

import (
	"context"

	"era_sporta_bot_ruletka/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SpinRepository struct {
	pool *pgxpool.Pool
}

func NewSpinRepository(pool *pgxpool.Pool) *SpinRepository {
	return &SpinRepository{pool: pool}
}

func (r *SpinRepository) Create(ctx context.Context, s *domain.Spin) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO spins (user_id, prize_id, result_value, ip_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`, s.UserID, s.PrizeID, s.ResultValue, s.IPHash).Scan(&s.ID, &s.CreatedAt)
}

func (r *SpinRepository) CountByUserID(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM spins WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *SpinRepository) ListByUserID(ctx context.Context, userID int64, limit int) ([]*domain.SpinWithPrize, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.user_id, s.prize_id, s.result_value, s.ip_hash, s.created_at,
		       p.id, p.name, p.type, p.value, p.probability_weight, p.is_active, p.created_at
		FROM spins s
		JOIN prizes p ON p.id = s.prize_id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*domain.SpinWithPrize
	for rows.Next() {
		var swp domain.SpinWithPrize
		swp.Prize = &domain.Prize{}
		err := rows.Scan(
			&swp.ID, &swp.UserID, &swp.PrizeID, &swp.ResultValue, &swp.IPHash, &swp.CreatedAt,
			&swp.Prize.ID, &swp.Prize.Name, &swp.Prize.Type, &swp.Prize.Value, &swp.Prize.ProbabilityWeight, &swp.Prize.IsActive, &swp.Prize.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &swp)
	}
	return result, rows.Err()
}
