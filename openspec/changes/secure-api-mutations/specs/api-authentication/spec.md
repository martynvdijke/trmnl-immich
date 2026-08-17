## ADDED Requirements

### Requirement: Plugin reads remain public

The service SHALL allow unauthenticated GET access to `/healthz` and the read-only TRMNL photo/stat endpoints, including `/api/trmnl/photo/{id}` where applicable.

#### Scenario: Photo plugin read

- WHEN a TRMNL client requests a photo read without credentials
- THEN the service returns the configured read payload or a safe not-found response

### Requirement: Future mutations require user and token

Any create, update, delete, API-configuration, or administrative route MUST require an authenticated user and a valid bearer API token, with ownership/role checks retained.

#### Scenario: Unauthenticated configuration write

- WHEN a client submits a future configuration mutation without both credentials
- THEN the service rejects it and changes no configuration

### Requirement: Token lifecycle is secure

Authenticated users SHALL create, list metadata for, revoke, and rotate owned tokens. Secrets MUST be shown only once, stored as hashes, and omitted from logs and later responses.

#### Scenario: Revoked token

- WHEN a revoked token is submitted to a protected route
- THEN the service rejects the request without disclosing token state
