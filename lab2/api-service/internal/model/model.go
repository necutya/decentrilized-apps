package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringSlice is a []string that serializes to/from JSON for SQLite storage.
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

type DeviceType struct {
	Peripheral  bool        `json:"peripheral"`
	PowerWatts  int32       `json:"power_watts"`
	HasCooler   bool        `json:"has_cooler"`
	Group       string      `json:"group"`
	Ports       StringSlice `json:"ports" gorm:"type:text"`
}

type Device struct {
	ID         uint       `gorm:"primaryKey;autoIncrement"`
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

func (d *Device) GetDeviceType() DeviceType {
	return DeviceType{
		Peripheral: d.Peripheral,
		PowerWatts: d.PowerWatts,
		HasCooler:  d.HasCooler,
		Group:      d.Group,
		Ports:      d.Ports,
	}
}
