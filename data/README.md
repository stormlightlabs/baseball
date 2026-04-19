# Legacy data fallback

Local ETL now prefers `tools/data` (or `--data-root` / `BASEBALL_DATA_ROOT`).

This `data/` tree remains as a backward-compat fallback during migration.
Persisted snapshots should live in the standalone `baseball-data` repository.
