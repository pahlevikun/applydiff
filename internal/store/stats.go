package store

import (
	"context"

	"github.com/pahlevikun/applydiff/internal/application"
)

// ChannelStats aggregates application metrics grouped by channel.
func (s *Store) ChannelStats(ctx context.Context) ([]application.ChannelStats, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT channel,
		       COUNT(*) AS total,
		       SUM(CASE WHEN status = 'followup' THEN 1 ELSE 0 END) AS replies,
		       SUM(CASE WHEN status IN ('shortlisted','tailoring','ready','submitted','followup') THEN 1 ELSE 0 END) AS interviews
		FROM applications
		GROUP BY channel
		ORDER BY channel ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []application.ChannelStats
	for rows.Next() {
		var ch string
		var total, replies, interviews int
		if err := rows.Scan(&ch, &total, &replies, &interviews); err != nil {
			return nil, err
		}
		rate := 0.0
		if total > 0 {
			rate = float64(interviews) / float64(total)
		}
		out = append(out, application.ChannelStats{
			Channel:       ch,
			Total:         total,
			Replies:       replies,
			InterviewRate: rate,
		})
	}
	return out, rows.Err()
}
