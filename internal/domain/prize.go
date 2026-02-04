package domain

import "time"

type Prize struct {
	ID                int
	Name              string
	Type              string // discount, bonus, free_month, merch
	Value             float64
	ProbabilityWeight int
	IsActive          bool
	CreatedAt         time.Time
}
