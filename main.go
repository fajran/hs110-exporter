package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	dimensions  = []string{"ip", "alias", "deviceId", "mac", "model"}
	errCodeDesc = prometheus.NewDesc("hs110_err_code", "Error Code",
		dimensions, nil)
	currentMaDesc = prometheus.NewDesc("hs110_current_ma", "Current (mA)",
		dimensions, nil)
	powerMwDesc = prometheus.NewDesc("hs110_power_mw", "Power (mW)",
		dimensions, nil)
	totalWhDesc = prometheus.NewDesc("hs110_total_wh", "Total (Wh)",
		dimensions, nil)
	voltageMvDesc = prometheus.NewDesc("hs110_voltage_mv", "Voltage (mV)",
		dimensions, nil)
	onTimeDesc = prometheus.NewDesc("hs110_on_time", "On Time",
		dimensions, nil)
)

type Result struct {
	Emeter struct {
		GetRealtime struct {
			CurrentMa int64 `json:"current_ma"`
			ErrCode   int64 `json:"err_code"`
			PowerMw   int64 `json:"power_mw"`
			TotalWh   int64 `json:"total_wh"`
			VoltageMv int64 `json:"voltage_mv"`
		} `json:"get_realtime"`
	} `json:"emeter"`
	System struct {
		GetSysinfo struct {
			Alias    string `json:"alias"`
			DevName  string `json:"dev_name"`
			DeviceId string `json:"deviceId"`
			ErrCode  int64  `json:"err_code"`
			Mac      string `json:"mac"`
			Model    string `json:"model"`
			OnTime   int64  `json:"on_time"`
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type PlugCollector struct {
	ip   string
	plug Hs1xxPlug
}

func (c *PlugCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *PlugCollector) Collect(ch chan<- prometheus.Metric) {
	data, err := c.plug.MeterInfo()
	if err != nil {
		log.Printf("Unable to collect meter info: %v", err)
		return
	}

	var result Result
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		log.Printf("Unable to decode response: %v", err)
		return
	}

	dimensions := []string{
		c.ip,
		result.System.GetSysinfo.Alias,
		result.System.GetSysinfo.DeviceId,
		result.System.GetSysinfo.Mac,
		result.System.GetSysinfo.Model,
	}

	ch <- prometheus.MustNewConstMetric(
		errCodeDesc,
		prometheus.GaugeValue,
		float64(result.Emeter.GetRealtime.ErrCode),
		dimensions...,
	)
	ch <- prometheus.MustNewConstMetric(
		currentMaDesc,
		prometheus.GaugeValue,
		float64(result.Emeter.GetRealtime.CurrentMa),
		dimensions...,
	)
	ch <- prometheus.MustNewConstMetric(
		powerMwDesc,
		prometheus.GaugeValue,
		float64(result.Emeter.GetRealtime.PowerMw),
		dimensions...,
	)
	ch <- prometheus.MustNewConstMetric(
		totalWhDesc,
		prometheus.GaugeValue,
		float64(result.Emeter.GetRealtime.TotalWh),
		dimensions...,
	)
	ch <- prometheus.MustNewConstMetric(
		voltageMvDesc,
		prometheus.GaugeValue,
		float64(result.Emeter.GetRealtime.VoltageMv),
		dimensions...,
	)
	ch <- prometheus.MustNewConstMetric(
		onTimeDesc,
		prometheus.GaugeValue,
		float64(result.System.GetSysinfo.OnTime),
		dimensions...,
	)
}

func handle(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ip := q.Get("ip")
	if ip == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handleIp(ip, w, r)
}

func handleIp(ip string, w http.ResponseWriter, r *http.Request) {
	plug := Hs1xxPlug{IPAddress: ip}
	collector := &PlugCollector{ip, plug}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	http.HandleFunc("/metrics", handle)
	http.ListenAndServe(":9119", nil)
}
