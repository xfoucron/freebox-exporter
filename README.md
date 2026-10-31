# freebox-exporter

Prometheus exporter for the [Freebox](https://dev.freebox.fr/sdk/os/) (Freebox OS API v8+).

## Metrics

| Subsystem    | Examples                                                                                  |
| ------------ | ----------------------------------------------------------------------------------------- |
| `connection` | state, current and available rates, transferred bytes, type/media/IP info                 |
| `system`     | firmware info, uptime, temperature sensors, fan speeds                                    |
| `lan`        | hosts per interface, per-host `active` / `reachable` and received / transmitted bytes     |
| `storage`    | disk temperature, spin state, partition size / used / free                                |
| `wifi`       | access point state, associated stations                                                   |

Every collector runs concurrently and reports `freebox_scrape_collector_success{collector}` and
`freebox_scrape_collector_duration_seconds{collector}`, so a failing endpoint never hides the others.

## Run

On first start the exporter has no token: it asks the Freebox for one and logs a message. Press the right
arrow on the Freebox front panel to authorize it. The token is stored in the data directory and reused on
every later start, so this happens only once.

```sh
freebox-exporter
```

Then, in Freebox OS (*Paramètres de la Freebox > Gestion des accès > Applications*), give the application
the permissions it needs to read the connection, settings and storage.

| Flag               | Environment         | Default                              |
| ------------------ | ------------------- | ------------------------------------ |
| `--freebox.url`    | `FREEBOX_URL`       | `http://mafreebox.freebox.fr`        |
| `--listen`         | `FREEBOX_LISTEN`    | `:9091`                              |
| `--data-dir`       | `FREEBOX_DATA_DIR`  | `<user config dir>/freebox-exporter` |
| `--scrape-timeout` |                     | `10s`                                |
| (none)             | `FREEBOX_TOKEN`     | optional, overrides the stored token |

The exporter serves `/metrics` and `/healthz`. At most two scrapes run at the same time; further requests
get a `503` until one finishes. Hosts without a name on the Freebox are exported with an empty `name` label,
which Prometheus treats as no label.

`/metrics` is not authenticated and `--listen` binds all interfaces by default. It exposes the Freebox serial
number and MAC address and the name and MAC address of every LAN host, so bind it to a trusted interface
(for example `--listen 127.0.0.1:9091`) or restrict access to the port.

### Docker

```sh
docker run -d --name freebox-exporter -p 9091:9091 -v freebox-data:/data ghcr.io/xfoucron/freebox-exporter
docker logs -f freebox-exporter
```

Press the arrow button on the Freebox when the logs ask for it. With Docker Compose:

```yaml
services:
  freebox-exporter:
    image: ghcr.io/xfoucron/freebox-exporter
    restart: unless-stopped
    ports:
      - "9091:9091"
    volumes:
      - freebox-data:/data

volumes:
  freebox-data:
```

### Prometheus

```yaml
scrape_configs:
  - job_name: freebox
    static_configs:
      - targets: ["localhost:9091"]
```

### Grafana

Import [`dashboards/freebox.json`](dashboards/freebox.json) in Grafana (*Dashboards > New > Import*) and pick
your Prometheus data source. The `job` variable selects the scrape job(s) of the exporter.

## Development

```sh
go test -race ./...
golangci-lint run
```
