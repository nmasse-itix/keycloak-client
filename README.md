## Basic keycloak client in go

This repo provides a basic keycloak client in go.

## Compile from sources

```sh
GO111MODULE=on go get github.com/golang/mock/mockgen@v1.4.4
cd toolbox
go generate
cd ..
go test ./...
```
