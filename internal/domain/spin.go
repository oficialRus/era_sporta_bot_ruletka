package domain

import "time"

type Spin struct {
	ID          int64
	UserID      int64
	PrizeID     int
	ResultValue float64
	IPHash      string
	CreatedAt   time.Time
}

type SpinWithPrize struct {
	Spin
	Prize *Prize
}
