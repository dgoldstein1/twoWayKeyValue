# Two Way Key Value Store

Web server for easy look up of value -> key and key -> value, for integration with [big-data graph DB](https://github.com/dgoldstein1/graphApi). Uses [BadgerDB](https://github.com/dgraph-io/badger) as underlying store.

- Maintainability
- Test
- CircleCI

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
export GRAPH_DB_STORE_DIR="/tmp/twowaykv"
export GRAPH_DB_STORE_PORT="5001"
./twowaykv
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

## Authors

* **David Goldstein** - [DavidCharlesGoldstein.com](http://www.davidcharlesgoldstein.com/?github-two-way-kv) - [Decipher Technology Studios](http://deciphernow.com/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
