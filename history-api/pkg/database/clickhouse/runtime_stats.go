package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/runtime-radar/runtime-radar/history-api/pkg/database"
	"github.com/runtime-radar/runtime-radar/history-api/pkg/model"
	"gorm.io/gorm"
)

type StatsRepository interface {
	CountEvents(ctx context.Context, from, to time.Time, filter any) (int, error)
	GetDetectorRating(ctx context.Context, from, to time.Time, filters, order any, limit int) ([]*model.DetectorCounter, error)
}

type StatsDatabase struct {
	*gorm.DB
}

func (s *StatsDatabase) CountEvents(ctx context.Context, from, to time.Time, filter any) (int, error) {
	if reason, ok := validateTimePeriod(from, to); !ok {
		return 0, fmt.Errorf("invalid time period: %s", reason)
	}

	if filter == nil {
		filter = ""
	}

	qb := s.WithContext(ctx).
		Model(&model.RuntimeEvent{}).
		Where(filter)

	qb = applyTimePeriod(qb, "registered_at", from, to)

	var cnt int64
	err := qb.Count(&cnt).Error

	return int(cnt), err
}

func (s *StatsDatabase) GetDetectorRating(ctx context.Context, from, to time.Time, filter, order any, limit int) ([]*model.DetectorCounter, error) {
	if reason, ok := validateTimePeriod(from, to); !ok {
		return nil, fmt.Errorf("invalid time period: %s", reason)
	}
	sanitizedOrder, err := database.SanitizeOrder(order)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		filter = ""
	}

	results := []*model.DetectorCounter{}
	qb := s.WithContext(ctx).
		Model(&model.RuntimeEvent{}).
		Where(filter).
		Select("detector_id, threats_severities[indexOf(threats_detectors, detector_id)] as severity, COUNT(*) as count").
		Joins("ARRAY JOIN threats_detectors as detector_id").
		Group("detector_id,severity").
		Order(sanitizedOrder)

	if limit > 0 {
		qb = qb.Limit(limit)
	}

	qb = applyTimePeriod(qb, "registered_at", from, to)

	err = qb.Scan(&results).Error

	return results, err
}

func validateTimePeriod(from, to time.Time) (reason string, ok bool) {
	if from.IsZero() {
		return "from is not set", false
	}
	if to.IsZero() {
		return "to is not set", false
	}
	if from.After(to) {
		return "from is after to", false
	}
	return "", true
}

// applyTimePeriod returns db with time period added to it. It only applies from/to if it's not zero-valued.
func applyTimePeriod(db *gorm.DB, field string, from, to time.Time) *gorm.DB {
	if !from.IsZero() {
		expr := fmt.Sprintf("%s > toDateTime64(?, 9, ?)", field)
		// time is converted to UTC explicitly, because time.Now() returns timestamp with "Local" location unless timezone is configured explicitly.
		// "Local" is not accepted by Clickhouse.
		db = db.Where(expr, from.UTC().Format(DateTimeFormat), from.UTC().Location().String())
	}

	if !to.IsZero() {
		expr := fmt.Sprintf("%s < toDateTime64(?, 9, ?)", field)
		// time is converted to UTC explicitly, because time.Now() returns timestamp with "Local" location unless timezone is configured explicitly.
		// "Local" is not accepted by Clickhouse.
		db = db.Where(expr, to.UTC().Format(DateTimeFormat), to.UTC().Location().String())
	}

	return db
}
