package sqlite

import (
	"database/sql"

	"github.com/kiddyt00/yiguan/internal/store"
)

func (s *Store) ListAds() ([]store.Ad, error) {
	rows, err := s.db.Query("SELECT id, name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order, created_at FROM ads ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Ad
	for rows.Next() {
		var a store.Ad
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.AdType, &a.ContentURL, &a.WatchDuration, &a.RewardQuota, &a.IsEnabled, &a.SortOrder, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) ListActiveAds() ([]store.Ad, error) {
	rows, err := s.db.Query("SELECT id, name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order, created_at FROM ads WHERE is_enabled = 1 ORDER BY sort_order, id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []store.Ad
	for rows.Next() {
		var a store.Ad
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.AdType, &a.ContentURL, &a.WatchDuration, &a.RewardQuota, &a.IsEnabled, &a.SortOrder, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (s *Store) GetAdByID(id int64) (*store.Ad, error) {
	a := &store.Ad{}
	err := s.db.QueryRow(
		"SELECT id, name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order, created_at FROM ads WHERE id = ?",
		id,
	).Scan(&a.ID, &a.Name, &a.Description, &a.AdType, &a.ContentURL, &a.WatchDuration, &a.RewardQuota, &a.IsEnabled, &a.SortOrder, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return a, err
}

func (s *Store) CreateAd(ad *store.Ad) error {
	result, err := s.db.Exec(
		"INSERT INTO ads (name, description, ad_type, content_url, watch_duration, reward_quota, is_enabled, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		ad.Name, ad.Description, ad.AdType, ad.ContentURL, ad.WatchDuration, ad.RewardQuota, ad.IsEnabled, ad.SortOrder,
	)
	if err != nil {
		return err
	}
	ad.ID, _ = result.LastInsertId()
	return nil
}

func (s *Store) UpdateAd(ad *store.Ad) error {
	_, err := s.db.Exec(
		"UPDATE ads SET name = ?, description = ?, ad_type = ?, content_url = ?, watch_duration = ?, reward_quota = ?, is_enabled = ?, sort_order = ? WHERE id = ?",
		ad.Name, ad.Description, ad.AdType, ad.ContentURL, ad.WatchDuration, ad.RewardQuota, ad.IsEnabled, ad.SortOrder, ad.ID,
	)
	return err
}

func (s *Store) DeleteAd(id int64) error {
	_, err := s.db.Exec("DELETE FROM ads WHERE id = ?", id)
	return err
}

func (s *Store) ToggleAd(id int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := s.db.Exec("UPDATE ads SET is_enabled = ? WHERE id = ?", v, id)
	return err
}

func (s *Store) CreateAdRecord(rec *store.AdRecord) error {
	result, err := s.db.Exec(
		"INSERT INTO ad_records (user_id, ad_id, watch_duration, status, rewarded) VALUES (?, ?, ?, ?, ?)",
		rec.UserID, rec.AdID, rec.WatchDuration, rec.Status, rec.Rewarded,
	)
	if err != nil {
		return err
	}
	rec.ID, _ = result.LastInsertId()
	return nil
}

func (s *Store) UpdateAdRecord(rec *store.AdRecord) error {
	_, err := s.db.Exec(
		"UPDATE ad_records SET watch_duration = ?, status = ?, rewarded = ? WHERE id = ?",
		rec.WatchDuration, rec.Status, rec.Rewarded, rec.ID,
	)
	return err
}

func (s *Store) GetAdRecord(userID, adID int64) (*store.AdRecord, error) {
	r := &store.AdRecord{}
	err := s.db.QueryRow(
		"SELECT id, user_id, ad_id, watch_duration, status, rewarded, created_at FROM ad_records WHERE user_id = ? AND ad_id = ? ORDER BY id DESC LIMIT 1",
		userID, adID,
	).Scan(&r.ID, &r.UserID, &r.AdID, &r.WatchDuration, &r.Status, &r.Rewarded, &r.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	return r, err
}

func (s *Store) GetAdStats() ([]store.AdStat, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.name,
			   COUNT(ar.id),
			   SUM(CASE WHEN ar.status = 'completed' THEN 1 ELSE 0 END),
			   SUM(CASE WHEN ar.rewarded = 1 THEN 1 ELSE 0 END)
		FROM ads a LEFT JOIN ad_records ar ON a.id = ar.ad_id
		GROUP BY a.id ORDER BY a.sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []store.AdStat
	for rows.Next() {
		var st store.AdStat
		if err := rows.Scan(&st.AdID, &st.AdName, &st.Total, &st.Completed, &st.RewardTotal); err != nil {
			return nil, err
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}
