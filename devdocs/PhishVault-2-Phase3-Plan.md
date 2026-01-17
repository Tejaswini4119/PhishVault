# PhishVault-2 Phase 3 Implementation Plan

## Goal Description
Implement "Network, Origin & Campaign Graph". This phase shifts the system from analyzing single artifacts to tracking entire campaigns. It introduces a Graph Database (Neo4j) to model relationships between Domains, IPs, and Certificates, and uses Clustering to group related attacks.

## Assignments

### Developer 1: Tejaswini (Data & Intelligence)
**Responsibility**: Build the "Memory Graph" and external intelligence connectors.

1.  **Graph Database (PhishDB Tier 2)**
    *   **Task**: Setup Neo4j database service.
    *   **Task**: Define the Graph Schema (Nodes: `Domain`, `IP`, `ASN`, `Certificate`; Edges: `RESOLVES_TO`, `HOSTED_ON`, `ISSUED_BY`).

2.  **Infrastructure Intelligence**
    *   **Task**: Implement connectors for external Threat Intel providers (e.g., Passive DNS, WHOIS, ASN lookups).
    *   **Note**: For MVP, this can be a simulated local provider or free API.

### Developer 2: Pardhu (Graph Logic & Clustering)
**Responsibility**: Build the "Connective Logic" to identify campaigns.

1.  **Campaign Intelligence Graph (CIG)**
    *   **Task**: Implement the logic that projects SAL signals into Geo-Spatial/Graph nodes in Neo4j.
    *   **Task**: Ensure every new scan updates the graph (e.g., if a Domain resolves to an existing bad IP, link them).

2.  **Clustering Engine**
    *   **Task**: Implement DBSCAN (or similar density-based clustering) to group artifacts based on shared features (DOM Hash, ASN, Logo).
    *   **Task**: Identify "Campaigns" (clusters of nodes).

3.  **Origin Path Mapping**
    *   **Task**: Trace the redirect chain to identify the true origin server, separating it from relay nodes.

## Proposed Changes (File Structure)

### Infrastructure
#### [MODIFY] `deploy/docker-compose.yml`
- Add Neo4j service.

### Core
#### [NEW] `core/domain/graph.go`
- Graph Node/Edge struct definitions.

### Intelligence Service (Tejaswini)
#### [NEW] `services/intel/neo4j.go`
- Neo4j driver wrapper.
#### [NEW] `services/intel/provider/whois.go`
- WHOIS/pDNS client.

### Analysis Service (Pardhu)
#### [NEW] `services/analysis/graph/projector.go`
- Logic to push SAL -> Neo4j.
#### [NEW] `services/analysis/clustering/dbscan.go`
- Clustering algorithm implementation.

## Verification Plan

### Automated Tests
- **Graph**: Unit tests for Node creation (ensure `Merge` is used to avoid duplicates).
- **Clustering**: Test with known clusters (e.g., 3 domains with same DOM hash should form a Cluster).

### Manual Verification
- **Graph Query**: Submit 2 different URLs resolving to the same IP -> Query Neo4j `MATCH (d:Domain)-[:RESOLVES_TO]->(i:IP) RETURN d, i` and verify they share the IP node.
