package site

// Record mirrors one row from the `site` table.
type Record struct {
	ID     uint64 `db:"id"`
	Host   string `db:"host"`
	DSN    string `db:"dsn"`
	Theme  string `db:"theme"`
	Status string `db:"status"`
}
