package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PromObj interface {
	Set(key string, labels []string, value float64)
}

type GaugeVecObj struct {
	obj *prometheus.GaugeVec
}

func (p GaugeVecObj) Set(key string, labels []string, value float64) {
	p.obj.WithLabelValues(labels...).Set(value)
}

type HistogramVecObj struct {
	obj *prometheus.HistogramVec
}

func (p HistogramVecObj) Set(key string, labels []string, value float64) {
	p.obj.WithLabelValues(labels...).Observe(value)
}

type HistogramObj struct {
	obj prometheus.Histogram
}

func (p HistogramObj) Set(key string, labels []string, value float64) {
	p.obj.Observe(value)
}

type CounterVecObj struct {
	obj *prometheus.CounterVec
}

func (p CounterVecObj) Set(key string, labels []string, value float64) {
	var cV = m[key]
	if value != cV {
		p.obj.WithLabelValues(labels...).Add(value - cV)
		m[key] = value
	}
}

type GaugeObj struct {
	obj prometheus.Gauge
}

func (p GaugeObj) Set(key string, labels []string, value float64) {
	p.obj.Set(value)
}

type CounterObj struct {
	obj prometheus.Counter
}

func (p CounterObj) Set(key string, labels []string, value float64) {
	var cV = m[key]
	if value != cV {
		p.obj.Add(value - cV)
		m[key] = value
	}
}
