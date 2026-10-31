package collector

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

var (
	diskTempDesc      = desc("storage", "disk_temperature_celsius", "Disk temperature.", "disk", "model")
	diskSpinningDesc  = desc("storage", "disk_spinning", "Whether the disk is spinning.", "disk", "model")
	diskTotalDesc     = desc("storage", "disk_size_bytes", "Disk capacity.", "disk", "model")
	partitionSizeDesc = desc("storage", "partition_size_bytes", "Partition capacity.", "disk", "partition")
	partitionUsedDesc = desc("storage", "partition_used_bytes", "Used space on the partition.", "disk", "partition")
	partitionFreeDesc = desc("storage", "partition_free_bytes", "Free space on the partition.", "disk", "partition")
)

func collectStorage(ctx context.Context, c *freebox.Client, ch chan<- prometheus.Metric) error {
	disks, err := freebox.Get[[]freebox.Disk](ctx, c, "storage/disk/")
	if err != nil {
		return err
	}
	for _, d := range disks {
		id := strconv.Itoa(d.ID)
		ch <- prometheus.MustNewConstMetric(diskTempDesc, prometheus.GaugeValue, d.Temp, id, d.Model)
		ch <- prometheus.MustNewConstMetric(diskSpinningDesc, prometheus.GaugeValue, boolValue(d.Spinning), id, d.Model)
		ch <- prometheus.MustNewConstMetric(diskTotalDesc, prometheus.GaugeValue, float64(d.TotalBytes), id, d.Model)
		for _, p := range d.Partitions {
			ch <- prometheus.MustNewConstMetric(partitionSizeDesc, prometheus.GaugeValue, float64(p.TotalBytes), id, p.Label)
			ch <- prometheus.MustNewConstMetric(partitionUsedDesc, prometheus.GaugeValue, float64(p.UsedBytes), id, p.Label)
			ch <- prometheus.MustNewConstMetric(partitionFreeDesc, prometheus.GaugeValue, float64(p.FreeBytes), id, p.Label)
		}
	}
	return nil
}
