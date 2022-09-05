# 
## Description
Placing pods into nodes in single zone with high availability, there're two node pools, one is the primary node pool in signle zone, other one is standby node pool with minmum number of instance is zero.

## Deployment

```sh

# Clone repo
git clone https://github.com/cc4i/multi-k8s.git
cd multi-k8s
cd asset/tod/bin && ./gke-affinity.sh

# Apply manifests
cd ../manifests/examples/single-zone
kustomize build . |kubectl apply -f -

# 
endpoint=`kubectl get svc/svc-1 -n run-tracker -o "jsonpath={.status.loadBalancer.ingress[0].ip}"`
# Access by http://${endpoint}/tracker-ui
open http://${endpoint}:8000/tracker-ui

```

## Notes
- Affinity is a feature from upstream Kubernetes. 