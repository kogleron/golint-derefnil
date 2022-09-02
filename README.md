# golint-derefnil

Another golang linter that checks that there is a check for nil for the dereferenced receiver in a method.

## Installation

```shell
go install github.com/kogleron/golint-derefnil/cmd/golint-derefnil@latest
```

## How to run

```shell
go vet -vettool=$(which ./bin/golint-derefnil) ./...
```

If there is the ".recvnil.ignore" file then errors from the file will be ignored.

## Flags

- dump-ignore - Dumps errors into '.recvnil.ignore' file.
