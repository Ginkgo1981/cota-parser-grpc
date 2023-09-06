### Cota Parser API
public API for parsing witness data from Cota


### generate gRPC code
``` protoc --go_out=. api/cotaparser/v1/cota.proto api/cotaparser/v1/entry.proto```

``` protoc --go_out=. --go-grpc_out=. api/cotaparser/v1/cota.proto ```


