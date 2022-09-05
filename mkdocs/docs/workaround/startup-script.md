#

## Description
In GKE Standard you can't run customized startup script when booting up a node due to not support customized host operation system (OS), fortunately we can use DaemonSet to address this issue. The following example tried to demostrate how to run a startup script through DeamonSet, you can build you own as the reference here.

## Deployment

```sh

# Clone repo
git clone https://github.com/cc4i/multi-k8s.git
cd asset/startup-script

# build image & push into registry
docker build . -t <image url>:<tag>
docker push <image url>:<tag>

# Modify script as per request betweend "# BO:" and "# EO:"

# Deploy to cluster
kustomize build ./manifests | kubectl apply -f -

```

## Notes
- The pod container can share the host process ID namespace.
- THe container is required root privilege.
