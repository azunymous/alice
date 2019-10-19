package dependencies

import (
	"strconv"
	"strings"
)

type Dependencies struct {
	// 0 = unhealthy
	// 1 = healthy
	// 2 = fallback
	all                    map[string]int
	markFallbacksUnhealthy bool
}

const ImageGroup = "images"

func Setup() Dependencies {
	dependencies := Dependencies{make(map[string]int), false}

	dependencies.all["minio"] = 0
	dependencies.all["redis"] = 0

	return dependencies
}

func (d *Dependencies) Healthy() bool {
	for _, health := range d.all {
		if d.markFallbacksUnhealthy && health != 1 {
			return false
		} else if health < 1 {
			return false
		}
	}
	return true
}

func (d *Dependencies) String() string {
	healthRecord := "{"
	for name, health := range d.all {
		healthRecord += `"` + name + `" : ` + strconv.Itoa(health) + ","
	}
	return strings.TrimSuffix(healthRecord, ",") + "}"
}

func (d *Dependencies) ImageGroup() string {
	return ImageGroup
}

func (d *Dependencies) SetUnhealthy(name string) {
	d.all[name] = 0
}

func (d *Dependencies) SetHealthy(name string) {
	d.all[name] = 1
}

func (d *Dependencies) SetFallback(name string) {
	d.all[name] = 2
}

func (d *Dependencies) MarkFallbacksUnhealthy(markUnhealthy bool) {
	d.markFallbacksUnhealthy = markUnhealthy
}
