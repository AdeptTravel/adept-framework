# adept-framework# Clone and build
git clone https://github.com/AdeptTravel/adept-framework.git
cd adept-framework
go mod tidy
go build ./cmd/web

# Set environment (development)
cat > .env <<EOF
GLOBAL_DB_DSN=adept:pass@tcp(127.0.0.1:3306)/adept_global
TENANT_IDLE_TTL=30m
TENANT_CACHE_MAX=100
EOF

# Run with an example site row already present
go run ./cmd/web
