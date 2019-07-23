# twowaykv

Web server for easy look up of value -> key and key -> value, for integration with [big-data graph DB](https://github.com/dgoldstein1/graphApi). Uses [BadgerDB](https://github.com/dgraph-io/badger) as underlying store.

[![CircleCI](https://circleci.com/gh/dgoldstein1/twowaykv.svg?style=svg)](https://circleci.com/gh/dgoldstein1/twowaykv)
[![Maintainability](https://api.codeclimate.com/v1/badges/6577886aa2f88c77bfc2/maintainability)](https://codeclimate.com/github/dgoldstein1/twowaykv/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/6577886aa2f88c77bfc2/test_coverage)](https://codeclimate.com/github/dgoldstein1/twowaykv/test_coverage)

## Install

```sh
go install github.com/dgoldstein1/twowaykv
```

or

```sh
docker pull dgoldstein1/twowaykv:latest
```


## Run it

```sh
export GRAPH_DB_STORE_DIR="/tmp/twowaykv" # storage directory
export GRAPH_DB_STORE_PORT="5001" # port served on
export GRAPH_DOCS_DIR="./api/*" # location of docs (warning: this entire dir is served up to the browser)
./twowaykv server
```


## Development

#### Local Development

- Install [inotifywait](https://linux.die.net/man/1/inotifywait)
```sh
./watch_dev_changes.sh
```

#### Testing

```sh
go test $(go list ./... | grep -v /vendor/)
```

## Generating New Documentation

```sh
pip install PyYAML
python api/swagger-yaml-to-html.py < api/swagger.yml > api/index.html
```


## Authors

* **David Goldstein** - [DavidCharlesGoldstein.com](http://www.davidcharlesgoldstein.com/?github-two-way-kv) - [Decipher Technology Studios](http://deciphernow.com/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
