package cmd

import (
	"context"
	"io"
	"sort"
	"time"

	"github.com/tinkerbell/tink/protos/workflow"
)

// GetWorkflow looks up a list of workflows from Tinkerbell's
// gRPC API
func GetWorkflow(text string) (string, error) {

	conn, err := getConnection()
	if err != nil {
		return "", err
	}

	workflowClient := workflow.NewWorkflowSvcClient(conn)

	ctx := context.Background()

	res, err := workflowClient.ListWorkflows(ctx, &workflow.Empty{})
	if err != nil {
		return "", err
	}

	var wf *workflow.Workflow
	var wfs Workflows

	out := "Workflows:\n"
	for wf, err = res.Recv(); err == nil && wf.Template != ""; wf, err = res.Recv() {
		wfs = append(wfs, wf)
	}

	if len(wfs) > 0 {
		sort.Sort(byUpdated{wfs})
	}

	for _, w := range wfs {
		updated := time.Unix(w.GetUpdatedAt().Seconds, 0)
		out = out + "\t" + updated.String() + "\t" + w.Id + "\n"
	}

	if err != nil && err != io.EOF {
		return "", err
	}

	return out, nil
}

type Workflows []*workflow.Workflow
type byUpdated struct{ Workflows }

func (s Workflows) Len() int      { return len(s) }
func (s Workflows) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less sorts with the latest updated coming last in the list
func (s byUpdated) Less(i, j int) bool {
	u1 := time.Unix(s.Workflows[i].GetUpdatedAt().Seconds, 0)
	u2 := time.Unix(s.Workflows[j].GetUpdatedAt().Seconds, 0)

	return u2.After(u1)
}
