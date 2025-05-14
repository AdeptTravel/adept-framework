package widgetinstance

import "time"

// WidgetInstance mirrors the DB table.
type WidgetInstance struct {
	ID       int64
	SiteID   string
	Area     string
	WidgetID string
	Ordering int
	Config   string // raw JSON
	Enabled  bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
