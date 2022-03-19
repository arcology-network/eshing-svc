package workers

import (
	"time"

	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

var (
	collectStart time.Time
	calcStart    time.Time
	CollectTime  = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Subsystem: "eshing",
		Name:      "collect_seconds",
		Help:      "The duration of collection step.",
	}, []string{})
	CollectTimeGauge = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Subsystem: "eshing",
		Name:      "collect_seconds_gauge",
		Help:      "The duration of collection step.",
	}, []string{})
	CalcTime = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Subsystem: "eshing",
		Name:      "calc_seconds",
		Help:      "The duration of calculation step.",
	}, []string{})
	CalcTimeGauge = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Subsystem: "eshing",
		Name:      "calc_seconds_gauge",
		Help:      "The duration of calculation step.",
	}, []string{})
)
