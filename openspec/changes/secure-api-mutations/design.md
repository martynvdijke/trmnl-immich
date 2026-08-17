## Context

The repository now contains Immich photo and stats plugins. Current application routes are primarily reads and proxy calls to an upstream Immich instance. Authentication must not break public TRMNL polling, while any future write surface starts protected by default.

## Goals / Non-Goals

### Goals

- Preserve public health, photo, and stats reads.
- Establish a reusable user-scoped bearer-token boundary for mutations.
- Keep upstream Immich API keys separate from application tokens.

### Non-Goals

- Making read-only TRMNL endpoints private.
- Exposing or rotating upstream Immich keys through this change.
- Generating a production token now.

## Decisions

- Add a reversible token store with owner, hash, timestamps, expiry, and revocation metadata.
- Add middleware that protects all non-read routes by default once a mutation is introduced, with explicit public-read allowlisting.
- Provide owner-only create/list/revoke/rotate operations and one-time secret output.
- Add route tests before adding any new mutation endpoint.

## Risks / Trade-offs

There may be no current application mutation route, so most work is preventive. An explicit allowlist avoids accidentally exposing future writes but requires route updates when new public reads are added.

## Migration Plan

1. Confirm the current read-only route inventory and add token persistence.
2. Implement lifecycle endpoints and middleware.
3. Add public-read and protected-future-mutation tests.
4. Require the policy in code review for every new route.

## Open Questions

- Should future photo-management writes be proxied through this service or remain upstream-only?
