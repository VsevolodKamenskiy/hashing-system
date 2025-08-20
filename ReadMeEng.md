# Hashing System
This project contains two Go microservices for computing and storing SHA3-256
hashes. Docker Compose orchestrates the services and supporting infrastructure.

## Services

### service1 – gRPC hash calculator (stateless)

`service1` exposes a gRPC API `HasherService` on port `50051`. The single
`CalculateHashes` method accepts a list of strings and returns SHA3‑256 hashes in
the same order.

Example using [grpcurl](https://github.com/fullstorydev/grpcurl):

```bash
grpcurl -plaintext \
  -d '{"strings":["foo","bar"]}' \
  localhost:50051 hasher.HasherService/CalculateHashes
```

Response:

```json
{
  "hashes": [
    "3338be0...", // hash of "foo"
    "fcde2b2..."  // hash of "bar"
  ]
}
```

### service2 – HTTP API and persistence (stateful)

`service2` provides an HTTP API on port `8080` and stores hashes in PostgreSQL.
It relies on `service1` via gRPC for hash calculations.

Endpoints:

* `POST /send` – body: JSON array of strings, returns array of
  objects `{id, hash}`.
* `GET /check?ids=1&ids=2` – returns saved hashes for the provided IDs,
  `204` if none found.

Example:

```bash
# calculate and store hashes
curl -X POST http://localhost:8080/send \
     -H 'Content-Type: application/json' \
     -d '["hello","world"]'

# retrieve hashes by IDs
curl "http://localhost:8080/check?ids=1&ids=2"
```

## Used technologies

- **Go** – implementation language for both services.
- **gRPC** – remote procedure calls between `service2` and `service1`.
- **HTTP/REST** – external API exposed by `service2`.
- **PostgreSQL** – persistent storage for calculated hashes.
- **Docker & Docker Compose** – containerization and orchestration of the services and dependencies.
- **Consul** – service discovery.
- **Graylog** – centralized logging.
- **Prometheus** – metrics collection.

## Running

Use Docker Compose to build and run both services and their dependencies
(PostgreSQL, Consul, Graylog, etc.):

```bash
docker-compose up --build
```

`service1` will be available on `localhost:50051`, and `service2` will listen at
`http://localhost:8080`.



