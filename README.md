# rototiller

## Developing

### Prerequisites

* golang is *required* - version 1.11.x or above is required for go mod to work
* docker is *required* - version 20.10.x is tested; earlier versions may also work
* docker-compose is *required* - version 1.29.x is tested; earlier versions may also work
* go mod is *required* for dependency management of golang packages
* make is *required* - version 3.81 is tested; earlier versions may also work

### Running

```sh
# setup services
make infra
# run rototiller
make up
# restart rototiller
make restart
```

### Migrations

#### Create Migration

```sh
# generate a migration version
version=`date -u +%Y%m%d%T | tr -cd [0-9]`
touch datastore/psql/migrations/${version}_my_title.up.sql
```

see [Postgres migration tutorial](https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md)

#### Migrate

```sh
# run migrations
rototiller migrate
```

#### Deploy to cluster
```sh
# We use semantic versioning
# Repo: [rototiller]
git tag -a 0.7.0 -m 0.7.0 
git push --follow-tags

# (Only need to do this once) Repo: [rototiller-chart]
helm repo add bitnami https://charts.bitnami.com/bitnami
helm dependency build

# Repo: [rototiller-gitops]
# update tag in values.yaml
make rototiller
git add/commit/push
```

#### Example k8 commands
```sh
# list all namespaces
kubectl get ns
# get specific namespace
kubectl get pod -n <namespace>
# set default namespace
kubectl config set-context --current --namespace <namespace>
# list all secrets in namespace
kubectl get secret -n <namespace>
# get a secret yaml
kubectl get secret -n <namespace> <kind> -o yaml
# decode a password
echo <password> | base64 -d
# get postgres password (default namespace rototiller-gitops must be set)
kubectl get secret -o json rototiller-postgresql | jq -r '.data["postgresql-password"]' | base64 -d
# execute into a container
kubectl exec -n <namespace> <kind> -it -- <cmd> 
```

#### Example API calls

```sh
# create buffer job
curl -X POST -H "Content-Type: application/zip" -H "X-API-Key: cus_LcKO8YPhzJZQgu" --data-binary '@/path/to/a.zip' "https://rototiller.logsquaredn.io/api/v1/jobs/buffer?buffer-distance=5&quadrant-segment-count=50"
# get job result
curl -X GET -H "Content-Type: application/zip" -H "X-API-Key: cus_LcKO8YPhzJZQgu" -o "/path/to/a.zip" "https://rototiller.logsquaredn.io/api/v1/jobs/9b45f141-a137-4f52-a36f-2640129d92e8/output/content"
# create storage
curl -X POST -H "Content-Type: application/zip" -H "X-API-Key: cus_LcKO8YPhzJZQgu" --data-binary '@/path/to/a.zip' "https://rototiller.logsquaredn.io/api/v1/storages?name=<name>"
# create vector lookup job
curl -X POST -H "Content-Type: application/zip" -H "X-API-Key: cus_LcKO8YPhzJZQgu" --data-binary '@/path/to/a.zip' "http://localhost:8080/api/v1/jobs/vectorlookup?attributes=RADII,ADVNUM&latitude=20.33&longitude=-64.23"
```
