## Why

The merged Immich service exposes read-only photo/stat data to two TRMNL plugins. Its API policy should explicitly keep those reads available while ensuring any future configuration or application writes cannot be unauthenticated.

## What Changes

- Keep `/healthz` and plugin-facing read endpoints such as `/api/trmnl/photo/{id}` public.
- Require an authenticated application user and bearer API token for any create/update/delete or API-configuration operation.
- Add secure token creation and lifecycle management before introducing mutations.

## Capabilities

### New Capabilities

- `api-authentication`

### Modified Capabilities

- `public-api-reads`

## Impact

The service route inventory, auth middleware, token persistence/migration, future mutation boundaries, tests, and documentation are affected. Upstream Immich credentials remain plugin configuration and are not replaced by this application token.
