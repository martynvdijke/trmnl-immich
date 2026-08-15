# trmnl-immich

A [TRMNL](https://usetrmnl.com) plugin that shows a random photo from your
[Immich](https://immich.app) library each day. Optionally restrict the photo
to a specific album or person.

## How it works

The plugin polls Immich **directly** — no middleware needed for the data:

```
┌──────────┐   POST /api/search/random     ┌─────────┐
│ TRMNL    │ ────────────────────────────► │ Immich  │
│ device   │   x-api-key header            │ server  │
└──────────┘ ◄──────────────────────────── └─────────┘
                 random photo JSON
```

- `polling_url`: `{{url}}/api/search/random`
- `polling_headers`: `x-api-key={{api_key}}`
- `polling_body`: `{"size":1,"albumIds":{{album_id}},"personIds":{{person_id}}}`
- `src/transform.py` reshapes Immich's response array into a photo card for
  the Liquid templates.

The `url`, `api_key`, `album_id`, and `person_id` custom fields are
interpolated into the request. Album/person filters are **JSON arrays of
UUIDs** (e.g. `["f3b9..."]`) and default to `[]` (the whole library).

### Why a small image proxy remains

Immich thumbnails (`GET /api/assets/{id}/thumbnail`) require the `x-api-key`
header, and the TRMNL device cannot attach headers to `<img>` tags. This repo
therefore keeps a **minimal Go backend** that proxies only the image bytes:

- `GET /api/trmnl/photo/{id}` — streams the Immich preview thumbnail.

The JSON data itself flows device → Immich directly.

## Configuration

| Variable         | Required | Default | Description                     |
| ---------------- | -------- | ------- | ------------------------------- |
| `IMMICH_URL`     | yes      | —       | Base URL of your Immich server  |
| `IMMICH_API_KEY` | yes      | —       | Immich API key                  |
| `PORT`           | no       | `8080`  | HTTP listen port                |

## Run the proxy

```sh
# from source
go run .

# or with docker
docker compose up -d
```

## TRMNL plugin setup

1. Create a new plugin in the TRMNL dashboard (or push via `trmnlp push`).
2. Set the custom fields:
   - **url** — the public URL of your Immich server (required).
   - **api_key** — an Immich API key (Account Settings > API Keys).
   - **album_id** — optional JSON array of album UUIDs, e.g. `["<uuid>"]`.
     Leave `[]` for the whole library.
   - **person_id** — optional JSON array of person UUIDs, e.g. `["<uuid>"]`.
     Leave `[]` for the whole library.
3. Set the refresh interval to daily (1440 minutes) for a photo-of-the-day.

The image proxy must be reachable at the URL used in the plugin's Liquid
templates (the **url** custom field points at Immich, while the image URL
served by the backend is `https://<proxy-host>/api/trmnl/photo/<id>` — set the
proxy host in your deployment and update the templates' image `src` prefix
accordingly if it differs).

For local development, run `trmnlp serve` inside `trmnl/`. Transform scripts
are validated with `python3 src/transform.py < fixture.json`.

> **Security notes:** the API key is stored in TRMNL's custom fields, and
> `src/transform.*` runs automatically when previewing/cloning the plugin.
> Review third-party plugins before serving them, or set
> `transform_runtime: disabled` in `.trmnlp.yml`.

## Development

```sh
task setup   # download Go modules
task dev     # run with air (auto-reload)
task test    # run unit tests
task build   # build the binary
```

## License

See [LICENSE](LICENSE).
