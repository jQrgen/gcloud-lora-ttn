package ttn

import (
	"bytes"
	"cloud.google.com/go/bigquery"
	"encoding/base64"
	"encoding/json"
	"github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	"time"
)

// UplinkMessage json struct that represents TTN Uplink message
// Autogenerated using https://mholt.github.io/json-to-go/
type UplinkMessage struct {
	AppID          string `json:"app_id"`
	DevID          string `json:"dev_id"`
	HardwareSerial string `json:"hardware_serial"`
	Port           int    `json:"port"`
	Counter        int    `json:"counter"`
	IsRetry        bool   `json:"is_retry"`
	Confirmed      bool   `json:"confirmed"`
	PayloadRaw     string `json:"payload_raw"`
	PayloadFields  struct {
	} `json:"payload_fields"`
	Metadata struct {
		Time       time.Time `json:"time"`
		Frequency  float64   `json:"frequency"`
		Modulation string    `json:"modulation"`
		DataRate   string    `json:"data_rate"`
		BitRate    int       `json:"bit_rate"`
		CodingRate string    `json:"coding_rate"`
		Gateways   []struct {
			GtwID     string  `json:"gtw_id"`
			Timestamp int     `json:"timestamp"`
			Time      string  `json:"time,omitempty"`
			Channel   int     `json:"channel"`
			Rssi      int     `json:"rssi"`
			Snr       float64 `json:"snr"`
			RfChain   int     `json:"rf_chain"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Altitude  float64 `json:"altitude"`
		} `json:"gateways"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Altitude  float64 `json:"altitude"`
	} `json:"metadata"`
	DownlinkURL string `json:"downlink_url"`
}

func parseDeviceData(payload string) map[string]interface{} {
	rawPayload, err := base64.StdEncoding.DecodeString(payload)

	if err != nil {
		return map[string]interface{}{
			"raw":   payload,
			"error": err.Error(),
		}
	}

	decoder := cayennelpp.NewDecoder(bytes.NewBuffer(rawPayload))
	target := NewUplinkTarget()

	err = decoder.DecodeUplink(target)

	if err != nil {
		return map[string]interface{}{
			"raw":   payload,
			"error": err.Error(),
		}
	}

	return target.GetValues()
}

// DeviceData represents a row item.
type DeviceData struct {
	DeviceID  string
	Data      map[string]interface{}
	Timestamp time.Time
}

// Save implements the ValueSaver interface.
func (dd *DeviceData) Save() (map[string]bigquery.Value, string, error) {
	data, err := json.Marshal(dd.Data)
	if err != nil {
		return nil, "", err
	}
	return map[string]bigquery.Value{
		"deviceId": dd.DeviceID,
		"data":     string(data),
		"time":     dd.Timestamp,
	}, "", nil
}

// GetDeviceUpdate convert uplink message into a firebase update map
func GetDeviceUpdate(msg UplinkMessage) map[string]interface{} {

	gateways := map[string]interface{}{}
	for _, gateway := range msg.Metadata.Gateways {
		gateways[gateway.GtwID] = map[string]interface{}{
			"id":        gateway.GtwID,
			"rssi":      gateway.Rssi,
			"snr":       gateway.Snr,
			"channel":   gateway.Channel,
			"time":      gateway.Timestamp,
			"latitude":  gateway.Latitude,
			"longitude": gateway.Longitude,
			"altitude":  gateway.Altitude,
		}
	}

	base := map[string]interface{}{
		"deviceId": msg.DevID,
		"serial":   msg.HardwareSerial,
		"data":     parseDeviceData(msg.PayloadRaw),
		"meta": map[string]interface{}{
			"updated":   msg.Metadata.Time,
			"frequency": msg.Metadata.Frequency,
			"latitude":  msg.Metadata.Latitude,
			"longitude": msg.Metadata.Longitude,
			"altitude":  msg.Metadata.Altitude,
			"gateways":  gateways,
		},
	}

	return base
}
