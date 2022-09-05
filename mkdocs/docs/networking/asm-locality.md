#

## Description
Visualize pods across the nodes in different zones to supports highly available and scalable, as well as leverage weight distrubution of Istio to reduce inter-zone traffic & cost.

## Deplyment

```sh
# Clone repo
git clone https://github.com/cc4i/multi-k8s.git
cd multi-k8s
cd asset/tod/bin && ./gke.sh
# Execute printed commands to install Anthos Services Mesh

# Provision & deployment
skaffold build 
tag=`skaffold build --dry-run --output='{{json .}}' --quiet |jq '.builds[].tag' -r`
skaffold deploy --images ${tag}

# Run
endpoint=`kubectl get svc/istio-ingressgateway -n run-tracker -o "jsonpath={.status.loadBalancer.ingress[0].ip}"`
curl -v http://${endpoint}/trip |jq

```
## Usage
The UI to understand accessing map.
![image info](../images/tracker-ui.png)

```sh
endpoint=`kubectl get svc/istio-ingressgateway -n run-tracker -o "jsonpath={.status.loadBalancer.ingress[0].ip}"`

# Access by http://${endpoint}/tracker-ui
open http://${endpoint}/tracker-ui

```

## Notes
- Install script was built to run on Linux, so [run them in Cloud Shell or Linux](https://cloud.google.com/service-mesh/docs/unified-install/install-dependent-tools#install_required_tools) to install Anthos Services Mesh.