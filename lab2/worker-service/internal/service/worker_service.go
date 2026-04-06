package service

import (
	"encoding/json"
	"log"
	"time"

	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/repo"
)

type WorkerService struct {
	deviceRepo *repo.DeviceRepo
	statsRepo  *repo.StatsRepo
}

func New(dr *repo.DeviceRepo, sr *repo.StatsRepo) *WorkerService {
	return &WorkerService{deviceRepo: dr, statsRepo: sr}
}

func (w *WorkerService) Process(body []byte) error {
	var msg model.DeviceMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}

	p := msg.Device
	d := &model.Device{
		ID:         p.ID,
		Name:       p.Name,
		Origin:     p.Origin,
		Price:      p.Price,
		Critical:   p.Critical,
		Peripheral: p.Peripheral,
		PowerWatts: p.PowerWatts,
		HasCooler:  p.HasCooler,
		Group:      p.Group,
		Ports:      p.Ports,
	}

	switch msg.Event {
	case "created", "updated":
		if err := w.deviceRepo.Upsert(d); err != nil {
			return err
		}
	case "deleted":
		if err := w.deviceRepo.Delete(d.ID); err != nil {
			return err
		}
	default:
		log.Printf("unknown event type: %s", msg.Event)
		return nil
	}

	if err := w.statsRepo.IncrementStat(msg.Event, d.Group); err != nil {
		log.Printf("stats increment error: %v", err)
	}
	if err := w.statsRepo.UpdateLastProcessed(time.Now().UTC()); err != nil {
		log.Printf("stats timestamp error: %v", err)
	}
	log.Printf("processed event=%s device_id=%d", msg.Event, d.ID)
	return nil
}
