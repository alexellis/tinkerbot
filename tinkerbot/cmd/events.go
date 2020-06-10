package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"sort"
	"time"

	"github.com/tinkerbell/tink/protos/workflow"
)

func GetEvents(text string) (string, error) {
	conn, err := getConnection()
	if err != nil {
		return "", err
	}
	workflowClient := workflow.NewWorkflowSvcClient(conn)

	id := text
	if len(id) == 0 {
		workflow, err := getLastWorkflow(workflowClient)
		if err != nil {
			return "", err
		}

		id = workflow.Id
	}

	req := workflow.GetRequest{Id: id}

	ctx := context.Background()

	events, err := workflowClient.ShowWorkflowEvents(ctx, &req)

	if err != nil {
		log.Fatal(err)
	}

	var wfEvents []*workflow.WorkflowActionStatus
	err = nil
	for event, err := events.Recv(); err == nil && event != nil; event, err = events.Recv() {
		wfEvents = append(wfEvents, event)
	}

	// {event.WorkerId, event.TaskName, event.ActionName, event.Seconds, event.Message, event.ActionStatus},
	if err != nil && err != io.EOF {
		return "", err
	}

	out := "Workflow events:\n"

	for _, event := range wfEvents {
		dur, _ := time.ParseDuration(fmt.Sprintf("%ds", event.Seconds))
		out = out + fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t\n", event.WorkerId, event.TaskName, event.ActionName, dur.String(), event.Message, event.ActionStatus)
	}

	return out, nil
}

func getLastWorkflow(workflowClient workflow.WorkflowSvcClient) (*workflow.Workflow, error) {
	ctx := context.Background()

	res, err := workflowClient.ListWorkflows(ctx, &workflow.Empty{})
	if err != nil {
		return nil, err
	}
	var wf *workflow.Workflow
	var wfs Workflows

	for wf, err = res.Recv(); err == nil && wf.Template != ""; wf, err = res.Recv() {
		wfs = append(wfs, wf)
	}

	if len(wfs) > 0 {
		sort.Sort(byUpdated{wfs})
	}

	return wfs[len(wfs)-1], nil
}
