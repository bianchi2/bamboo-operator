# Bamboo Kubernetes Operator

## Project State

The project is in its pre-alpha state. A lot of things need to be changed, mainly related to Bamboo configuration and agent pool auto management.

## Pre Reqs

* kubectl with configured context
* cluster admin privileges
* at least 2 free persistent volumes (RWO)
* ingress controller

## Kubernetes Support

The operator has been tested on K8s 1.18, but will likely work on earlier versions too, since the operator does not create any special K8s objects.
This operator will not work on OpenShift (no route object, and images may simply not start).

## How to Deploy

### Create a Custom Resource Definition

You will need cluster admin privileges extend K8s api:
```
kubectl apply -f config/crd/bases/install.atlassian.com_bambooes.yaml
```

### Create a namespace

```
kubectl create namespace atl
```

If you choose a different namespace name, open bamboo-operator.yaml and edit ClusterRoleBinding:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: bamboo-operator
subjects:
  - kind: ServiceAccount
    name: bamboo-operator
    namespace: ${YOUR_NAMESPACE}
```

### Customize a Custom Resource

This is the most important part of the installation, since the operator will parse CR spec to get all necessary info.

Open config/samples/install_v1alpha1_bamboo.yaml and carefully read comments to each and every field.

#### TLS

The operator will create an Ingress object. You may provide own annotations, as well as a secret name.

Here's an example of using Nginx ingress controller with cert-manager:

```
spec:
  ingress:
    host: bamboo.kubedemo.ml
    tls: true
    tlsSecretName: bamboo-tls
    annotations:
      kubernetes.io/ingress.class: "nginx"
      nginx.ingress.kubernetes.io/ssl-redirect: "true"
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
```

### Create a Custom Resource

```
kubectl apply -f config/samples/install_v1alpha1_bamboo.yaml -n atl
```

### Deploy Bamboo Operator

```
kubectl apply -f bamboo-operator -n atl
```

### Check Pods

```
kubectl get pods -n atl
```

When a pod named `install-k8s-bamboo-*` is in a completed state, it means the installation has been completed, however, Bamboo still needs a minute or so to finish the installation.
It can take another minute or so for remote agents to register themselves with the server.

### Check Operator Logs

```
kubectl logs 0f deployment/bamboo-operator -n atl
```

At first, you will see the following logs:

```
invalid character '<' looking for beginning of value
Failed to get build queue size. Error: invalid character '<' looking for beginning of value
2020-08-07T06:35:58.536Z	INFO	bamboo-operator	unable to use autoscaling based on build queue due to an error. Bamboo maybe starting up or is otherwise unavailable
2020-08-07T06:36:27.324Z	INFO	bamboo-operator	Remote agents enabled
invalid character '<' looking for beginning of value
```

It means Bamboo is still loading, and the operator cannot reach its API endpoint.
You will also see this error if for some reason Bamboo is unavailable.

## Operator Features

### Installation

The operator will install all K8s objects required for a working Bamboo instance: deployment, configmaps, service, ingress and a persistent volume claim.
The operator watches all objects, so, if you delete, say, Bamboo service or an ingress, it will be recreated.

### Automatic Backups

The Operator will automatically backup Postgres database (.sql file is saved to /var/lib/postgresql/data),
as well as create a tar of ${BAMBOO_HOME} at `${BAMBOO_HOME}/backups` before an upgrade. Requested images and tags will be compared with existing ones,
and if the operator notices any difference, a backup will be created. During the backup, Bamboo is shut down, and the operator will make an exec into Postgres container.
To create a backup of ${BAMBOO_HOME}, the operator will start a special pod that will be bound to bamboo PVC.

### Automatic Remote Agent Scaling

This is a VERY experimental feature, so anything can happen :) However, the current logic is:

* the operator will monitor Bamboo build queue from time to time (either every 30 seconds, or anytime a watched object has been changed)
* if build queue > maxBuildInQueue an operator will launch 1 additional remote agent and wait till it registers itself with the server
* there's a limit for max number of remote agents, so the operator will respect this value
* the operator will try to scale down agents if build queue == 0 and current number of idle agents > maxIdleAgents

Remote agents are K8s deployments and persistent volume claims. So, starting a new agent is creating a deployment + pvc.
Deleting an agent is deletion of an agent in Bamboo API + deleting respective deployment and pvc.

Remote agent names == deployment and pvc names. Such a mapping enables having predictable deployment and pvc names which can be deleted if agents are idle.
An agent pod has an init container with bamboo-agent.cfg that has name and UID pre-defined.

You may turn off auto-scaling feature by updating your custom resource:

```
spec:
    autoManagement:
      # enabled by default. 
      enabled: false
```
In this case, you may still and remove agent manually, by changing replicas in custom resource spec:

```
spec:
  remoteagents:
    # whether or not to deploy remote agents in K8s and manage them
    enabled: true
    # number of remote agents to create
    replicas: 10
    autoManagement:
      # enabled by default. 
      enabled: false
```
