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

func TestCreateExecution(t *testing.T) {
	in_enum := Execution{
		Domains: "example.com",
	}
	result := db.Create(&in_enum)
	if result.Error != nil {
		t.Errorf("Error creating enum: %v\n", result.Error)
	}

	var enum Execution
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

	enum2 := Execution{
		Domains: "foo.bar",
	}
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

func TestCreateExecutionLog(t *testing.T) {
	// Create an Execution
	in_enum := Execution{
		Domains: "example.com",
	}
	result := db.Create(&in_enum)
	if result.Error != nil {
		t.Errorf("Error creating execution: %v\n", result.Error)
	}

	fqdn := FQDN{
		Name: "example.com",
		Tld:  "com"}

	fqdn_content, err := json.Marshal(fqdn)
	if err != nil {
		t.Errorf("Error marshalling FQDN: %v\n", err)
	}

	in_asset := Asset{
		Type:    "FQDN",
		Content: datatypes.JSON(fqdn_content)}

	result = db.Create(&in_asset)
	if result.Error != nil {
		t.Errorf("Error creating FQDN asset: %v\n", result.Error)
	}

	// Create an ExecutionLog referencing the Execution and Asset
	in_el := ExecutionLog{
		ExecutionID: in_enum.ID,
		AssetID:     in_asset.ID}

	result = db.Create(&in_el)
	if result.Error != nil {
		t.Errorf("Error creating execution log: %v\n", result.Error)
	}

	var el ExecutionLog
	db.Last(&el)

	if el.ID != in_el.ID {
		t.Errorf("ExecutionLog ID is not equal to inserted ID: %v, %v\n", el.ID, in_el.ID)
	}

	if el.ExecutionID != in_el.ExecutionID {
		t.Errorf("ExecutionLog ExecutionID is not equal to inserted ExecutionID: %v, %v\n", el.ExecutionID, in_el.ExecutionID)
	}

	if el.AssetID != in_el.AssetID {
		t.Errorf("ExecutionLog AssetID is not equal to inserted AssetID: %v, %v\n", el.AssetID, in_el.AssetID)
	}

	if result.RowsAffected != 1 {
		t.Errorf("Rows affected is not 1: %v\n", result.RowsAffected)
	}

	if el.CreatedAt.IsZero() {
		t.Errorf("ExecutionLog created at is zero: %v\n", el.CreatedAt)
	}

	// Create another asset
	fqdn2 := FQDN{
		Name: "foo.bar",
		Tld:  "bar"}

	fqdn2_content, err := json.Marshal(fqdn2)
	if err != nil {
		t.Errorf("Error marshalling FQDN: %v\n", err)
	}

	in_asset2 := Asset{
		Type:    "FQDN",
		Content: datatypes.JSON(fqdn2_content)}

	result = db.Create(&in_asset2)
	if result.Error != nil {
		t.Errorf("Error creating FQDN asset 2: %v\n", result.Error)
	}

	// Create an ExecutionLog referencing the Execution and Asset
	in_el2 := ExecutionLog{
		ExecutionID: in_enum.ID,
		AssetID:     in_asset2.ID}

	result = db.Create(&in_el2)
	if result.Error != nil {
		t.Errorf("Error creating execution log: %v\n", result.Error)
	}

	// Fetch the created asset and test that it can get the related execution logs
	var asset Asset
	db.First(&asset)
	var execution_logs []ExecutionLog
	err = db.Model(&asset).Association("ExecutionLogs").Find(&execution_logs)
	if err != nil {
		t.Errorf("Error fetching execution logs: %v\n", err)
	}

	if len(execution_logs) != 1 {
		t.Errorf("Execution Logs length is not 1: %v\n", len(execution_logs))
	}

	// Fetch the discovered assets during an Execution
	// Produces SELECT "assets"."id","assets"."created_at","assets"."type","assets"."content"
	// FROM "assets" JOIN "execution_logs" ON "execution_logs"."asset_id" = "assets"."id"
	//   AND "execution_logs"."execution_id" = 3
	var execution Execution
	db.Last(&execution)
	var assets []Asset
	err = db.Model(&execution).Association("Assets").Find(&assets)
	if err != nil {
		t.Errorf("Error fetching assets from an given execution: %v\n", err)
	}

	if len(assets) != 2 {
		t.Errorf("Assets length is not 2: %v\n", len(assets))
	}

	// Fetch all the assets found during an Execution through the ExecutionLog
	// Produces SELECT * FROM "assets" WHERE "assets"."id" IN (1,2)
	var execution_logs2 []ExecutionLog
	err = db.Model(&execution).Association("ExecutionLogs").Find(&execution_logs2)
	if err != nil {
		t.Errorf("Error fetching execution logs from an given execution: %v\n", err)
	}

	err = db.Model(&execution_logs2).Association("Asset").Find(&assets)
	if err != nil {
		t.Errorf("Error fetching assets from different execution logs: %v\n", err)
	}

	if len(assets) != 2 {
		t.Errorf("Assets length is not 2: %v\n", len(assets))
	}

}

func TestCreateFQDNAsset(t *testing.T) {
	fqdn := FQDN{
		Name: "example.com",
		Tld:  "com"}

	fqdn_content, err := json.Marshal(fqdn)
	if err != nil {
		t.Errorf("Error marshalling FQDN: %v\n", err)
	}

	in_asset := Asset{
		Type:    "FQDN",
		Content: datatypes.JSON(fqdn_content)}

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
	ip := IPAddress{
		Address: net.IP([]byte{127, 0, 0, 1}),
		Type:    "v4"}

	ip_content, err := json.Marshal(ip)
	if err != nil {
		t.Errorf("Error marshalling IP: %v\n", err)
	}

	in_asset := Asset{
		Type:    "IP",
		Content: datatypes.JSON(ip_content)}

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
	as := AutonomousSystem{
		Number: 1234}

	as_content, err := json.Marshal(as)
	if err != nil {
		t.Errorf("Error marshalling AS: %v\n", err)
	}

	in_asset := Asset{
		Type:    "AS",
		Content: datatypes.JSON(as_content)}

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
		Type:    "Netblock",
		Content: datatypes.JSON(netblock_content)}

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
		Type:    "RIROrganization",
		Content: datatypes.JSON(riro_content)}

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
	fqdn := FQDN{
		Name: "example.com",
		Tld:  "com"}

	fqdn_content, err := json.Marshal(fqdn)
	if err != nil {
		t.Errorf("Error marshalling FQDN: %v\n", err)
	}

	fqdn_asset := Asset{
		Type:    "FQDN",
		Content: datatypes.JSON(fqdn_content)}

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
		Type:    "AS",
		Content: datatypes.JSON(as_content)}

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
