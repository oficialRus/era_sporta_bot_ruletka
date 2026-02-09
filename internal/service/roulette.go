package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"era_sporta_bot_ruletka/internal/domain"
	"era_sporta_bot_ruletka/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RouletteService struct {
	pool      *pgxpool.Pool
	prizeRepo *repository.PrizeRepository
	spinRepo  *repository.SpinRepository
	userRepo  *repository.UserRepository
	spinLimit int
}

func NewRouletteService(
	pool *pgxpool.Pool,
	prizeRepo *repository.PrizeRepository,
	spinRepo *repository.SpinRepository,
	userRepo *repository.UserRepository,
	spinLimit int,
) *RouletteService {
	return &RouletteService{
		pool:      pool,
		prizeRepo: prizeRepo,
		spinRepo:  spinRepo,
		userRepo:  userRepo,
		spinLimit: spinLimit,
	}
}

func (s *RouletteService) Spin(ctx context.Context, userID int64, ipHash string) (*domain.SpinWithPrize, error) {
	// Check limit
	count, err := s.spinRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count spins: %w", err)
	}
	if count >= s.spinLimit {
		return nil, ErrSpinLimitExceeded
	}

	prizes, err := s.prizeRepo.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list prizes: %w", err)
	}
	prizes = filterDisabledPrizes(prizes)
	if len(prizes) == 0 {
		return nil, fmt.Errorf("no active prizes")
	}

	// Weighted random
	var totalWeight int
	for _, p := range prizes {
		totalWeight += p.ProbabilityWeight
	}
	rnd := rand.Intn(totalWeight)
	var chosen *domain.Prize
	for _, p := range prizes {
		rnd -= p.ProbabilityWeight
		if rnd < 0 {
			chosen = p
			break
		}
	}
	if chosen == nil {
		chosen = prizes[len(prizes)-1]
	}

	spin := &domain.Spin{
		UserID:      userID,
		PrizeID:     chosen.ID,
		ResultValue: chosen.Value,
		IPHash:      ipHash,
	}

	// Use transaction with advisory lock to prevent race
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Advisory lock by user_id
	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", userID)
	if err != nil {
		return nil, fmt.Errorf("advisory lock: %w", err)
	}

	// Recheck limit inside transaction
	var cnt int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM spins WHERE user_id = $1", userID).Scan(&cnt)
	if err != nil {
		return nil, err
	}
	if cnt >= s.spinLimit {
		return nil, ErrSpinLimitExceeded
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO spins (user_id, prize_id, result_value, ip_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`, spin.UserID, spin.PrizeID, spin.ResultValue, spin.IPHash).Scan(&spin.ID, &spin.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &domain.SpinWithPrize{Spin: *spin, Prize: chosen}, nil
}

const disabledPrizeName = "Безлимит посещений на 1 месяц"

func filterDisabledPrizes(prizes []*domain.Prize) []*domain.Prize {
	if len(prizes) == 0 {
		return prizes
	}
	out := prizes[:0]
	for _, p := range prizes {
		if p == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(p.Name), disabledPrizeName) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (s *RouletteService) GetConfig(ctx context.Context) ([]*domain.Prize, error) {
	return s.prizeRepo.ListActive(ctx)
}

func (s *RouletteService) GetHistory(ctx context.Context, userID int64, limit int) ([]*domain.SpinWithPrize, error) {
	return s.spinRepo.ListByUserID(ctx, userID, limit)
}

var ErrSpinLimitExceeded = fmt.Errorf("spin limit exceeded")
