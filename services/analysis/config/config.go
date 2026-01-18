package config

import (
	"os"
	"strconv"
)

// AnalysisConfig holds configuration parameters for the Analysis Service.
type AnalysisConfig struct {
	GraphBatchSize   int
	ClusterBatchSize int
	Neo4jURI         string
	Neo4jUser        string
	Neo4jPassword    string
}

// LoadConfig reads configuration from environment variables or usage defaults.
func LoadConfig() *AnalysisConfig {
	return &AnalysisConfig{
		GraphBatchSize:   getEnvAsInt("GRAPH_BATCH_SIZE", 10),
		ClusterBatchSize: getEnvAsInt("CLUSTER_BATCH_SIZE", 20),
		Neo4jURI:         getEnv("NEO4J_URI", "neo4j://localhost:7687"),
		Neo4jUser:        getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:    getEnv("NEO4J_PASSWORD", "password"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
