package repo

import (
	"time"

	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

// IncrementStat upserts the (eventType, group) row, incrementing count.
func (r *StatsRepo) IncrementStat(eventType, group string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var stat model.EventStat
		err := tx.Where("event_type = ? AND `group` = ?", eventType, group).First(&stat).Error
		if err == gorm.ErrRecordNotFound {
			stat = model.EventStat{EventType: eventType, Group: group, Count: 1}
			return tx.Create(&stat).Error
		} else if err != nil {
			return err
		}
		return tx.Model(&stat).Update("count", stat.Count+1).Error
	})
}

// UpdateLastProcessed sets the single-row timestamp.
func (r *StatsRepo) UpdateLastProcessed(t time.Time) error {
	lp := model.LastProcessed{ID: 1, ProcessedAt: t}
	return r.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&lp).Error
}

type Summary struct {
	TotalCreated    int64
	TotalUpdated    int64
	TotalDeleted    int64
	ByGroup         map[string]int64
	LastProcessedAt time.Time
}

func (r *StatsRepo) GetSummary() (*Summary, error) {
	var stats []model.EventStat
	if err := r.db.Find(&stats).Error; err != nil {
		return nil, err
	}

	s := &Summary{ByGroup: make(map[string]int64)}
	for _, st := range stats {
		switch st.EventType {
		case "created":
			s.TotalCreated += st.Count
		case "updated":
			s.TotalUpdated += st.Count
		case "deleted":
			s.TotalDeleted += st.Count
		}
		if st.Group != "" {
			s.ByGroup[st.Group] += st.Count
		}
	}

	var lp model.LastProcessed
	if err := r.db.First(&lp, 1).Error; err == nil {
		s.LastProcessedAt = lp.ProcessedAt
	}
	return s, nil
}
