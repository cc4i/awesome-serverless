# 
## Description

Securely acccess AWS services from GKE through Workload Identity without using long-living AWS credentials (AKSK), we leverage [gtoken](https://github.com/doitintl/gtoken) to automatically inject AWS credentials into pods.


## Prerequisite

- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- [GCLOUD CLI](https://cloud.google.com/sdk/docs/install)
- [Enable Workload Identity on GKE](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#enable)
- Configure application to [use Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#authenticating_to)

## Deployment

```sh
# 0. Configure environment variables for following steps
export PROJECT_ID=<Project ID>
export CLUSTER_NAME=<The name of GKE cluster>
export GSA_NAME=<The name of Service Account>
export GSA_ID=$(gcloud iam service-accounts describe --format json ${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com  | jq -r '.uniqueId')
export KSA_NAME=<The name of Service Account in Kubernetes>
export KSA_NAMESPACE=<The name of Namespace where the Service Account is located>
export AWS_ROLE_NAME=<The name of AWS role>
export AWS_POLICY_NAME=<The name of policy that was granted to AWS role>


# 1. Create the Service Account and grant necessary permissions
gcloud iam service-accounts create ${GSA_NAME} \
    --description="Service account for Workload Identity and test accessing AWS resource." \
    --display-name=${GSA_NAME}

gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role="roles/container.admin" \
    --role="roles/resourcemanager.projectIamAdmin"


# 2. Create the AWS role for demo purpose (Customize your role if need to)
# Add trusted principal for Google account
cat > gcp-trust-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "accounts.google.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "accounts.google.com:sub": "${GSA_ID}"
        }
      }
    }
  ]
}
EOF
# Create the AWS role 
aws iam create-role --role-name ${AWS_ROLE_NAME} --assume-role-policy-document file://gcp-trust-policy.json
# Attache the policy to the AWS role
aws iam attach-role-policy --role-name ${AWS_ROLE_NAME} --policy-arn arn:aws:iam::aws:policy/${AWS_POLICY_NAME}
# Retrieve ARN from the AWS role
export AWS_ROLE_ARN=$(aws iam get-role --role-name ${AWS_ROLE_NAME} --query Role.Arn --output text)


# 3. Create the namespace and related services account in GKE
# Create the namespace
kubectl create namespace ${KSA_NAMESPACE}
# Create the service account
kubectl create serviceaccount --namespace ${KSA_NAMESPACE} ${KSA_NAME}

# Binding Service Account with Workload Identity
gcloud iam service-accounts add-iam-policy-binding ${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${KSA_NAMESPACE}/${KSA_NAME}]"

kubectl annotate serviceaccount --namespace ${KSA_NAMESPACE} ${KSA_NAME} \
  iam.gke.io/gcp-service-account=${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com

kubectl annotate serviceaccount --namespace ${KSA_NAMESPACE} ${KSA_NAME} \
  amazonaws.com/role-arn=${AWS_ROLE_ARN}


# 4. Run demo to verify services accessing
cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: ${KSA_NAMESPACE}
spec:
  serviceAccountName: ${KSA_NAME}
  containers:
  - name: test-pod
    image: amazon/aws-cli
    command: ["tail", "-f", "/dev/null"]
EOF

# Login into Pod Shell
kubectl exec -it pods/test-pod -- sh
# In Pod shell: check AWS assumed role
aws sts get-caller-identity


```


## References
- [https://github.com/doitintl/gtoken](https://github.com/doitintl/gtoken)