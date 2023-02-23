package models

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"testing"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

func TestMain(m *testing.M) {
	migrations := &migrate.FileMigrationSource{
		Dir: "../migrations/postgres",
	}

	dsn := os.Getenv("PG_DATABASE_URL")
	db, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		log.Printf("Error opening db: %v\n", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error creating generic database: %v\n", err)
	}

	p, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Down)
	if err != nil {
		log.Printf("Error in migrate down, applied %v migrations: %v\n", p, err)
	}

	n, err := migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	if err != nil {
		log.Printf("Error making migrations up, applied %v migrations: %v\n", n, err)
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestCreateEnumExecution(t *testing.T) {
	in_enum := EnumExecution{}
	result := db.Create(&in_enum)
	if result.Error != nil {
		t.Errorf("Error creating enum: %v\n", result.Error)
	}

	var enum EnumExecution
	db.Last(&enum)

	if enum.ID != in_enum.ID {
		t.Errorf("Enum ID is not equal to inserted ID: %v, %v\n", enum.ID, in_enum.ID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if enum.CreatedAt.IsZero() {
		t.Errorf("Enum created at is zero: %v\n", enum.CreatedAt)
	}

	enum2 := EnumExecution{}
	result2 := db.Create(&enum2)
	if result2.Error != nil {
		t.Errorf("Error creating enum: %v\n", result2.Error)
	}

	if enum2.ID < enum.ID {
		t.Errorf("Enum2 ID is less than enum ID: %v, %v\n", enum2.ID, enum.ID)
	}

	if enum2.CreatedAt.Before(enum.CreatedAt) {
		t.Errorf("Enum2 created at is before enum created at: %v, %v\n", enum2.CreatedAt, enum.CreatedAt)
	}
}

func TestCreateFQDNAsset(t *testing.T) {
	enum := EnumExecution{}
	enum_result := db.Create(&enum)
	if enum_result.Error != nil {
		t.Errorf("Error creating enum: %v\n", enum_result.Error)
	}

	fqdn := FQDN{
		Name: "example.com",
		Tld:  "com"}

	fqdn_content, err := json.Marshal(fqdn)
	if err != nil {
		t.Errorf("Error marshalling FQDN: %v\n", err)
	}

	in_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "FQDN",
		Content:         datatypes.JSON(fqdn_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		t.Errorf("Error creating FQDN asset: %v\n", result.Error)
	}

	var asset Asset
	db.Last(&asset)

	if asset.ID != in_asset.ID {
		t.Errorf("Asset ID is not equal to inserted ID: %v, %v\n", asset.ID, in_asset.ID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if asset.CreatedAt.IsZero() {
		t.Errorf("Asset created at is zero: %v\n", asset.CreatedAt)
	}

	if asset.EnumExecutionID != in_asset.EnumExecutionID {
		t.Errorf("Asset enum execution ID is not equal to inserted enum execution ID: %v, %v\n", asset.EnumExecutionID, in_asset.EnumExecutionID)
	}

	if asset.Type != in_asset.Type {
		t.Errorf("Asset type is not equal to inserted type: %v, %v\n", asset.Type, in_asset.Type)
	}

	var result_content FQDN
	err = json.Unmarshal(asset.Content, &result_content)
	if err != nil {
		t.Errorf("Error unmarshalling asset content: %v\n", err)
	}

	if result_content != fqdn {
		t.Errorf("Result content is not equal to original: %v, %v\n", result_content, fqdn)
	}

}

func TestCreateIPAsset(t *testing.T) {
	enum := EnumExecution{}
	enum_result := db.Create(&enum)
	if enum_result.Error != nil {
		t.Errorf("Error creating enum: %v\n", enum_result.Error)
	}

	ip := IPAddress{
		Address: net.IP([]byte{127, 0, 0, 1}),
		Type:    "v4"}

	ip_content, err := json.Marshal(ip)
	if err != nil {
		t.Errorf("Error marshalling IP: %v\n", err)
	}

	in_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "IP",
		Content:         datatypes.JSON(ip_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		t.Errorf("Error creating asset: %v\n", result.Error)
	}

	var asset Asset
	db.Last(&asset)

	if asset.ID != in_asset.ID {
		t.Errorf("Asset ID is not equal to inserted ID: %v, %v\n", asset.ID, in_asset.ID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if asset.CreatedAt.IsZero() {
		t.Errorf("Asset created at is zero: %v\n", asset.CreatedAt)
	}

	if asset.EnumExecutionID != in_asset.EnumExecutionID {
		t.Errorf("Asset enum execution ID is not equal to inserted enum execution ID: %v, %v\n", asset.EnumExecutionID, in_asset.EnumExecutionID)
	}

	if asset.Type != in_asset.Type {
		t.Errorf("Asset type is not equal to inserted type: %v, %v\n", asset.Type, in_asset.Type)
	}

	var result_content IPAddress
	err = json.Unmarshal(asset.Content, &result_content)
	if err != nil {
		t.Errorf("Error unmarshalling asset content: %v\n", err)
	}

	if result_content.Address.String() != ip.Address.String() {
		t.Errorf("Address content is not equal to original: %v, %v\n",
			result_content.Address.String(), ip.Address.String())
	}

	if result_content.Type != ip.Type {
		t.Errorf("Type content is not equal to original: %v, %v\n",
			result_content.Type, ip.Type)
	}
}

func TestCreateASAsset(t *testing.T) {
	enum := EnumExecution{}
	enum_result := db.Create(&enum)
	if enum_result.Error != nil {
		t.Errorf("Error creating enum: %v\n", enum_result.Error)
	}

	as := AutonomousSystem{
		Number: 1234}

	as_content, err := json.Marshal(as)
	if err != nil {
		t.Errorf("Error marshalling AS: %v\n", err)
	}

	in_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "AS",
		Content:         datatypes.JSON(as_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		t.Errorf("Error creating asset: %v\n", result.Error)
	}

	var asset Asset
	db.Last(&asset)

	if asset.ID != in_asset.ID {
		t.Errorf("Asset ID is not equal to inserted ID: %v, %v\n", asset.ID, in_asset.ID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if asset.CreatedAt.IsZero() {
		t.Errorf("Asset created at is zero: %v\n", asset.CreatedAt)
	}

	if asset.EnumExecutionID != in_asset.EnumExecutionID {
		t.Errorf("Asset enum execution ID is not equal to inserted enum execution ID: %v, %v\n", asset.EnumExecutionID, in_asset.EnumExecutionID)
	}

	if asset.Type != in_asset.Type {
		t.Errorf("Asset type is not equal to inserted type: %v, %v\n", asset.Type, in_asset.Type)
	}

	var result_content AutonomousSystem
	err = json.Unmarshal(asset.Content, &result_content)
	if err != nil {
		t.Errorf("Error unmarshalling asset content: %v\n", err)
	}

	if result_content != as {
		t.Errorf("Result content is not equal to original: %v, %v\n", result_content, as)
	}
}

func TestCreateNetblockAsset(t *testing.T) {
	enum := EnumExecution{}
	enum_result := db.Create(&enum)
	if enum_result.Error != nil {
		t.Errorf("Error creating enum: %v\n", enum_result.Error)
	}

	cidr_val := net.IPNet{IP: net.IP([]byte{127, 0, 0, 0}), Mask: net.IPMask([]byte{255, 255, 255, 0})}
	netblock := Netblock{
		Cidr: cidr_val,
		Type: "v4",
	}

	netblock_content, err := json.Marshal(netblock)
	if err != nil {
		t.Errorf("Error marshalling netblock: %v\n", err)
	}

	in_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "Netblock",
		Content:         datatypes.JSON(netblock_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		t.Errorf("Error creating asset: %v\n", result.Error)
	}

	var asset Asset
	db.Last(&asset)

	if asset.ID != in_asset.ID {
		t.Errorf("Asset ID is not equal to inserted ID: %v, %v\n", asset.ID, in_asset.ID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if asset.CreatedAt.IsZero() {
		t.Errorf("Asset created at is zero: %v\n", asset.CreatedAt)
	}

	if asset.EnumExecutionID != in_asset.EnumExecutionID {
		t.Errorf("Asset enum execution ID is not equal to inserted enum execution ID: %v, %v\n", asset.EnumExecutionID, in_asset.EnumExecutionID)
	}

	if asset.Type != in_asset.Type {
		t.Errorf("Asset type is not equal to inserted type: %v, %v\n", asset.Type, in_asset.Type)
	}

	var result_content Netblock
	err = json.Unmarshal(asset.Content, &result_content)
	if err != nil {
		t.Errorf("Error unmarshalling asset content: %v\n", err)
	}

	if result_content.Cidr.String() != netblock.Cidr.String() {
		t.Errorf("CIDR content is not equal to original: %v, %v\n",
			result_content.Cidr.String(), netblock.Cidr.String())
	}

	if result_content.Type != netblock.Type {
		t.Errorf("Type content is not equal to original: %v, %v\n",
			result_content.Type, netblock.Type)
	}

}

func TestCreateRIROrgAsset(t *testing.T) {
	enum := EnumExecution{}
	enum_result := db.Create(&enum)
	if enum_result.Error != nil {
		t.Errorf("Error creating enum: %v\n", enum_result.Error)
	}

	riro := RIROrganization{
		Name:  "RIROrg",
		RIRId: "RIR-1",
		RIR:   "RIR",
	}

	riro_content, err := json.Marshal(riro)
	if err != nil {
		t.Errorf("Error marshalling RIROrg: %v\n", err)
	}

	in_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "RIROrganization",
		Content:         datatypes.JSON(riro_content)}

	result := db.Create(&in_asset)
	if result.Error != nil {
		t.Errorf("Error creating asset: %v\n", result.Error)
	}

	var asset Asset
	db.Last(&asset)

	if asset.ID != in_asset.ID {
		t.Errorf("Asset ID is not equal to inserted ID: %v, %v\n", asset.ID, in_asset.ID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if asset.CreatedAt.IsZero() {
		t.Errorf("Asset created at is zero: %v\n", asset.CreatedAt)
	}

	if asset.EnumExecutionID != in_asset.EnumExecutionID {
		t.Errorf("Asset enum execution ID is not equal to inserted enum execution ID: %v, %v\n", asset.EnumExecutionID, in_asset.EnumExecutionID)
	}

	if asset.Type != in_asset.Type {
		t.Errorf("Asset type is not equal to inserted type: %v, %v\n", asset.Type, in_asset.Type)
	}

	var result_content RIROrganization
	err = json.Unmarshal(asset.Content, &result_content)
	if err != nil {
		t.Errorf("Error unmarshalling asset content: %v\n", err)
	}

	if result_content != riro {
		t.Errorf("Result content is not equal to original: %v, %v\n", result_content, riro)
	}
}

func TestCreateRelation(t *testing.T) {
	enum := EnumExecution{}
	enum_result := db.Create(&enum)
	if enum_result.Error != nil {
		t.Errorf("Error creating enum: %v\n", enum_result.Error)
	}

	fqdn := FQDN{
		Name: "example.com",
		Tld:  "com"}

	fqdn_content, err := json.Marshal(fqdn)
	if err != nil {
		t.Errorf("Error marshalling FQDN: %v\n", err)
	}

	fqdn_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "FQDN",
		Content:         datatypes.JSON(fqdn_content)}

	fqdn_result := db.Create(&fqdn_asset)
	if fqdn_result.Error != nil {
		t.Errorf("Error creating FQDN asset: %v\n", fqdn_result.Error)
	}

	as := AutonomousSystem{
		Number: 1234}

	as_content, err := json.Marshal(as)
	if err != nil {
		t.Errorf("Error marshalling AS: %v\n", err)
	}

	as_asset := Asset{
		EnumExecutionID: enum.ID,
		Type:            "AS",
		Content:         datatypes.JSON(as_content)}

	as_result := db.Create(&as_asset)
	if as_result.Error != nil {
		t.Errorf("Error creating asset: %v\n", as_result.Error)
	}

	in_relation := Relation{
		FromAssetID: fqdn_asset.ID,
		ToAssetID:   as_asset.ID,
		Type:        "related",
	}

	relation_result := db.Create(&in_relation)
	if relation_result.Error != nil {
		t.Errorf("Error creating relation: %v\n", relation_result.Error)
	}

	var relation Relation
	db.Last(&relation)

	if relation.ID != in_relation.ID {
		t.Errorf("Relation ID is not equal to inserted ID: %v, %v\n", relation.ID, in_relation.ID)
	}

	if relation_result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", relation_result.RowsAffected)
	}

	if relation.CreatedAt.IsZero() {
		t.Errorf("Relation created at is zero: %v\n", relation.CreatedAt)
	}

	if relation.FromAssetID != in_relation.FromAssetID {
		t.Errorf("FromAssetID is not equal to inserted FromAssetID: %v, %v\n", relation.FromAssetID, in_relation.FromAssetID)
	}

	if relation.ToAssetID != in_relation.ToAssetID {
		t.Errorf("ToAssetID is not equal to inserted ToAssetID: %v, %v\n", relation.ToAssetID, in_relation.ToAssetID)
	}
}
