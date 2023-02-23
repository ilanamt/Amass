package models

import (
	"net"
	"time"

	"gorm.io/datatypes"
)

type EnumExecution struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:true"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
}

type Asset struct {
	ID              int64     `gorm:"primaryKey;autoIncrement:true"`
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
	EnumExecutionID int64
	EnumExecution   EnumExecution
	Type            string
	Content         datatypes.JSON
}

type Relation struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:true"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP()"`
	Type        string
	FromAssetID int64
	ToAssetID   int64
	FromAsset   Asset
	ToAsset     Asset
}

type FQDN struct {
	Name string `json:"name"`
	Tld  string `json:"tld"`
}

type AutonomousSystem struct {
	Number int64 `json:"number"`
}

type RIROrganization struct {
	Name  string `json:"name"`
	RIRId string `json:"rir_id"`
	RIR   string `json:"rir"`
}

type IPAddress struct {
	Address net.IP `json:"address"`
	Type    string `json:"type"`
}

type Netblock struct {
	Cidr net.IPNet `json:"cidr"`
	Type string    `json:"type"`
}
