# Elasticsearch Snapshots Cleanup


## Build and usage

```
go build
./es-snapshots-cleanup
```
- List snapshots on elasticsearch cluster
- Discover snapshots repositories and pick the first found
- Clean snapshots older than x days
- Doesn't clean the very last snapshot stored (in case backups gone wrong)
- Make use of env variables for easier usage in cloud infrastructures

## Deploy on docker / k8s

Features a Dockerfile for containerized deployments and a cronjob object for Kubernetes clusters.

Get latest image on [Docker Hub](https://hub.docker.com/r/novitnc/es-snapshot-cleanup).

```
docker pull novitnc/es-snapshot-cleanup:latest
```

## Env

| VAR | DEFAULT | ROLE |
|-----|---------|------|
| ES_ENDPOINT | es:9200 | ES API in format hostname:port |
| ES_REPO | (autodiscovered) | Explicit ES snapshot repository to request |
| ES_CLEAN_AFTER_DAYS | 14 | Clean snapshots after X days |
| ES_DRY_RUN | (empty) | When not null, dry run mode is enabled (no deletion occurs) |

For example :

```
env ES_ENDPOINT=elasticsearch:9200 ES_CLEAN_AFTER_DAYS=7 ./es-snapshots-cleanup
```
