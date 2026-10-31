package collector_test

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/xfoucron/freebox-exporter/internal/collector"
	"github.com/xfoucron/freebox-exporter/internal/freebox"
	"github.com/xfoucron/freebox-exporter/internal/freebox/freeboxtest"
)

var fixtures = map[string]string{
	"connection/": `{"state":"up","type":"rfc2684","media":"ftth","ipv4":"203.0.113.7","ipv6":"2001:db8::1",
		"rate_up":100,"rate_down":2000,"bandwidth_up":600000000,"bandwidth_down":900000000,"bytes_up":1000,"bytes_down":5000}`,
	"system/": `{"firmware_version":"4.8.0","board_name":"fbxgw7r","serial":"SN1","mac":"AA:BB:CC:DD:EE:FF","uptime_val":3600,
		"sensors":[{"id":"temp_cpu","name":"CPU","value":55}],"fans":[{"id":"fan0","name":"Ventilateur 1","value":2100}]}`,
	"lan/browser/interfaces/": `[{"name":"pub","host_count":2}]`,
	"lan/browser/pub/": `[
		{"primary_name":"laptop","active":true,"reachable":true,"l2ident":{"id":"11:22:33:44:55:66"},"access_point":{"rx_bytes":10,"tx_bytes":20}},
		{"names":[{"name":"tv"}],"active":false,"reachable":true,"l2ident":{"id":"66:55:44:33:22:11"}}]`,
	"storage/disk/": `[{"id":1,"model":"WD","temp":36,"spinning":true,"total_bytes":1000,
		"partitions":[{"label":"Freebox","total_bytes":900,"used_bytes":300,"free_bytes":600}]}]`,
	"wifi/ap/":            `[{"id":0,"name":"wifi0","status":{"state":"active"}}]`,
	"wifi/ap/0/stations/": `[{"mac":"11:22:33:44:55:66"},{"mac":"66:55:44:33:22:11"}]`,
}

func newCollector(t *testing.T, results map[string]string) *collector.Collector {
	t.Helper()
	srv := freeboxtest.NewServer(t, results)
	client, err := freebox.New(context.Background(), srv.URL, freeboxtest.AppID, freeboxtest.AppToken, nil)
	if err != nil {
		t.Fatalf("freebox.New: %v", err)
	}
	return collector.New(client, 5*time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestCollect(t *testing.T) {
	expected := `
# HELP freebox_connection_up Whether the internet connection is up.
# TYPE freebox_connection_up gauge
freebox_connection_up 1
# HELP freebox_connection_download_rate_bytes Current download rate in bytes per second.
# TYPE freebox_connection_download_rate_bytes gauge
freebox_connection_download_rate_bytes 2000
# HELP freebox_connection_uploaded_bytes_total Total bytes uploaded since the connection was established.
# TYPE freebox_connection_uploaded_bytes_total counter
freebox_connection_uploaded_bytes_total 1000
# HELP freebox_system_uptime_seconds Time since the Freebox booted.
# TYPE freebox_system_uptime_seconds gauge
freebox_system_uptime_seconds 3600
# HELP freebox_system_temperature_celsius Temperature reported by a sensor.
# TYPE freebox_system_temperature_celsius gauge
freebox_system_temperature_celsius{sensor="CPU"} 55
# HELP freebox_system_fan_rpm Fan speed in revolutions per minute.
# TYPE freebox_system_fan_rpm gauge
freebox_system_fan_rpm{fan="Ventilateur 1"} 2100
# HELP freebox_lan_hosts Number of hosts known on the interface.
# TYPE freebox_lan_hosts gauge
freebox_lan_hosts{interface="pub"} 2
# HELP freebox_lan_host_active Whether the host is currently active.
# TYPE freebox_lan_host_active gauge
freebox_lan_host_active{interface="pub",mac="11:22:33:44:55:66",name="laptop"} 1
freebox_lan_host_active{interface="pub",mac="66:55:44:33:22:11",name="tv"} 0
# HELP freebox_lan_host_received_bytes_total Bytes received by the Freebox from the host.
# TYPE freebox_lan_host_received_bytes_total counter
freebox_lan_host_received_bytes_total{interface="pub",mac="11:22:33:44:55:66",name="laptop"} 10
# HELP freebox_storage_partition_used_bytes Used space on the partition.
# TYPE freebox_storage_partition_used_bytes gauge
freebox_storage_partition_used_bytes{disk="1",partition="Freebox"} 300
# HELP freebox_wifi_stations Number of stations associated to the access point.
# TYPE freebox_wifi_stations gauge
freebox_wifi_stations{ap="0",name="wifi0"} 2
# HELP freebox_scrape_collector_success Whether the collector succeeded during the last scrape.
# TYPE freebox_scrape_collector_success gauge
freebox_scrape_collector_success{collector="connection"} 1
freebox_scrape_collector_success{collector="lan"} 1
freebox_scrape_collector_success{collector="storage"} 1
freebox_scrape_collector_success{collector="system"} 1
freebox_scrape_collector_success{collector="wifi"} 1
`
	err := testutil.CollectAndCompare(newCollector(t, fixtures), strings.NewReader(expected),
		"freebox_connection_up",
		"freebox_connection_download_rate_bytes",
		"freebox_connection_uploaded_bytes_total",
		"freebox_system_uptime_seconds",
		"freebox_system_temperature_celsius",
		"freebox_system_fan_rpm",
		"freebox_lan_hosts",
		"freebox_lan_host_active",
		"freebox_lan_host_received_bytes_total",
		"freebox_storage_partition_used_bytes",
		"freebox_wifi_stations",
		"freebox_scrape_collector_success",
	)
	if err != nil {
		t.Error(err)
	}
}

func TestCollectReportsFailingCollector(t *testing.T) {
	partial := map[string]string{"connection/": fixtures["connection/"]}
	expected := `
# HELP freebox_scrape_collector_success Whether the collector succeeded during the last scrape.
# TYPE freebox_scrape_collector_success gauge
freebox_scrape_collector_success{collector="connection"} 1
freebox_scrape_collector_success{collector="lan"} 0
freebox_scrape_collector_success{collector="storage"} 0
freebox_scrape_collector_success{collector="system"} 0
freebox_scrape_collector_success{collector="wifi"} 0
`
	err := testutil.CollectAndCompare(newCollector(t, partial), strings.NewReader(expected), "freebox_scrape_collector_success")
	if err != nil {
		t.Error(err)
	}
}
