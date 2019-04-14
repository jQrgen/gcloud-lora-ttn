package ttn

import (
	"cloud.google.com/go/bigquery"
	"encoding/base64"
	"encoding/json"
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
			GtwID     string    `json:"gtw_id"`
			Timestamp int       `json:"timestamp"`
			Time      time.Time `json:"time"`
			Channel   int       `json:"channel"`
			Rssi      int       `json:"rssi"`
			Snr       float64   `json:"snr"`
			RfChain   int       `json:"rf_chain"`
			Latitude  float64   `json:"latitude"`
			Longitude float64   `json:"longitude"`
			Altitude  float64   `json:"altitude"`
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

	tempInt := (int(rawPayload[1]) << 8) | int(rawPayload[2])
	co2Int := (int(rawPayload[3]) << 8) | int(rawPayload[4])
	tvocInt := (int(rawPayload[5]) << 8) | int(rawPayload[6])
	batInt := (int(rawPayload[7]) << 8) | int(rawPayload[8])

	// Decode Feather ID
	featherID := rawPayload[0]
	// Decode int to float
	temp := float64(tempInt) / 100.0
	co2 := float64(co2Int) / 100.0
	tvoc := float64(tvocInt) / 100.0
	bat := float64(batInt) / 100.0

	return map[string]interface{}{
		"feather_id":  featherID,
		"temperature": temp,
		"co2":         co2,
		"tvoc":        tvoc,
		"battery":     bat,
	}
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
