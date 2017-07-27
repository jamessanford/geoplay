/*
geoplay is a toy grpc server and client. It returns up to 10 nearest Starbucks locations when given a latitude/longitude.

CLIENT MODE

When used with the -client flag, it dials the server and displays results
on standard output.

	geoplay -client
	geoplay -client -loc=37.486,-122.232 # Redwood City, CA
	geoplay -client -loc=37,122 # China

SERVER MODE

The default mode is an RPC server mode.

	geoplay

Try the LetMeGRPC handler at http://localhost:6796/GeoLookup/Lookup or the built-in Prometheus metrics at http://localhost:6796/metrics. After making some RPCs, look at http://localhost:6796/debug/requests.

INCLUDED DATA

The data/locations.json file came from https://github.com/mmcloughlin/starbucks
*/
package main
