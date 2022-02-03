# AWS Hello World Operator
Hello World AWS Service Operataor.</br>
Created with operator-sdk. https://sdk.operatorframework.io.</br>
The operator supports creation and deletion of a S3 Bucket.

![Alt text](static/demo.gif?raw=true "Title")

## Prerequisites
1. kubectl
2. eksctl
3. Permissions to create/delete EKS cluster, create/delete IAM roles/policies

## Installation Guide
1. Clone the repositroy</br> `git clone https://github.com/ronerizon/aws-hello-world-operator.git`
2. Create EKS cluster on AWS. <b> The cluster will be created with oidc enabled, It will also create the required service account for the operator.</b></br> `eksctl create cluster -f manifests/eksctl.yaml`
3. Create the operator.<b> If your region is different from eu-west-1 change AWS_REGION env in manifest/operator.yaml</b></br> `kubectl apply -f manifests/operator.yaml`
4. Create S3 bucket and check console for creation.</br>`kubectl apply -f manifests/s3_bucket.yaml`
5. Delete S3 bucket and check console for deletion.</br>`kubectl delete -f manifests/s3_bucket.yaml`
6. Delete the EKS cluster if you finished using it. </br>`eksctl delete cluster -f manifests/eksctl.yaml`