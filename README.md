# tinkerbot

A Slackbot for Tinkerbell

tinkerbot uses the both the gRPC API of Tinkerbell, and the HTTP REST API of ELK for querying logs.

This bot uses the `golang-middleware` template and is written to be deployed to OpenFaaS.

## Commands:

### Logs

Fetch logs from ELK for a specific Tinkerbell component

```
/logs INDEX
```

Examples include:

```
/logs worker
/logs boots
/logs nginx
/logs registry
```

### Workflows

Get the events for the last updated workflow.

This command is ideal when you are trying out a workflow or debugging it to see how far it has got.

```
/last-workflow
```

## Installation

### Pre-reqs

* OpenFaaS installed with reachability to your ElasticSearch server, i.e. on the provisioner
* A public IP for Slack to send webhooks for the Slash commands, or use inlets for this

### Setup OpenFaaS

On the provisioner run k3d:

```bash
curl -s https://raw.githubusercontent.com/rancher/k3d/master/install.sh | bash
```

Create a cluster with k3d:

```bash
k3d create cluster --name tinkerbot --server-arg "--no-deploy=traefik" --server-arg "--no-deploy=servicelb"
export KUBECONFIG="$(k3d get-kubeconfig --name tinkerbot)"
```

```bash
curl -sLS https://dl.get-arkade.dev | sudo sh
```

Install OpenFaaS:

```bash
arkade install openfaas
```

Follow the instructions to install `faas-cli`, to port-forward the gateway, and to log in.

### Configure the bot

You will create two secrets:

* `basic-auth-password` - for your gateway admin user
* `slack-token` - Slack token for verification

Run the following:

```sh
export SLACK_TOKEN="test"
export PAYLOAD_SECRET="test"

faas-cli secret create slack-token --from-literal $SLACK_TOKEN
faas-cli secret create payload-secret --from-literal $PAYLOAD_SECRET
```

Now edit stack.yml and update the environment variable for where your ELK cluster is:

```
    environment:
      elk_host: http://192.168.0.61:9200/
```

Here, I am using an IP on my local network. By default Tinkerbell exposes ElasticSearch on all interfaces on your provisioner. Check the interfaces with `ip addr` or `ifconfig`.

### Deploy the bot

```bash
faas-cli deploy -f stack.yml
```

If you wish to hack on the bot and deploy a new version, replace `alexellis2` in the `image:` field with your own Docker Hub account, and then run `faas-cli up` instead of `faas-cli deploy`.

