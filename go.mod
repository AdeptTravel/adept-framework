module github.com/AdeptTravel/adept-framework

go 1.24

// DELETE the self-require line.  You are the module.
// require github.com/AdeptTravel/adept-framework v0.0.0-00010101000000-000000000000

replace github.com/AdeptTravel/adept-framework => .

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
)

require filippo.io/edwards25519 v1.1.0 // indirect
