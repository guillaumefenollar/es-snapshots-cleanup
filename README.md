# Elasticsearch Snapshots Cleanup

## Usage

./es-snapshots-cleanup

- List snapshots on elasticsearch cluster
- Clean snapshots older than x days
- Doesn't clean the very last snapshot stored (in case backups gone wrong)
- Make use of env variables for easier usage in cloud infrastructures

## Env

| VAR | DEFAULT | ROLE |
|-----|---------|------|
| ES_ENDPOINT | es:9200 | ES API in format hostname:port |
| ES_REPO | default | ES snapshot repository to request |
| ES_CLEAN_AFTER_DAYS | 7 | Clean snapshots after X days |
| ES_DRY_RUN | <empty> | When not null, dry run mode is enabled (no deletion occurs) |


