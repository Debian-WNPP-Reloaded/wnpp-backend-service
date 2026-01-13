# GO Backend Service for Debian WNPP Reloaded

This project provides a backend API written in Go for interacting with Debian's Ultimate Debian Database (UDD). It exposes endpoints to query WNPP (Work-Needing and Prospective Packages) data.

---

## How to Run

```bash
export DATABASE_URL="postgresql://udd-mirror:udd-mirror@udd-mirror.debian.net:5432/udd"
go run ./cmd/api
```

## Examples of requests

### Health check
```bash
curl http://localhost:8080/health
```

### Basic WNPP Queries
```bash
curl "http://localhost:8080/api/wnpp" | jq
curl http://localhost:8080/api/wnpp/count
```

### Pagination
```bash
curl "http://localhost:8080/api/wnpp?limit=2&offset=0"
```

### Ordering
```bash
curl "http://localhost:8080/api/wnpp?order=arrival&limit=2&offset=0"
curl "http://localhost:8080/api/wnpp?order=arrival"
curl "http://localhost:8080/api/wnpp?order=installs"
```

### Filtering by Type
```bash
curl "http://localhost:8080/api/wnpp?type=O&limit=2&offset=0"
curl "http://localhost:8080/api/wnpp/count?type=ITP"
curl "http://localhost:8080/api/wnpp?type=ITP&type=RFP&type=O" | jq
```

### Search Queries
```bash
curl "http://localhost:8080/api/wnpp?q=liba&limit=2&offset=0"
curl "http://localhost:8080/api/wnpp?q=efficient,%20simple%20and%20secure%20bittorrent%20client"
```


### Owner Filter
```bash
curl "http://localhost:8080/api/wnpp?owner=false" | jq
```

## Notes
 - jq is used in examples to pretty-print JSON responses.
 - The API is designed to be flexible: you can combine query parameters like limit, offset, order, type, and q to refine results.
