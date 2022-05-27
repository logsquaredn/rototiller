# geocloud

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
# run geocloud
make up
# restart geocloud
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
geocloud migrate
```

curl -X POST -H "Content-Type: application/zip" -H "X-API-Key: cus_LcKO8YPhzJZQgu" --data-binary '@/home/phish3y/Documents/input/hurricane.zip' "https://geocloud.logsquaredn.io/api/v1/job/buffer?buffer-distance=5&quadrant-segment-count=50"
curl -X GET -H "Content-Type: application/zip" -H "X-API-Key: cus_LcKO8YPhzJZQgu" -o "/home/phish3y/Downloads/output.zip" "https://geocloud.logsquaredn.io/api/v1/job/9b45f141-a137-4f52-a36f-2640129d92e8/output/content"