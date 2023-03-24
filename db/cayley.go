package db

import (
	"context"
	"database/sql"

	"github.com/caffix/netmap"
	migrate "github.com/rubenv/sql-migrate"
)

// Cayley implements the Upsert interface
type Cayley struct {
	db    *Database
	graph *netmap.Graph
}

func NewCayley(db *Database) *Cayley {
	cayley := netmap.NewCayleyGraph(db.System, db.URL, db.Options)
	g := netmap.NewGraph(cayley)
	return &Cayley{db: db, graph: g}
}

func NewCayleyGraph(g *netmap.Graph) *Cayley {
	return &Cayley{graph: g}
}

func (c *Cayley) NodeSources(ctx context.Context, node netmap.Node, execID int64) ([]string, error) {
	return c.graph.NodeSources(ctx, node, string(execID))
}

func (c *Cayley) InsertFQDN(info InsertInfo, Fqdn FQDN) (int64, error) {
	_, err := c.graph.UpsertFQDN(info.Ctx, Fqdn.Name, info.Source, string(info.EventID))
	return 0, err
}

func (c *Cayley) InsertCNAME(info InsertInfo, dns DNSRecord) error {
	return c.graph.UpsertCNAME(info.Ctx, dns.Fqdn, dns.Target, info.Source, string(info.EventID))
}

func (c *Cayley) InsertPTR(info InsertInfo, dns DNSRecord) error {
	return c.graph.UpsertPTR(info.Ctx, dns.Fqdn, dns.Target, info.Source, string(info.EventID))
}

func (c *Cayley) InsertSRV(info InsertInfo, srv Service) error {
	return c.graph.UpsertSRV(info.Ctx, srv.Fqdn, srv.Service, srv.Target, info.Source, string(info.EventID))
}

func (c *Cayley) InsertNS(info InsertInfo, dns DNSRecord) error {
	return c.graph.UpsertNS(info.Ctx, dns.Fqdn, dns.Target, info.Source, string(info.EventID))
}

func (c *Cayley) InsertMX(info InsertInfo, dns DNSRecord) error {
	return c.graph.UpsertMX(info.Ctx, dns.Fqdn, dns.Target, info.Source, string(info.EventID))
}

func (c *Cayley) InsertInfrastructure(info InsertInfo, infra Infrastructure) error {
	return c.graph.UpsertInfrastructure(info.Ctx, infra.Asn, infra.Description, infra.Address, infra.Cidr, info.Source, string(info.EventID))
}

func (c *Cayley) InsertA(info InsertInfo, record HostRecord) error {
	return c.graph.UpsertA(info.Ctx, record.Fqdn, record.Address, info.Source, string(info.EventID))
}

func (c *Cayley) InsertAAAA(info InsertInfo, record HostRecord) error {
	return c.graph.UpsertAAAA(info.Ctx, record.Fqdn, record.Address, info.Source, string(info.EventID))
}

func (c *Cayley) IsCNAMENode(ctx context.Context, fqdn string) (bool, error) {
	return c.graph.IsCNAMENode(ctx, fqdn), nil
}

func (c *Cayley) InsertExecution(sources []string) (int64, error) {
	return 0, nil
}

func (c *Cayley) Migrate(ctx context.Context, graph *netmap.Graph) error {
	return c.graph.Migrate(ctx, graph)
}

func (c *Cayley) NamesToAddrs(ctx context.Context, execID int64, names ...string) ([]*NameAddrPair, error) {
	netmapRes, err := c.graph.NamesToAddrs(ctx, string(execID), names...)
	var res []*NameAddrPair
	for _, netmapPair := range netmapRes {
		pair := &NameAddrPair{
			Name: netmapPair.Name,
			Addr: netmapPair.Addr,
		}
		res = append(res, pair)
	}
	return res, err

}

func (c *Cayley) EventFQDNs(ctx context.Context, execID int64) []string {
	return c.graph.EventFQDNs(ctx, string(execID))
}

func (c *Cayley) getAppliedMigrationsCount() (int, error) {
	return 0, nil
}

func (c *Cayley) getSqlConnection() (*sql.DB, error) {
	return nil, nil
}

func (c *Cayley) getMigrationsSource() *migrate.FileMigrationSource {
	return nil
}

func (c *Cayley) GetPendingMigrationsCount() (int, error) {
	return 0, nil
}

func (c *Cayley) CreateDatabaseIfNotExists() error {
	return nil
}

func (c *Cayley) DropDatabase() error {
	return nil
}

func (c *Cayley) IsDatabaseCreated() (bool, error) {
	return false, nil
}

func (c *Cayley) RunInitMigration() error {
	return nil
}

func (c *Cayley) RunMigrations() error {
	return nil
}
