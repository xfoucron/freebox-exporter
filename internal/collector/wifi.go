package collector

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

var (
	wifiAPUpDesc     = desc("wifi", "access_point_up", "Whether the access point is active.", "ap", "name")
	wifiStationsDesc = desc("wifi", "stations", "Number of stations associated to the access point.", "ap", "name")
)

func collectWiFi(ctx context.Context, c *freebox.Client, ch chan<- prometheus.Metric) error {
	aps, err := freebox.Get[[]freebox.AccessPoint](ctx, c, "wifi/ap/")
	if err != nil {
		return err
	}
	for _, ap := range aps {
		id := strconv.Itoa(ap.ID)
		ch <- prometheus.MustNewConstMetric(wifiAPUpDesc, prometheus.GaugeValue, boolValue(ap.Status.State == "active"), id, ap.Name)

		stations, err := freebox.Get[[]freebox.Station](ctx, c, "wifi/ap/"+id+"/stations/")
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(wifiStationsDesc, prometheus.GaugeValue, float64(len(stations)), id, ap.Name)
	}
	return nil
}
