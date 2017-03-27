package meraki

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// DATE format
const tsLayout string = "2006-01-02T15:04:05.000Z"

// Timestamp is the Timestamp formart
type Timestamp struct {
	Time time.Time
}

// Config represents configuration
type Config struct {
	MerakiAPI          string `json:"MerakiAPI,omitempty"`
	MerakiKey          string `json:"MerakiKey,omitempty"`
	NetworkID          string `json:"NetworkID,omitempty"`
	MerakiCMXValidator string `json:"MerakiCMXValidator,omitempty"`
	MerakiCMXSecret    string `json:"MerakiCMXSecret,omitempty"`
}

// ESSID represents a essid entity response
type ESSID struct {
	Number   int    `json:"number,omitempty"`
	Name     string `json:"name,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	AuthMode string `json:"enabled,omitempty"`
}

// AP ...
type AP struct {
	Name      string  `json:"name"`
	Lat       float32 `json:"lat"`
	Lng       float32 `json:"lng"`
	Serial    string  `json:"serial"`
	Mac       string  `json:"mac"`
	Model     string  `json:"model"`
	Address   string  `json:"address"`
	LanIP     string  `json:"lanIp"`
	Tags      string  `json:"tags"`
	NetworkID string  `json:"networkId"`
}

// Location ...
type Location struct {
	Lat float32   `json:"lat"`
	Lng float32   `json:"lng"`
	Unc float32   `json:"unc"`
	X   []float32 `json:"x"`
	Y   []float32 `json:"y"`
}

// Client ...
type Client struct {
	ClientMac    string    `json:"clientMac"`
	IPV4         string    `json:"ipv4"`
	IPV6         string    `json:"ipv6"`
	SeenTime     Timestamp `json:"seenTime"`
	SeenEpoch    int       `json:"seenEpoch"`
	SSID         string    `json:"ssid"`
	RSSI         int       `json:"rssi"`
	Manufacturer string    `json:"manufacturer"`
	OS           string    `json:"os"`
	Location     Location  `json:"location"`
}

// CMXData ...
type CMXData struct {
	ApMac        string   `json:"apMac"`
	ApTags       []string `json:"apTags"`
	ApFloors     []string `json:"apFloors"`
	Observations []Client `json:"observations"`
}

// CMXScanning ...
type CMXScanning struct {
	Version string  `json:"version"`
	Secret  string  `json:"secret"`
	Type    string  `json:"type"`
	Data    CMXData `json:"data"`
}

// MarshalJSON ...
func (ct *Timestamp) MarshalJSON() ([]byte, error) {
	// Default location is Mexico_City
	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		panic(err)
	}

	local := ct.Time.In(loc)
	return []byte(`"` + local.Format(tsLayout) + `"`), nil
}

// UnmarshalJSON ... parse date
func (ct *Timestamp) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	ct.Time, err = time.Parse(time.RFC3339, string(b))

	// If can't parse date, use default date
	if err != nil {
		ct.Time, err = time.Parse(time.RFC3339, string(tsLayout))
	}
	return
}

// GetESSIDs return a list of ESSIDS
func GetESSIDs(config Config, APIKey string, networkID string) []ESSID {
	var essids = []ESSID{}
	url := config.MerakiAPI + "/networks/" + networkID + "/ssids"
	println("Doing http request to: ", url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Cisco-Meraki-API-Key", APIKey)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Can't get essid:", err)
		return essids
	}

	url = resp.Request.URL.String()
	println("Doing http request to: ", url)
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Add("X-Cisco-Meraki-API-Key", APIKey)
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Can't make a http request")
		return essids
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("can't parse body")
	}
	if resp.StatusCode != 200 {
		fmt.Println("You have an error because:", body, APIKey)
		return essids
	}

	json.Unmarshal(body, &essids)

	return essids
}

// GetAPs return a list of APS
func GetAPs(config Config, APIKey string, networkID string) []AP {
	var aps = []AP{}
	var apsProcessed = []AP{}
	url := config.MerakiAPI + "/networks/" + networkID + "/devices"
	println("Doing http request to: ", url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Cisco-Meraki-API-Key", APIKey)
	resp, err := client.Do(req)

	url = resp.Request.URL.String()
	println("Doing http request to: ", url)
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Add("X-Cisco-Meraki-API-Key", APIKey)
	resp, err = client.Do(req)

	defer resp.Body.Close()
	if err != nil {
		fmt.Println("Can't make a http request")
		return aps
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print("can't parse body")
	}
	if resp.StatusCode != 200 {
		fmt.Println("You have an error because:", body, APIKey)
		return aps
	}

	json.Unmarshal(body, &aps)

	for _, ap := range aps {
		// MR is a prefix that meraki use to indentify ap
		if ap.Model[0:2] == "MR" {
			apsProcessed = append(apsProcessed, ap)
		}
	}

	return apsProcessed
}

