package models

import (
	"time"

	"github.com/jackc/pgx/pgtype"
)

type EnumExecution struct {
	ID      int64     `gorm:"primaryKey;autoIncrement:true"`
	Created time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
}

type Asset struct {
	ID              int64     `gorm:"primaryKey;autoIncrement:true"`
	Created         time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
	EnumExecutionID int64
	EnumExecution   EnumExecution
	Type            string
	Content         pgtype.JSONB
}

type Relation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:true"`
	Created     time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
	Type        string
	FromAssetID int64
	ToAssetID   int64
	FromAsset   Asset
	ToAsset     Asset
}
