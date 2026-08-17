## 1. Inventory and storage

- [ ] 1.1 Confirm all current health, photo, stats, and upstream proxy routes and classify reads.
- [ ] 1.2 Add reversible hashed-token migration and indexes.

## 2. Auth boundary

- [ ] 2.1 Implement bearer validation, expiry, revocation, and owner matching.
- [ ] 2.2 Allowlist public reads and default future mutations to protected.
- [ ] 2.3 Implement owner-only token create/list/revoke/rotate flows.

## 3. Verification

- [ ] 3.1 Test public plugin reads and unauthorized future mutation behavior.
- [ ] 3.2 Test token ownership, malformed/expired/revoked tokens, and secret redaction.
- [ ] 3.3 Document application tokens separately from upstream Immich API keys and run CI.
