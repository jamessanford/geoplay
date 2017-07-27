package server

import (
	"github.com/jamessanford/geoplay/latlonpb"
	"github.com/jamessanford/geoplay/lookup"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	grpc_codes "google.golang.org/grpc/codes"
	grpc_status "google.golang.org/grpc/status"
)

const defaultMaxResults = 10

// GeoLookup is the primary GRPC service from latlonpb
type GeoLookup struct {
	Logger *zap.Logger
	Search *lookup.DB
}

// Lookup returns the nearest locations according to input lat/lon
func (s *GeoLookup) Lookup(ctx context.Context, in *latlonpb.LookupReq) (*latlonpb.LookupResp, error) {
	res, err := s.Search.FindNearest(in.Lat, in.Lon, defaultMaxResults)
	if err != nil {
		s.Logger.Error("FindNearest failed",
			zap.Error(err),
			zap.Float64("Lat", in.Lat),
			zap.Float64("Lon", in.Lon),
			zap.Int("Max", defaultMaxResults),
		)
		return &latlonpb.LookupResp{},
			grpc_status.Errorf(grpc_codes.Unavailable, "%s", err)
	}

	// NOTE: avoid boilerplate and let FindNearest use latlonpb directly?
	var loc []*latlonpb.Location
	for _, r := range res {
		loc = append(loc, &latlonpb.Location{
			Name:     r.Name,
			Lat:      r.Lat,
			Lon:      r.Lon,
			Distance: r.Distance,
		})
	}
	return &latlonpb.LookupResp{Location: loc}, nil
}
