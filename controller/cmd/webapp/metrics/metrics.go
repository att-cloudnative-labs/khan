package metrics

import (
	"egbitbucket.dtvops.net/com/controller/internal/controller"

	"egbitbucket.dtvops.net/com/goatt"
)

// customMetric is the string used for the channel call metric. Labels are channel and device.
const customMetric = "custom_metric_example"

// Set is to register and set the different metrics
func Set(reporter goatt.Reporter) {
	reporter.RegisterCounter(customMetric, "This is the number of times the metric was called.", "label1")
	metricExample := func(label1 string) {
		reporter.CounterAdd(customMetric, 1, label1)
	}

	controller.MetricWelcomeCount = metricExample

	goatt.GenerateMetricsDocumentation("METRICS.MD")
}