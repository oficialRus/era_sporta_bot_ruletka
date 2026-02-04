package repository

import (
	"context"

	"era_sporta_bot_ruletka/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PrizeRepository struct {
	pool *pgxpool.Pool
}

func NewPrizeRepository(pool *pgxpool.Pool) *PrizeRepository {
	return &PrizeRepository{pool: pool}
}

func (r *PrizeRepository) ListActive(ctx context.Context) ([]*domain.Prize, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, type, value, probability_weight, is_active, created_at
		FROM prizes WHERE is_active = true ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prizes []*domain.Prize
	for rows.Next() {
		var p domain.Prize
		if err := rows.Scan(&p.ID, &p.Name, &p.Type, &p.Value, &p.ProbabilityWeight, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		prizes = append(prizes, &p)
	}
	return prizes, rows.Err()
}

func (r *PrizeRepository) GetByID(ctx context.Context, id int) (*domain.Prize, error) {
	var p domain.Prize
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, type, value, probability_weight, is_active, created_at
		FROM prizes WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.Type, &p.Value, &p.ProbabilityWeight, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
