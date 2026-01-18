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
