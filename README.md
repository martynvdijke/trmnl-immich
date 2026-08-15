# trmnl-immich

A [TRMNL](https://usetrmnl.com) plugin that shows a random photo from your
[Immich](https://immich.app) library each day. Optionally restrict the photo to
a specific album or person.

## How it works

```
┌──────────┐   polling   ┌────────────────┐  x-api-key  ┌─────────┐
│ TRMNL    │ ──────────► │ trmnl-immich   │ ──────────► │ Immich  │
│ device   │ ◄────────── │ Go backend     │ ◄────────── │ server  │
└──────────┘  JSON+image └────────────────┘             └─────────┘
```

Immich requires an API key (`x-api-key` header) for every request. The TRMNL
device cannot hold that key, so this small Go backend proxies Immich:

- `GET /api/trmnl/photo-of-the-day` — polls Immich's `POST /api/search/random`
  and returns a random photo as JSON for the TRMNL device.
- `GET /api/trmnl/photo/{id}` — streams the Immich thumbnail so the device can
  render the image without an API key.

The `trmnl/` directory is a [trmnlp](https://github.com/owise1/trmnlp) plugin
project (Liquid templates + settings) that is pushed to your TRMNL plugin via
`trmnlp push`.

## Configuration

| Variable         | Required | Default | Description                     |
| ---------------- | -------- | ------- | ------------------------------- |
| `IMMICH_URL`     | yes      | —       | Base URL of your Immich server  |
| `IMMICH_API_KEY` | yes      | —       | Immich API key                  |
| `PORT`           | no       | `8080`  | HTTP listen port                |

## Run

```sh
# from source
go run .

# or with docker
docker compose up -d
```

## TRMNL plugin setup

1. Host this backend somewhere public (`https://<backend>/healthz` should
   return `ok`).
2. Create a new plugin in the TRMNL dashboard (or push via `trmnlp push`) with
   polling URL `https://<backend>/api/trmnl/photo-of-the-day`.
3. Set the custom fields:
   - **url** — the public backend URL (required).
   - **album_id** — optional Immich album ID (leave empty for the whole
     library). Find it in the Immich web UI or via `GET /api/albums`.
   - **person_id** — optional Immich person ID. Find it in the Immich web UI
     or via `GET /api/people`.
4. Set the refresh interval to daily (1440 minutes) for a photo-of-the-day.

For local development, run `trmnlp serve` inside `trmnl/`.

## Development

```sh
task setup   # download Go modules
task dev     # run with air (auto-reload)
task test    # run unit tests
task build   # build the binary
```

## License

See [LICENSE](LICENSE).
