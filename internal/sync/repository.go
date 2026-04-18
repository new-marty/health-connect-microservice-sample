package sync

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Log(ctx context.Context, source string, added, updated int, status string, errMsg *string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO sync_log
		(source, records_added, records_updated, status, error_msg)
		VALUES (?,?,?,?,?)`,
		source, added, updated, status, errMsg)
	return err
}

func (r *Repository) GetStatus(ctx context.Context) ([]SyncStatus, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT source,
		MAX(synced_at) AS last_sync_at,
		(SELECT status FROM sync_log s2 WHERE s2.source = s1.source ORDER BY synced_at DESC LIMIT 1) AS last_status,
		COALESCE(SUM(records_added), 0) AS total_added
		FROM sync_log s1
		GROUP BY source
		ORDER BY source`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []SyncStatus
	for rows.Next() {
		var s SyncStatus
		if err := rows.Scan(&s.Source, &s.LastSyncAt, &s.LastStatus, &s.RecordsAdded); err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}
	return statuses, rows.Err()
}
