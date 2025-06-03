module github.com/yanizio/adept

go 1.24

// DELETE the self-require line.  You are the module.
// require github.com/yaniz.io/adept v0.0.0-00010101000000-000000000000

replace github.com/yanizio/adept => .

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/prometheus/client_golang v1.22.0
	golang.org/x/sync v0.14.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/avct/uasurfer v0.0.0-20250506104815-f2613aa2d406 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/mssola/user_agent v0.6.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oschwald/geoip2-golang v1.8.0 // indirect
	github.com/oschwald/maxminddb-golang v1.10.0 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	golang.org/x/sys v0.30.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)
