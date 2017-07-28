package lookup

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kellydunn/golang-geo"
	"github.com/tidwall/buntdb"
)

// Location holds one positions data as returned from the DB.
type Location struct {
	Name     string
	Lat      float64
	Lon      float64
	Distance float64
}

type Locations []*Location

func (s Locations) Len() int      { return len(s) }
func (s Locations) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByDistance is a convenience type for sorting locations by their distance.
type ByDistance struct{ Locations }

func (s ByDistance) Less(i, j int) bool { return s.Locations[i].Distance < s.Locations[j].Distance }

const initialRadius = 5.0 // kilometers
const maxRadius = 1000.0

// FindNearest returns up to maxResults near lat/lon.  Naive implementation.
func (lu *DB) FindNearest(lat, lon float64, maxResults int) (Locations, error) {
	if lat == 0.0 && lon == 0.0 {
		lat, lon = 37.486, -122.232 // default Redwood City, CA
	}
	dest := geo.NewPoint(lat, lon)
	found := make(Locations, 0, maxResults)

	// actually searches the square that fits inside circle of radius
	radius := initialRadius
	for ; radius < maxRadius && len(found) < maxResults; radius *= 2 {
		found = found[:0] // erase results from previous pass

		ul := dest.PointAtDistanceAndBearing(radius, 360.0-45.0)
		lr := dest.PointAtDistanceAndBearing(radius, 90.0+45.0)

		err := lu.db.View(func(tx *buntdb.Tx) error {
			// buntdb uses lon/lat
			bound := fmt.Sprintf("[%f %f],[%f %f]",
				ul.Lng(),
				ul.Lat(),
				lr.Lng(),
				lr.Lat())
			return tx.Intersects("store", bound, func(key, val string) bool {
				res := strings.TrimSuffix(strings.TrimPrefix(key, "store:"), ":pos")
				r, _ := buntdb.IndexRect(val)
				dist := dest.GreatCircleDistance(geo.NewPoint(r[1], r[0]))
				found = append(found, &Location{
					Name:     res,
					Lat:      r[1],
					Lon:      r[0],
					Distance: dist})
				return true
			})
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Sort(ByDistance{found})
	if len(found) > maxResults {
		found = found[:maxResults]
	}
	return found, nil
}
