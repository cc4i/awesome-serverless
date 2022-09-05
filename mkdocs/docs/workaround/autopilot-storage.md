#

## Description

In GKE Autopilot [the ephemeral storage limit](https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-resource-requests#min-max-requests) must be within 10 MiB and 10 GiB for all compute classes, it's a potential issue for those applicaitons need big size storage. 

There are three diffrent way to surpass the limit and you can choose one of those based on your scenario.

## Use memory for emptyDir
Ephmeral storage could be emptyDir, be default there's limit 10G, howerver emptyDir has different type, which could be host file system or memory directly, so we can use this kind of type to surpass 10G limit (memory limit is up to 80G at the moment). However it's kind of expensive and also CPU and Memory must be match ratio between 1:1 to 1:6.5.

For example, create 15Gi ephemeral storage that's going to occupied half of total memory 30Gi.

```sh

# Create Pod in Autopilot with emptyDir
kubectl apply -f - << __EOF__
apiVersion: v1
kind: Pod
metadata:
  name: buildah-emptydir 
spec:
  containers:
    - name: buildah
      image: quay.io/buildah/stable:v1.23.1
      command: ["sleep", "infinity"]  
      volumeMounts:
      - mountPath: /var/lib/containers
        name: container-storage
      resources:
        requests:
          cpu: 5000m
          memory: 30Gi
  volumes:
  - name: container-storage
    emptyDir:
     medium: Memory
     sizeLimit: 15Gi
__EOF__

# Validate the size of ephemeral storage
kubectl exec -it pods/buildah-emptydir -- bash
df -h

```


## Use Cloud Filestore

By default StorageClass for Filestore and Persistent Disk were been installed when you launched the GKE Autopilot. For example, create 1Ti storage with Filestore through StorageClass dynamically.

```sh
# Apply storage class
kubectl apply -f - << __EOF__
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: filestore-test
provisioner: filestore.csi.storage.gke.io
volumeBindingMode: Immediate
allowVolumeExpansion: true
parameters:
  tier: standard
  network: default
__EOF__


# Provision PVC through StorageClass dynamically 
kubectl apply -f - << __EOF__
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: podpvc
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: filestore-test
  resources:
    requests:
      storage: 1Ti
__EOF__

# Using the storage in following Nginx example
kubectl apply -f - << __EOF__
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-server-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        volumeMounts:
        - mountPath: /usr/share/nginx/html
          name: mypvc
      volumes:
      - name: mypvc
        persistentVolumeClaim:
          claimName: podpvc
__EOF__

```
## Use Persistent Disk

This example is going to provision 30Gi storage through PD storage class and mount by Nginx Pod.

```sh
# Check out default storage classes
kubectl get storageClass -o wide

# Provison 30G PD through storageClass
kubectl apply -f - << __EOF__
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pvc-demo
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 30Gi
  storageClassName: standard-rwo
__EOF__

# Using the storage in following Nginx example
kubectl apply -f - << __EOF__
kind: Pod
apiVersion: v1
metadata:
  name: pod-demo
spec:
  volumes:
    - name: pvc-demo-vol
      persistentVolumeClaim:
       claimName: pvc-demo
  containers:
    - name: pod-demo
      image: nginx
      resources:
        limits:
          cpu: 10m
          memory: 80Mi
        requests:
          cpu: 10m
          memory: 80Mi
      ports:
        - containerPort: 80
          name: "http-server"
      volumeMounts:
        - mountPath: "/usr/share/nginx/html"
          name: pvc-demo-vol
__EOF__
```


## Allowed Storage type 

- "configMap"
- "csi"
- "downwardAPI"
- "emptyDir"
- "gcePersistentDisk"
- "hostPath",
- "nfs"
- "persistentVolumeClaim"
- "projected"
- "secret"

## Refernces

- https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-overview