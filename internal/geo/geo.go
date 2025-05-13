package geo

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type DB struct {
	mmdb *geoip2.Reader
}

// Open memory-maps the MaxMind database file.
func Open(path string) (*DB, error) {
	m, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &DB{mmdb: m}, nil
}

// Close frees the mmap.
func (db *DB) Close() error { return db.mmdb.Close() }

type Location struct {
	CountryISO string
	Country    string
	City       string
	Lat        float64
	Lon        float64
}

// Lookup returns best-effort location data; zero struct if lookup fails.
func (db *DB) Lookup(ipStr string) (Location, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Location{}, nil
	}
	rec, err := db.mmdb.City(ip)
	if err != nil {
		return Location{}, err
	}
	return Location{
		CountryISO: rec.Country.IsoCode,
		Country:    rec.Country.Names["en"],
		City:       rec.City.Names["en"],
		Lat:        rec.Location.Latitude,
		Lon:        rec.Location.Longitude,
	}, nil
}
