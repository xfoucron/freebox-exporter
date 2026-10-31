package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

var (
	connectionUpDesc        = desc("connection", "up", "Whether the internet connection is up.")
	connectionInfoDesc      = desc("connection", "info", "Connection information.", "type", "media", "ipv4", "ipv6")
	connectionRateUpDesc    = desc("connection", "upload_rate_bytes", "Current upload rate in bytes per second.")
	connectionRateDownDesc  = desc("connection", "download_rate_bytes", "Current download rate in bytes per second.")
	connectionBandUpDesc    = desc("connection", "upload_bandwidth_bits", "Available upload bandwidth in bits per second.")
	connectionBandDownDesc  = desc("connection", "download_bandwidth_bits", "Available download bandwidth in bits per second.")
	connectionBytesUpDesc   = desc("connection", "uploaded_bytes_total", "Total bytes uploaded since the connection was established.")
	connectionBytesDownDesc = desc("connection", "downloaded_bytes_total", "Total bytes downloaded since the connection was established.")
)

func collectConnection(ctx context.Context, c *freebox.Client, ch chan<- prometheus.Metric) error {
	conn, err := freebox.Get[freebox.Connection](ctx, c, "connection/")
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(connectionUpDesc, prometheus.GaugeValue, boolValue(conn.State == "up"))
	ch <- prometheus.MustNewConstMetric(connectionInfoDesc, prometheus.GaugeValue, 1, conn.Type, conn.Media, conn.IPv4, conn.IPv6)
	ch <- prometheus.MustNewConstMetric(connectionRateUpDesc, prometheus.GaugeValue, float64(conn.RateUp))
	ch <- prometheus.MustNewConstMetric(connectionRateDownDesc, prometheus.GaugeValue, float64(conn.RateDown))
	ch <- prometheus.MustNewConstMetric(connectionBandUpDesc, prometheus.GaugeValue, float64(conn.BandwidthUp))
	ch <- prometheus.MustNewConstMetric(connectionBandDownDesc, prometheus.GaugeValue, float64(conn.BandwidthDown))
	ch <- prometheus.MustNewConstMetric(connectionBytesUpDesc, prometheus.CounterValue, float64(conn.BytesUp))
	ch <- prometheus.MustNewConstMetric(connectionBytesDownDesc, prometheus.CounterValue, float64(conn.BytesDown))
	return nil
}
