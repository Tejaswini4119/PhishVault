# PhishVault-2 Phase 1 Implementation Plan

## Goal Description
Implement "The Signal Core (MVP)" as defined in the development plan. This phase establishes the system spine: ingestion, basic scanning, and storage.

## Assignments

### Developer 1: Tejaswini (Infrastructure & Ingestion)
**Responsibility**: Build the nervous system and skeleton (API, Messaging, Database, Data Contracts).

1.  **Unified Signal Abstraction Layer (SAL)**
    *   **Task**: Define the strict JSON schema that serves as the contract between components.
    *   **Output**: `sal_schema.json` (or Go struct equivalent).

2.  **Infrastructure Setup**
    *   **Task**: Configure RabbitMQ for task distribution. Create the basic Worker service scaffold.
    *   **Task**: Setup PostgreSQL (PhishDB Tier 1). Define initial migration scripts for storing Artifacts and Signals.

3.  **Ingestion Engine (REST API)**
    *   **Task**: Create the HTTP API entry point.
    *   **Features**: Accepts URLs, performs basic SHA-256 deduplication, publishes jobs to RabbitMQ.

### Developer 2: Pardhu (Scanning Engines)
**Responsibility**: Build the eyes and sensors (Browsers, HTTP clients, URL processing).

1.  **URL Canonicalization**
    *   **Task**: Implement logic to strip tracking parameters and unwind redirects to find the final effective URL.

2.  **Headless Browser Engine (V2)**
    *   **Task**: Implement a Playwright-based scanner.
    *   **Features**: Visit URL, capture complete DOM, capture Screenshot (full page/viewport).

3.  **Lightweight Fetch Engine**
    *   **Task**: Implement a fast `fasthttp` (or standard lib) scanner.
    *   **Features**: Fetch headers, status codes, and server fingerprints without rendering.

## Proposed Changes (File Structure)

### Core/Shared
#### [NEW] `core/domain/sal.go`
- struct definitions for the SAL.

### Ingestion Service (Tejaswini)
#### [NEW] `services/ingestion/main.go`
- API setup.
#### [NEW] `services/ingestion/producer.go`
- RabbitMQ publisher logic.

### Scanning Service (Pardhu)
#### [NEW] `services/scanner/browser/playwright.go`
- Playwright integration.
#### [NEW] `services/scanner/http/fetch.go`
- Lightweight fetcher.
#### [NEW] `services/scanner/utils/canonicalize.go`
- URL processing logic.

### Infrastructure
#### [NEW] `deploy/docker-compose.yml`
- Definitions for RabbitMQ, Postgres.

## Verification Plan

### Automated Tests
- Unit tests for SAL marshaling/unmarshaling.
- Integration tests for RabbitMQ publishing/consuming.

### Manual Verification
- **End-to-End**: Submit `http://example.com` to API -> Check if RabbitMQ receives msg -> Check if Worker picks up -> Check if DB has entry.
