**Toy grpc server and client in Go.**

Toy server accepts RPCs with latitude/longitude and returns the 10 nearest Starbucks locations.

#### Install, run, request

```
go get -u github.com/jamessanford/geoplay
geoplay &
geoplay -client -loc=37.486,-122.232
```

#### Included endpoints

A letmegrpc HTTP handler is at http://localhost:6796/GeoLookup/Lookup

Prometheus metrics with `go-grpc-prometheus` support are at http://localhost:6796/metrics

grpc is on port 6797 by default (-grpc.listen flag)

grpc uses [x/net/trace](https://godoc.org/golang.org/x/net/trace) by default, check out http://localhost:6796/debug/requests and http://localhost:6796/debug/events after making some RPCs.

grpc [reflection](https://godoc.org/google.golang.org/grpc/reflection) is enabled, try grpc_cli:

```
grpc_cli call localhost:6797 latlonpb.GeoLookup.Lookup 'lat:37.486 lon:-122.232'
```

#### Suggestions

Instead of "rolling your own" this way, there are various "microservice helpers", and you probably want to use one or make a standard one for your organization.

-	https://github.com/google/go-microservice-helpers
-	https://github.com/go-kit/kit

The `data/locations.json` file came from https://github.com/mmcloughlin/starbucks
