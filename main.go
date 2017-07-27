package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jamessanford/geoplay/lookup"

	"go.uber.org/zap"
)

var grpcAddr = flag.String("grpc.addr", ":6797", "grpc server on this address")
var httpAddr = flag.String("http", ":6796", "HTTP server on this address")
var modeCreate = flag.Bool("create", false, "create database and exit")
var modeClient = flag.Bool("client", false, "act as a client")

var dbFile = flag.String("db", "geoplaydata.db", "store cached buntdb here")
var locFile = flag.String("locations", "data/locations.json", "file with json locations")
var locFlag = flag.String("loc", "", "lat,lon for client mode")

var logger *zap.Logger

func main() {
	flag.Parse()
	z, err := zap.NewDevelopment()
	logger = z // package level logger
	if err != nil {
		log.Fatalf("zap logger failed: %v", err)
	}
	defer logger.Sync()

	if *modeCreate {
		if *dbFile == "" || *locFile == "" {
			logger.Error("Must specify -locations= and -db=")
			os.Exit(1)
		}
		logger.Info("creating buntDB",
			zap.String("db", *dbFile),
			zap.String("locations", *locFile))
		start := time.Now()
		if err = lookup.CreateBuntDB(*dbFile, *locFile); err != nil {
			logger.Fatal("failed to create", zap.Error(err))
		}
		fmt.Printf("create took %v\n", time.Since(start))
		return
	}
	if *modeClient {
		logger.Info("begin client", zap.String("dial", *grpcAddr))
		lat, lon, err := dealWithLocFlag(*locFlag)
		if err != nil {
			logger.Fatal("invalid location", zap.Error(err))
		}
		runClient(*grpcAddr, lat, lon)
		return
	}

	runServer()
}
