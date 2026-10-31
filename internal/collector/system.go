package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

var (
	systemInfoDesc   = desc("system", "info", "Freebox information.", "firmware", "board", "serial", "mac")
	systemUptimeDesc = desc("system", "uptime_seconds", "Time since the Freebox booted.")
	systemTempDesc   = desc("system", "temperature_celsius", "Temperature reported by a sensor.", "sensor")
	systemFanDesc    = desc("system", "fan_rpm", "Fan speed in revolutions per minute.", "fan")
)

func collectSystem(ctx context.Context, c *freebox.Client, ch chan<- prometheus.Metric) error {
	sys, err := freebox.Get[freebox.System](ctx, c, "system/")
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(systemInfoDesc, prometheus.GaugeValue, 1, sys.FirmwareVersion, sys.BoardName, sys.Serial, sys.MAC)
	ch <- prometheus.MustNewConstMetric(systemUptimeDesc, prometheus.GaugeValue, float64(sys.Uptime))
	for _, s := range sys.Sensors {
		ch <- prometheus.MustNewConstMetric(systemTempDesc, prometheus.GaugeValue, s.Value, s.Name)
	}
	for _, f := range sys.Fans {
		ch <- prometheus.MustNewConstMetric(systemFanDesc, prometheus.GaugeValue, f.Value, f.Name)
	}
	return nil
}
