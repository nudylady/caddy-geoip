package geoip

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func (gip GeoIP) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	gip.addGeoIPHeaders(w, r)
	return gip.Next.ServeHTTP(w, r)
}

var record struct {
	Country struct {
		ISOCode           string            `maxminddb:"iso_code"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union"`
		Names             map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`

	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`

	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
		TimeZone  string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
}

func (gip GeoIP) addGeoIPHeaders(w http.ResponseWriter, r *http.Request) {
	clientIP, _ := getClientIP(r, true)

	err := gip.Config.DBHandler.Lookup(clientIP, &record)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Add(gip.Config.HeaderNameCountryCode, record.Country.ISOCode)
	w.Header().Add(gip.Config.HeaderNameCountryIsEU, strconv.FormatBool(record.Country.IsInEuropeanUnion))
	w.Header().Add(gip.Config.HeaderNameCountryName, record.Country.Names["en"])

	w.Header().Add(gip.Config.HeaderNameCityName, record.City.Names["en"])

	w.Header().Add(gip.Config.HeaderNameLocationLat, strconv.FormatFloat(record.Location.Latitude, 'f', 6, 64))
	w.Header().Add(gip.Config.HeaderNameLocationLon, strconv.FormatFloat(record.Location.Longitude, 'f', 6, 64))
	w.Header().Add(gip.Config.HeaderNameLocationTimeZone, record.Location.TimeZone)
}

func getClientIP(r *http.Request, strict bool) (net.IP, error) {
	var ip string

	// Use the client ip from the 'X-Forwarded-For' header, if available.
	if fwdFor := r.Header.Get("X-Forwarded-For"); fwdFor != "" && !strict {
		ips := strings.Split(fwdFor, ", ")
		ip = ips[0]
	} else {
		// Otherwise, get the client ip from the request remote address.
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return nil, err
		}
	}
	//	ip = "212.50.99.193"
	// Parse the ip address string into a net.IP.
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, errors.New("unable to parse address")
	}

	return parsedIP, nil
}
