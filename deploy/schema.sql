-- PhishDB Tier 1 Schema

CREATE TABLE IF NOT EXISTS scans (
    scan_id VARCHAR(64) PRIMARY KEY,
    url TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ingestion_source VARCHAR(50),
    verdict VARCHAR(20),
    risk_score FLOAT
);

CREATE TABLE IF NOT EXISTS scan_details (
    scan_id VARCHAR(64) PRIMARY KEY REFERENCES scans(scan_id),
    request_method VARCHAR(10),
    request_headers JSONB,
    response_status INT,
    response_headers JSONB,
    final_url TEXT,
    ip VARCHAR(45),
    asn VARCHAR(50),
    body_hash VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS artifacts (
    id SERIAL PRIMARY KEY,
    scan_id VARCHAR(64) REFERENCES scans(scan_id),
    artifact_type VARCHAR(50), -- e.g., 'screenshot', 'dom'
    path TEXT, -- MinIO path or local path
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scans_url ON scans(url);
CREATE INDEX idx_scans_timestamp ON scans(timestamp);
