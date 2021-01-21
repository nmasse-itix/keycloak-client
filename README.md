# Golang Keycloak REST Client

This Golang library provides Types and Methods to drive a Keycloak instance through its REST Admin interface.

## Supported Features

* **Realms**: CRUD, Export, Import
* **Clients**: CRU
* **Users**: CRUD
* **Components**: CRUD

## Hello, World example

```go
package main

import (
	"fmt"
	"log"
	"time"

	keycloak "github.com/nmasse-itix/keycloak-client"
)

func main() {
	config := keycloak.Config{
		AddrTokenProvider: "http://localhost:8080/auth/realm/master",
		AddrAPI:           "http://localhost:8080/auth",
		Timeout:           10 * time.Second,
	}

	client, err := keycloak.NewClient(config)
	if err != nil {
		log.Fatalf("could not create keycloak client: %v", err)
	}

	accessToken, err := client.GetToken("master", "admin", "admin")
	if err != nil {
		log.Fatalf("could not get access token: %v", err)
	}

	realms, err := client.GetRealms(accessToken)
	if err != nil {
		log.Fatalf("could not get realms: %v", err)
	}

	for _, realm := range realms {
		fmt.Println(*realm.Realm)
	}
}
```
