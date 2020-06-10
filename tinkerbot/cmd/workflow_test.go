package cmd

import (
	"sort"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/tinkerbell/tink/protos/workflow"
)

func Test_Sorter(t *testing.T) {
	var wfs []*workflow.Workflow

	wfs = append(wfs, &workflow.Workflow{Id: "a", UpdatedAt: &timestamp.Timestamp{Seconds: 150}})
	wfs = append(wfs, &workflow.Workflow{Id: "b", UpdatedAt: &timestamp.Timestamp{Seconds: 200}})
	wfs = append(wfs, &workflow.Workflow{Id: "c", UpdatedAt: &timestamp.Timestamp{Seconds: 100}})

	sort.Sort(byUpdated{wfs})
	wantLast := "b"
	if wfs[len(wfs)-1].Id != "b" {
		t.Fatalf(`last should be "%s" but was not`, wantLast)
	}
}
