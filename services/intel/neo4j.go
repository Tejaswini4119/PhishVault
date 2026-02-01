package intel

import (
	"context"
	"fmt"
	"log"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jClient struct {
	driver neo4j.DriverWithContext
	ctx    context.Context
}

// NewNeo4jClient initializes the Neo4j driver.
func NewNeo4jClient(uri, username, password string) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to neo4j: %w", err)
	}

	return &Neo4jClient{
		driver: driver,
		ctx:    ctx,
	}, nil
}

func (c *Neo4jClient) Close() error {
	return c.driver.Close(c.ctx)
}

// ExecuteBatch writes nodes and edges to Neo4j.
// For efficiency, it should use UNWIND in Cypher, but for MVP clarity we might iterate.
// Let's use parameters properly.
func (c *Neo4jClient) ExecuteBatch(nodes []domain.GraphNode, edges []domain.GraphEdge) error {
	if len(nodes) == 0 && len(edges) == 0 {
		return nil
	}

	session := c.driver.NewSession(c.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(c.ctx)

	_, err := session.ExecuteWrite(c.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 1. Merge Nodes
		for _, n := range nodes {
			// Construct dynamic Cypher based on Label?
			// Ideally we use a parameterized query.
			// MERGE (n:Label {key: $key}) SET n += $props
			query := fmt.Sprintf("MERGE (n:%s {key: $key}) SET n += $props", n.Label)
			params := map[string]interface{}{
				"key":   n.Key,
				"props": n.Properties,
			}
			if _, err := tx.Run(c.ctx, query, params); err != nil {
				return nil, fmt.Errorf("failed to merge node %s: %w", n.Key, err)
			}
		}

		// 2. Merge Edges
		for _, e := range edges {
			// MATCH (a:SourceType {key: $src}), (b:TargetType {key: $tgt})
			// MERGE (a)-[r:RELATION]->(b) SET r += $props
			query := fmt.Sprintf(`
				MATCH (a:%s {key: $srcKey}), (b:%s {key: $tgtKey})
				MERGE (a)-[r:%s]->(b)
				SET r += $props
			`, e.SourceType, e.TargetType, e.Relation)

			params := map[string]interface{}{
				"srcKey": e.SourceKey,
				"tgtKey": e.TargetKey,
				"props":  e.Properties,
			}
			if _, err := tx.Run(c.ctx, query, params); err != nil {
				// Don't fail entire batch if one edge fails (e.g. node missing), but here nodes *should* exist from step 1.
				log.Printf("Failed to merge edge %s->%s: %v", e.SourceKey, e.TargetKey, err)
			}
		}
		return nil, nil
	})

	return err
}

// GetGraphStats returns the total count of nodes and edges in the graph.
func (c *Neo4jClient) GetGraphStats() (map[string]int, error) {
	session := c.driver.NewSession(c.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c.ctx)

	result, err := session.ExecuteRead(c.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		nodesResult, _ := tx.Run(c.ctx, "MATCH (n) RETURN count(n) as count", nil)
		nodesRecord, _ := nodesResult.Single(c.ctx)
		nodesCount, _ := nodesRecord.Get("count")

		edgesResult, _ := tx.Run(c.ctx, "MATCH ()-[r]->() RETURN count(r) as count", nil)
		edgesRecord, _ := edgesResult.Single(c.ctx)
		edgesCount, _ := edgesRecord.Get("count")

		campaignsResult, _ := tx.Run(c.ctx, "MATCH (n:Campaign) RETURN count(n) as count", nil)
		campaignsRecord, _ := campaignsResult.Single(c.ctx)
		campaignsCount, _ := campaignsRecord.Get("count")

		return map[string]int{
			"nodes":     int(nodesCount.(int64)),
			"edges":     int(edgesCount.(int64)),
			"campaigns": int(campaignsCount.(int64)),
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(map[string]int), nil
}

// GetCampaigns returns a list of identified campaigns from the graph.
func (c *Neo4jClient) GetCampaigns() ([]map[string]interface{}, error) {
	session := c.driver.NewSession(c.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(c.ctx)

	result, err := session.ExecuteRead(c.ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Mock query for now, assuming we will have Campaign nodes later
		// Real query would be: MATCH (c:Campaign) RETURN c.id, c.name, ...
		// For MVP, if no campaigns exist, return empty or we can infer from clusters.
		query := `
			MATCH (c:Campaign) 
			RETURN c.id as id, c.name as name, c.threat_actor as actor, c.target_sector as sector, c.risk_level as risk
			LIMIT 50
		`
		result, err := tx.Run(c.ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var campaigns []map[string]interface{}
		for result.Next(c.ctx) {
			rec := result.Record()
			id, _ := rec.Get("id")
			name, _ := rec.Get("name")
			actor, _ := rec.Get("actor")
			sector, _ := rec.Get("sector")
			risk, _ := rec.Get("risk")

			campaigns = append(campaigns, map[string]interface{}{
				"id":            id,
				"name":          name,
				"threat_actor":  actor,
				"target_sector": sector,
				"risk_level":    risk,
				"node_count":    0, // Could count connected nodes with another match
			})
		}
		return campaigns, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]map[string]interface{}), nil
}
