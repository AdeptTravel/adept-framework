package site

import "time"

// Record mirrors one row in the persistent `site` table.  The operational
// state is captured by two nullable timestamps:
//
//   - SuspendedAt – site is temporarily disabled (e.g., billing).
//   - DeletedAt   – site is permanently removed.
//
// Either timestamp being non-NULL prevents the lazy-loader from serving the
// site.
type Record struct {
	ID          uint64     `db:"id"`
	Host        string     `db:"host"`
	DSN         string     `db:"dsn"`
	Theme       string     `db:"theme"`
	Title       string     `db:"title"`
	Locale      string     `db:"locale"`
	SuspendedAt *time.Time `db:"suspended_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
