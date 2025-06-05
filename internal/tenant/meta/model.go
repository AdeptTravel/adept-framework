// internal/tenant/meta/record.go
//
// `site` table row model.
//
// Context
// -------
// The `Record` struct mirrors one row in the persistent **site** table,
// capturing host routing preferences, theme, DSN, and soft-delete flags.
// It is used by the tenant loader to build the in-memory cache and by
// admin tooling that lists or edits sites.
//
// Schema reference (2025-06-05)
//
//	CREATE TABLE site (
//	    id            INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
//	    host          VARCHAR(256)  NOT NULL UNIQUE,
//	    dsn           VARCHAR(512)  NOT NULL,
//	    theme         VARCHAR(128)  NOT NULL DEFAULT 'base',
//	    locale        VARCHAR(16)   NOT NULL DEFAULT 'en_US',
//	    routing_mode  VARCHAR(6)    NOT NULL DEFAULT 'path',
//	    route_version INT           NOT NULL DEFAULT 0,
//	    preload       TINYINT(1)    NOT NULL DEFAULT 0,
//	    suspended_at  TIMESTAMP NULL,
//	    deleted_at    TIMESTAMP NULL,
//	    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
//	);
//
// Notes
// -----
// • `Preload` is mapped to a Go `bool`; zero = false.
// • Nullable timestamps are `*time.Time`; callers must nil-check before use.
// • `CreatedAt` and `UpdatedAt` are NOT NULL, so plain `time.Time` is safe.
// • This struct contains no behaviour—pure data model for sqlx scans.
package meta

import "time"

// Record mirrors one row in the `site` table.
type Record struct {
	ID           uint64     `db:"id"`
	Host         string     `db:"host"`
	DSN          string     `db:"dsn"`
	Theme        string     `db:"theme"`
	Locale       string     `db:"locale"`
	RoutingMode  string     `db:"routing_mode"`
	RouteVersion int        `db:"route_version"`
	Preload      bool       `db:"preload"`
	SuspendedAt  *time.Time `db:"suspended_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}
