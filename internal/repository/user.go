package repository

import (
	"context"

	"era_sporta_bot_ruletka/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramUserID int64) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, telegram_user_id, phone, first_name, last_name, username, created_at, updated_at
		FROM users WHERE telegram_user_id = $1
	`, telegramUserID).Scan(
		&u.ID, &u.TelegramUserID, &u.Phone, &u.FirstName, &u.LastName, &u.Username, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx, `
		SELECT id, telegram_user_id, phone, first_name, last_name, username, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.TelegramUserID, &u.Phone, &u.FirstName, &u.LastName, &u.Username, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Upsert(ctx context.Context, u *domain.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (telegram_user_id, phone, first_name, last_name, username, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (telegram_user_id) DO UPDATE SET
			phone = EXCLUDED.phone,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`, u.TelegramUserID, u.Phone, u.FirstName, u.LastName, u.Username).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}
