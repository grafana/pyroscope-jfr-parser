package pyroscope

import (
	"testing"

	v1 "github.com/grafana/pyroscope/api/gen/proto/go/types/v1"
	"github.com/stretchr/testify/assert"
)

func labelMap(ls []*v1.LabelPair) map[string][]string {
	m := make(map[string][]string)
	for _, l := range ls {
		m[l.Name] = append(m[l.Name], l.Value)
	}
	return m
}

// Fixed labels are always present regardless of inputs.
func TestLabels_FixedLabelsAlwaysPresent(t *testing.T) {
	m := labelMap(Labels(nil, "cpu", "process_cpu", "myapp", "jfr"))

	assert.Equal(t, []string{"jfr"}, m[LabelNamePyroscopeSpy])
	assert.Equal(t, []string{"false"}, m[LabelNameDelta])
	assert.Equal(t, []string{"cpu"}, m[LabelNameJfrEvent])
	assert.Equal(t, []string{"process_cpu"}, m[LabelNameProfileName])
}

// No service_name in seriesLabels: appName is added as service_name.
func TestLabels_AppNameAddedAsServiceName(t *testing.T) {
	m := labelMap(Labels(map[string]string{"env": "prod"}, "cpu", "process_cpu", "myapp", "jfr"))

	assert.Equal(t, []string{"myapp"}, m[LabelNameServiceName])
	assert.Nil(t, m["app_name"])
}

// service_name already in seriesLabels: appName is added as app_name.
func TestLabels_AppNameFallsBackToAppName(t *testing.T) {
	m := labelMap(Labels(map[string]string{LabelNameServiceName: "existing"}, "cpu", "process_cpu", "myapp", "jfr"))

	assert.Equal(t, []string{"existing"}, m[LabelNameServiceName])
	assert.Equal(t, []string{"myapp"}, m["app_name"])
}

// Both service_name and app_name in seriesLabels: app_name must not be duplicated.
func TestLabels_NoDuplicateAppName(t *testing.T) {
	m := labelMap(Labels(
		map[string]string{LabelNameServiceName: "svc", "app_name": "original"},
		"cpu", "process_cpu", "myapp", "jfr",
	))

	assert.Equal(t, []string{"original"}, m["app_name"])
}

// appName is empty: neither service_name nor app_name is appended.
func TestLabels_EmptyAppNameNotAppended(t *testing.T) {
	m := labelMap(Labels(map[string]string{"env": "prod"}, "cpu", "process_cpu", "", "jfr"))

	assert.Nil(t, m[LabelNameServiceName])
	assert.Nil(t, m["app_name"])
}

// Private labels (starting with __) are filtered out, except __session_id__.
func TestLabels_PrivateLabelsFiltered(t *testing.T) {
	m := labelMap(Labels(
		map[string]string{"__private__": "secret", LabelNameSessionID: "sess123", "env": "prod"},
		"cpu", "process_cpu", "myapp", "jfr",
	))

	assert.Nil(t, m["__private__"])
	assert.Equal(t, []string{"sess123"}, m[LabelNameSessionID])
	assert.Equal(t, []string{"prod"}, m["env"])
}
