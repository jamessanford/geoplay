package latlonpb

//go:generate protoc -I . --go_out=plugins=grpc:. latlon.proto
//go:generate protoc -I . --letmegrpc_out=. latlon.proto
