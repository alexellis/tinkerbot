package cmd

import (
	"context"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

func (s byUpdated) Less(i, j int) bool {
	u1 := time.Unix(s.Workflows[i].GetUpdatedAt().Seconds, 0)
	u2 := time.Unix(s.Workflows[j].GetUpdatedAt().Seconds, 0)

	return u2.After(u1)
}

// GetConnection returns a gRPC client connection
func getConnection() (*grpc.ClientConn, error) {
	certURL := os.Getenv("TINKERBELL_CERT_URL")
	if certURL == "" {
		return nil, errors.New("undefined TINKERBELL_CERT_URL")
	}

	resp, err := http.Get(certURL)
	if err != nil {
		return nil, errors.Wrap(err, "fetch cert")
	}
	defer resp.Body.Close()

	certs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read cert")
	}

	cp := x509.NewCertPool()
	ok := cp.AppendCertsFromPEM(certs)
	if !ok {
		return nil, errors.Wrap(err, "parse cert")
	}

	grpcAuthority := os.Getenv("TINKERBELL_GRPC_AUTHORITY")
	if grpcAuthority == "" {
		return nil, errors.New("undefined TINKERBELL_GRPC_AUTHORITY")
	}

	creds := credentials.NewClientTLSFromCert(cp, "")
	conn, err := grpc.Dial(grpcAuthority, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrap(err, "connect to tinkerbell server")
	}

	return conn, nil
}
