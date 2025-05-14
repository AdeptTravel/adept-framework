package widgetinstance

import "github.com/AdeptTravel/adept-framework/internal/data"

// DAO wraps queries for widget_instance.
type DAO struct{ db *data.DB }

func New(db *data.DB) *DAO { return &DAO{db: db} }

// ListBySiteArea returns enabled instances ordered for rendering.
func (d *DAO) ListBySiteArea(siteID, area string) ([]WidgetInstance, error) {
	rows, err := d.db.Query(`
	    SELECT id, site_id, area, widget_id, ordering, config, enabled,
	           created_at, updated_at
	      FROM widget_instance
	     WHERE site_id = ? AND area = ? AND enabled = 1
	     ORDER BY ordering`, siteID, area)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []WidgetInstance
	for rows.Next() {
		var wi WidgetInstance
		if err := rows.Scan(&wi.ID, &wi.SiteID, &wi.Area, &wi.WidgetID,
			&wi.Ordering, &wi.Config, &wi.Enabled, &wi.CreatedAt, &wi.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, wi)
	}
	return out, rows.Err()
}
