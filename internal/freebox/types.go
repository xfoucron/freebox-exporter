package freebox

type Connection struct {
	State         string `json:"state"`
	Type          string `json:"type"`
	Media         string `json:"media"`
	IPv4          string `json:"ipv4"`
	IPv6          string `json:"ipv6"`
	RateUp        uint64 `json:"rate_up"`
	RateDown      uint64 `json:"rate_down"`
	BandwidthUp   uint64 `json:"bandwidth_up"`
	BandwidthDown uint64 `json:"bandwidth_down"`
	BytesUp       uint64 `json:"bytes_up"`
	BytesDown     uint64 `json:"bytes_down"`
}

type System struct {
	FirmwareVersion string   `json:"firmware_version"`
	BoardName       string   `json:"board_name"`
	Serial          string   `json:"serial"`
	MAC             string   `json:"mac"`
	Uptime          uint64   `json:"uptime_val"`
	Sensors         []Sensor `json:"sensors"`
	Fans            []Fan    `json:"fans"`
}

type Sensor struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Fan struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type LANInterface struct {
	Name      string `json:"name"`
	HostCount int    `json:"host_count"`
}

type Host struct {
	PrimaryName string `json:"primary_name"`
	DefaultName string `json:"default_name"`
	Names       []struct {
		Name string `json:"name"`
	} `json:"names"`
	Active    bool `json:"active"`
	Reachable bool `json:"reachable"`
	L2Ident   struct {
		ID string `json:"id"`
	} `json:"l2ident"`
	AccessPoint *struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"access_point"`
}

func (h Host) Name() string {
	switch {
	case h.PrimaryName != "":
		return h.PrimaryName
	case h.DefaultName != "":
		return h.DefaultName
	case len(h.Names) > 0:
		return h.Names[0].Name
	}
	return ""
}

type Disk struct {
	ID         int         `json:"id"`
	Model      string      `json:"model"`
	Temp       float64     `json:"temp"`
	Spinning   bool        `json:"spinning"`
	TotalBytes uint64      `json:"total_bytes"`
	Partitions []Partition `json:"partitions"`
}

type Partition struct {
	Label      string `json:"label"`
	TotalBytes uint64 `json:"total_bytes"`
	UsedBytes  uint64 `json:"used_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
}

type AccessPoint struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status struct {
		State string `json:"state"`
	} `json:"status"`
}

type Station struct {
	MAC string `json:"mac"`
}
