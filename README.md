# golint-recvnil

Another golang linter that checks that there is a check for nil for the dereferenced receiver in a method.

## Installation

```shell
go install github.com/kogleron/golint-recvnil/cmd/recvnil@latest
```

## How to run

```shell
go vet -vettool=$(which ./bin/recvnil) ./...
```

If there is the ".recvnil.ignore" file then errors from the file will be ignored.

## Flags

- dump-ignore - Dumps errors into '.recvnil.ignore' file.

## TODO

- [x] Show in report the position of a receiver derederencing instead of a method.
- [x] Add ignore list.
- [x] Report about dereferencing function arguments