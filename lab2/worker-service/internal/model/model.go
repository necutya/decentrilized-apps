package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(src any) error {
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal([]byte(raw), s)
}

type Device struct {
	ID         uint    `gorm:"primaryKey"`
	Name       string
	Origin     string
	Price      float64
	Critical   bool
	Peripheral bool
	PowerWatts int32
	HasCooler  bool
	Group      string
	Ports      StringSlice `gorm:"type:text"`
}

// EventStat tracks (event_type, group) -> count.
type EventStat struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	EventType string `gorm:"uniqueIndex:idx_event_group"`
	Group     string `gorm:"uniqueIndex:idx_event_group"`
	Count     int64
}

// LastProcessed is a single-row settings table.
type LastProcessed struct {
	ID          uint `gorm:"primaryKey"`
	ProcessedAt time.Time
}

// DeviceMessage is the JSON envelope received from RabbitMQ.
type DeviceMessage struct {
	Event  string         `json:"event"`
	Device DevicePayload  `json:"device"`
}

type DevicePayload struct {
	ID         uint        `json:"ID"`
	Name       string      `json:"Name"`
	Origin     string      `json:"Origin"`
	Price      float64     `json:"Price"`
	Critical   bool        `json:"Critical"`
	Peripheral bool        `json:"Peripheral"`
	PowerWatts int32       `json:"PowerWatts"`
	HasCooler  bool        `json:"HasCooler"`
	Group      string      `json:"Group"`
	Ports      StringSlice `json:"Ports"`
}
