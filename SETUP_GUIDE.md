# PhishVault-2 Infrastructure Setup Guide

This guide details how to initialize and run the PhishVault-2 infrastructure stack (Phase 3).

## Prerequisites

- **Docker Desktop**: Must be installed and running.
- **Git**: To clone the repository.
- **Go 1.24+**: For running services (optional if running locally).

## 1. Start the Infrastructure

The infrastructure is containerized using Docker Compose. It includes:

- **Neo4j**: Graph Database (Campaign Intelligence)
- **RabbitMQ**: Message Broker (Async Ingestion)
- **MinIO**: Object Storage (Artifacts/Screenshots)
- **Postgres**: Relational Database (Tier 1 Metadata)

### Command

Open a terminal in the `deploy/` directory and run:

```bash
cd deploy
docker compose up -d
```

## 2. Verify Services

Once the containers are running, verify access to the management consoles:

### Neo4j Browser (Graph DB)

- **URL**: [http://localhost:7474](http://localhost:7474)
- **Username**: `neo4j`
- **Password**: `password`
- **Action**: Login and verify you can see the query interface.

### MinIO Console (Object Storage)

- **URL**: [http://localhost:9001](http://localhost:9001)
- **Username**: `minioadmin`
- **Password**: `minioadmin`
- **Action**: Create a bucket named `phishvault-artifacts` if it doesn't exist.

### RabbitMQ Management (Messaging)

- **URL**: [http://localhost:15672](http://localhost:15672)
- **Username**: `guest`
- **Password**: `guest`

## 3. Troubleshooting

- **"Connection Refused"**: Ensure Docker Desktop is running.
- **"Port Conflict"**: Check if another service is using ports 7474, 9001, or 5432.
