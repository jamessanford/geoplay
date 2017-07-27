package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jamessanford/geoplay/latlonpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var errLocationBad = errors.New("location expects '37.486,-122.232' format")

func runClient(addr string, lat, lon float64) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithBackoffConfig(grpc.DefaultBackoffConfig))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := latlonpb.NewGeoLookupClient(conn)

	r, err := c.Lookup(context.Background(),
		&latlonpb.LookupReq{Lat: lat, Lon: lon},
		grpc.FailFast(false))
	if err != nil {
		log.Fatal(err)
	}

	// This is silly, but the built in proto.MarshalTextString
	// generates octal escapes. See protobuf/text/text.go:func writeString
	// To get valid UTF-8, serialize it ourselves.
	for _, f := range r.Location {
		fmt.Printf("location: <\n  name: %q\n  lat: %f\n  lon: %f\n  distance: %f\n>\n", f.Name, f.Lat, f.Lon, f.Distance)
	}
}

func dealWithLocFlag(location string) (lat float64, lon float64, err error) {
	if location == "" {
		return
	}
	loc := strings.SplitN(location, ",", 2)
	if len(loc) < 2 {
		err = errLocationBad
		return
	}
	slat := strings.TrimSpace(loc[0])
	slon := strings.TrimSpace(loc[1])
	lat, err = strconv.ParseFloat(slat, 64)
	if err != nil {
		return
	}
	lon, err = strconv.ParseFloat(slon, 64)
	if err != nil {
		return
	}
	return lat, lon, err
}
