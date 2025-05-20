// Command web boots the framework inside a FreeBSD jail.
//
// Flags:
//
//	-host   host name to serve (overrides SITE_HOST env)   default ""
//	-addr   listen address                                 default ":8080"
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/AdeptTravel/adept-framework/internal/database"
	"github.com/AdeptTravel/adept-framework/internal/site"
)

const (
	envFile = "/usr/local/etc/adept-framework/global.env"
	logDir  = "log"                     // ./log/YYYY-MM-DD.log
	geoCity = "data/GeoLite2-City.mmdb" // hard-coded for now
)

func init() {
	// Load env file for local jails.  In production you can omit the file
	// and inject variables directly through rc.d or jail.conf.
	_ = godotenv.Load(envFile)
}

func main() {
	hostFlag := flag.String("host", "", "virtual host served by this instance")
	addrFlag := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	host := firstNonEmpty(*hostFlag, os.Getenv("SITE_HOST"))
	if host == "" {
		log.Fatal("no site host specified (flag -host or SITE_HOST env)")
	}

	globalDSN := os.Getenv("GLOBAL_DB_DSN")
	if globalDSN == "" {
		log.Fatal("GLOBAL_DB_DSN is not set")
	}

	// Connect to control-plane database.
	globalDB, err := database.Open(globalDSN)
	if err != nil {
		log.Fatalf("connect global DB: %v", err)
	}
	defer globalDB.Close()

	// Fetch site record by host.
	siteRec, err := site.ByHost(globalDB, host)
	if err != nil {
		log.Fatalf("site lookup failed for %s: %v", host, err)
	}

	// Open tenant database.
	siteDB, err := database.Open(siteRec.DSN)
	if err != nil {
		log.Fatalf("connect site DB: %v", err)
	}
	defer siteDB.Close()

	log.Printf("serving %s (%s) on %s", siteRec.Host, siteRec.Theme, *addrFlag)

	// TODO: attach router, middlewares, module loader, etc.
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	if err := http.ListenAndServe(*addrFlag, nil); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

func firstNonEmpty(v1, v2 string) string {
	if v1 != "" {
		return v1
	}
	return v2
}
