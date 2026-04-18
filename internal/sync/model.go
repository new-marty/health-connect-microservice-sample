package sync

type SyncLog struct {
	ID             int64   `json:"id"`
	Source         string  `json:"source"`
	SyncedAt       string  `json:"synced_at"`
	RecordsAdded   int     `json:"records_added"`
	RecordsUpdated int     `json:"records_updated"`
	Status         string  `json:"status"`
	ErrorMsg       *string `json:"error_msg"`
}

type SyncStatus struct {
	Source       string  `json:"source"`
	LastSyncAt   *string `json:"last_sync_at"`
	LastStatus   *string `json:"last_status"`
	RecordsAdded int     `json:"records_added"`
}
