package db

import (
	"context"

	"github.com/caffix/netmap"
)

// Cayley implements the Upsert interface
type Cayley struct {
	db    *Database
	graph *netmap.Graph
}

// Create FQDN if it does not exist, otherwise return the ID of the existing FQDN
func (c *Cayley) UpsertFQDN(ctx context.Context, name string, source string, eventID int64) (int64, error) {
	_, err := c.graph.UpsertFQDN(ctx, name, source, string(eventID))
	//node_id := c.graph.NodeToID(node)

	return 0, err
}

func (c *Cayley) UpsertCNAME(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return c.graph.UpsertCNAME(ctx, fqdn, target, source, string(eventID))
}

func (c *Cayley) UpsertPTR(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return c.graph.UpsertPTR(ctx, fqdn, target, source, string(eventID))
}

func (c *Cayley) UpsertSRV(ctx context.Context, fqdn string, service string, target string, source string, eventID int64) error {
	return c.graph.UpsertSRV(ctx, fqdn, service, target, source, string(eventID))
}

func (c *Cayley) UpsertNS(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return c.graph.UpsertNS(ctx, fqdn, target, source, string(eventID))
}

func (c *Cayley) UpsertMX(ctx context.Context, fqdn string, target string, source string, eventID int64) error {
	return c.graph.UpsertMX(ctx, fqdn, target, source, string(eventID))
}

func (c *Cayley) UpsertInfrastructure(ctx context.Context, asn int, desc string, addr string, cidr string, source string, eventID int6464) error {
	return c.graph.UpsertInfrastructure(ctx, asn, desc, addr, cidr, source, string(eventID))
}

func (c *Cayley) UpsertA(ctx context.Context, fqdn string, addr string, source string, eventID int64) error {
	return c.graph.UpsertA(ctx, fqdn, addr, source, string(eventID))
}

func (c *Cayley) UpsertAAAA(ctx context.Context, fqdn string, addr string, source string, eventID int64) error {
	return c.graph.UpsertAAAA(ctx, fqdn, addr, source, string(eventID))
}
