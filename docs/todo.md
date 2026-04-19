---
title: Parking Lot
---

Team IDs shouldn't be case sensitive, i.e

```sh
curl http://localhost:8080/api/v1/teams/sea
```

returns

```json
{
"error": "team season not found"
}
```

Also, why is 2025 missing?

```sh
curl http://localhost:8080/api/v1/teams/SEA?year=2024
```

works because 2024 is the default but 2025 should be populated.

---

2026 should be populated via the stats api.
