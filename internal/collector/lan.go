package collector

import (
	"context"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/xfoucron/freebox-exporter/internal/freebox"
)

var (
	lanHostsDesc         = desc("lan", "hosts", "Number of hosts known on the interface.", "interface")
	lanHostActiveDesc    = desc("lan", "host_active", "Whether the host is currently active.", "interface", "name", "mac")
	lanHostReachableDesc = desc("lan", "host_reachable", "Whether the host is currently reachable.", "interface", "name", "mac")
	lanHostRxDesc        = desc("lan", "host_received_bytes_total", "Bytes received by the Freebox from the host.", "interface", "name", "mac")
	lanHostTxDesc        = desc("lan", "host_transmitted_bytes_total", "Bytes sent by the Freebox to the host.", "interface", "name", "mac")
)

func collectLAN(ctx context.Context, c *freebox.Client, ch chan<- prometheus.Metric) error {
	interfaces, err := freebox.Get[[]freebox.LANInterface](ctx, c, "lan/browser/interfaces/")
	if err != nil {
		return err
	}
	for _, iface := range interfaces {
		ch <- prometheus.MustNewConstMetric(lanHostsDesc, prometheus.GaugeValue, float64(iface.HostCount), iface.Name)

		hosts, err := freebox.Get[[]freebox.Host](ctx, c, "lan/browser/"+url.PathEscape(iface.Name)+"/")
		if err != nil {
			return err
		}
		for _, h := range hosts {
			ch <- prometheus.MustNewConstMetric(lanHostActiveDesc, prometheus.GaugeValue, boolValue(h.Active), iface.Name, h.Name(), h.L2Ident.ID)
			ch <- prometheus.MustNewConstMetric(lanHostReachableDesc, prometheus.GaugeValue, boolValue(h.Reachable), iface.Name, h.Name(), h.L2Ident.ID)
			if h.AccessPoint != nil {
				ch <- prometheus.MustNewConstMetric(lanHostRxDesc, prometheus.CounterValue, float64(h.AccessPoint.RxBytes), iface.Name, h.Name(), h.L2Ident.ID)
				ch <- prometheus.MustNewConstMetric(lanHostTxDesc, prometheus.CounterValue, float64(h.AccessPoint.TxBytes), iface.Name, h.Name(), h.L2Ident.ID)
			}
		}
	}
	return nil
}
