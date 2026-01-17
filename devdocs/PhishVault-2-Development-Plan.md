# PhishVault-2: Detailed Development Plan & Staged Rollout

**Reference Document**: `PhishVault-2-TW.md`  
**Strategy**: Operational Feasibility (MVP) First -> Conclusive Intelligence Features Later

This document dissects the technical workbook into four structured execution phases. Each phase builds upon the previous one, ensuring a working, testable system at the end of every stage.

---

## Phase 1: The Signal Core (MVP)

**Objective**: Establish the "Spine" of the system. Enable the ingestion of a URL, execution of successful scans, normalization of data into SAL, and storage in a relational database.

### Focus Areas

1. **Pipeline Infrastructure**: Message Bus (RabbitMQ) + Worker setup.
2. **Core Ingestion**: URL handling only.
3. **Primary Scanning**: Headless Browser & Lightweight HTTP.
4. **Data Definition**: Establishing the SAL JSON Schema.

### Included Features (from TW)

* **5.1 Ingestion Engine (Partial)**:  REST API entry point for **URLs only**. Basic SHA-256 deduplication.
* **5.2 URL Canonicalization**: Implementation of tracking stripping and redirect unwinding (critical for the browser to see the real page).
* **5.14 Unified Signal Abstraction Layer (SAL)**: Defining the strict JSON schema that all engines must output. This is the API contract for the entire system.
* **5.5 Headless Browser Engine (V2)**: The primary analysis tool. Implementation of a basic Playwright container to fetch the DOM and Screenshot.
* **5.6 Lightweight Fetch**: Fast-path `fasthttp` headers scan for quick wins (server server fingerprinting).
* **5.19 PhishDB (Tier 1: Postgres)**: Setup PostgreSQL for storing Artifacts and SAL Signals. No Graph DB yet.

### Operational Feasibility at Completion

* **Capability**: System can take a URL `http://evil.com`, resolve it, capture a screenshot/DOM, and save the raw data to a DB.
* **Missing**: No advanced "Verdict" (just raw signals), no Emails, no Graph.
* **Deliverable**: A CLI or basic API that accepts a URL and returns a JSON object with Screenshot path and DOM.

---

## Phase 2: Deep Intelligence & The Verdict

**Objective**: Transform raw signals into actionable intelligence. Add Email support, AI-driven analysis, and the decision engine to determine "Safe" vs "Malicious".

### Focus Areas

1. **Email Support**: Parsing `.eml` and `.msg`.
2. **AI & Evidence**: Visual AI and NLP models.
3. **The Decision Brain**: Rego-based Verdict Engine.
4. **Blob Storage**: S3/MinIO for mass evidence storage.

### Included Features (from TW)

* **5.3 Ingestion (Email)**: Update ingress to parse Email bodies, headers, and attachments.
* **5.19 PhishDB (Tier 3: S3/MinIO)**: Deploy MinIO to store the screenshots and DOMs generated in Phase 1, instead of ephemeral storage.
* **5.5 Browser Engine (Enhanced)**: Enable stealth modules (WebGL spoofing) and evasion countermeasures.
* **5.9 Brand & Visual AI**: Deploy the Siamese Neural Network to compare Phase 1 screenshots against a "Golden Set".
* **5.10 NLP Engine**: Deploy Transformer models to analyze text from the Browser (DOM text) and Email bodies.
* **5.7 Email Header Intel**: SPF/DKIM/DMARC checks.
* **5.16 Verdict Engine v3**: Integrate Open Policy Agent (OPA). Write the first set of Rego rules (e.g., `Visual_Match > 0.8 -> Malicious`).

### Operational Feasibility at Completion

* **Capability**: A fully functional "Scanner". It can ingest Emails/URLs, run AI models, and give a distinct **Safe/Malicious** verdict based on policy.
* **Deliverable**: A backend service that functions as a complete Phishing Detection API.

---

## Phase 3: Network, Origin & Campaign Graph

**Objective**: Shift from "Single Artifact" analysis to "Campaign" analysis. Implement the Graph Database and infrastructure pivots to track attackers, not just attacks.

### Focus Areas

1. **Graph Database**: Neo4j implementation.
2. **Infrastructure Intel**: pDNS, WHOIS, ASN data.
3. **Clustering**: Grouping artifacts into Campaigns.

### Included Features (from TW)

* **5.19 PhishDB (Tier 2: Neo4j)**: Deploy Neo4j. Design the schema (`:Domain`, `:IP`, `:Cert`).
* **5.8 Infrastructure Intel**: Integration with pDNS/WHOIS providers to enrich artifacts.
* **5.20 Campaign Intelligence Graph**: Logic to push SAL signals into Neo4j nodes and relationships.
* **5.13 Clustering Engine**: Implementation of DBSCAN algorithm to group new artifacts with existing campaigns (e.g., "Same DOM Hash" + "Same ASN").
* **5.23 Origin Path Mapping**: Logic to trace the "Relay Chain" (Redirector -> Origin).
* **5.25 Origin Confidence Scoring**: Validating attribution confidence.

### Operational Feasibility at Completion

* **Capability**: System identifies *who* is attacking and *how* campaigns are connected. It can answer "Show me all phishing sites hosted on ASN 12345 using this specific TLS fingerprint."
* **Deliverable**: A high-fidelity Intelligence Platform.

---

## Phase 4: Operational Governance, Time & Scale

**Objective**: Enterprise readiness. User Interface, Temporal Tracking, Reporting standards, and Production Hardening.

### Focus Areas

1. **The Human Layer**: Analyst Workbench UI.
2. **Time Travel**: Temporal reconstruction.
3. **Governance**: Audits, Overrides, Reporting.
4. **Safety**: Abuse prevention.

### Included Features (from TW)

* **5.26 Analyst Workbench UI**: A React/Next.js dashboard to visualize the data from Postgres and Neo4j.
* **5.22 Temporal Reconstruction**: Scheduler to re-scan artifacts at T+4h, T+12h, etc., and record state changes.
* **5.18 Audit & Override**: UI features to allow analysts to manually override verdicts and log the justification.
* **5.27 Reporting**: Export logic for STIX 2.1 JSON artifacts.
* **5.30 Abuse Prevention**: Quotas and internal-domain filtering to prevent self-dosing or data leakage.
* **5.31 Kubernetes/Scaling**: Helm charts and auto-scaling rules for the Worker nodes.

### Operational Feasibility at Completion

* **Capability**: A distinct commercial-grade product. usable by SOC teams for daily operations, with audit trails and compliance reporting.
* **Deliverable**: PhishVault-2 (Gold Master).
