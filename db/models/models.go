package models

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type EnumerationExecution struct {
	ID      int64     `gorm:"primaryKey"`
	Created time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
}

type Asset struct {
	ID      int64                `gorm:"primaryKey"`
	Enum    EnumerationExecution `gorm:"embedded"`
	Type    string
	Content pgtype.JSONB
}

type Relation struct {
	ID   int64 `gorm:"primaryKey"`
	Type string
	From Asset `gorm:"embedded"`
	To   Asset `gorm:"embedded"`
}
