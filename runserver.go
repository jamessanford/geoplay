package main

import (
	"context"
	"flag"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/jamessanford/geoplay/latlonpb"
	"github.com/jamessanford/geoplay/lookup"
	"github.com/jamessanford/geoplay/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var useHistogram = flag.Bool("histogram", false, "enable GRPC Prometheus histogram metrics")

const httpBanner = `geoplay - <a href="/GeoLookup/Lookup">/GeoLookup/Lookup</a>`
const httpShutdownTimeout = time.Second * 10

func registerSignalHandler(server *grpc.Server, httpser *http.Server) {
	ch := make(chan os.Signal, 1)
	go func() {
		called := false
		for {
			<-ch
			if called {
				logger.Fatal("forcing shutdown")
			}
			called = true
			logger.Info("server shutdown requested")
			server.GracefulStop()
			d := time.Now().Add(httpShutdownTimeout)
			ctx, cancel := context.WithDeadline(context.Background(), d)
			httpser.Shutdown(ctx)
			cancel() // not really needed, won't do anything here
		}
	}()
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}

// registerLetMeGRPC adds a HTTP endpoint at /GeoLookup/Lookup
func registerLetMeGRPC() {
	h, err := latlonpb.NewHandler(*grpcAddr,
		latlonpb.DefaultHtmlStringer,
		grpc.WithInsecure())
	if err != nil {
		logger.Error("letmegrpc handler failed", zap.Error(err))
		return
	}
	http.Handle("/GeoLookup/", h)
}

func runServer() {
	lu, err := lookup.TryOpenDB(*dbFile, *locFile)
	if err != nil {
		logger.Fatal("unable to open db", zap.Error(err))
	}
	defer lu.Close()

	logger.Info("grpc server starting", zap.String("listen", *grpcAddr))
	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		logger.Fatal("unable to start server", zap.String("listen", *grpcAddr), zap.Error(err))
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor))
	ll := &server.GeoLookup{Logger: logger, Search: lu}
	latlonpb.RegisterGeoLookupServer(s, ll)

	if *useHistogram {
		grpc_prometheus.EnableHandlingTimeHistogram()
	}
	grpc_prometheus.Register(s)

	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, httpBanner)
	})
	http.Handle("/metrics", promhttp.Handler())

	httpsrv := &http.Server{Addr: *httpAddr}
	registerSignalHandler(s, httpsrv)
	registerLetMeGRPC()

	// If either server exits, we consider ourselves failed.
	// Try to exit cleanly so that the defers get run.
	done := make(chan struct{})

	go func() {
		logger.Info("http server starting", zap.String("http", *httpAddr))
		err = httpsrv.ListenAndServe()
		logger.Error("http server exited", zap.Error(err))
		done <- struct{}{}
	}()

	reflection.Register(s) // enable grpc_cli, call this last
	go func() {
		err = s.Serve(lis)
		logger.Error("grpc server exited", zap.Error(err))
		done <- struct{}{}
	}()

	<-done // exit when ANY of the servers die
}
