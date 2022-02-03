/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	awsv1alpha1 "platform.operatorhello.com/v1alpha1/api/v1alpha1"
)

// S3Reconciler reconciles a S3 object
type S3Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type S3BucketReconcilerState struct {
	Exists      bool
	AccessList  *s3.GetBucketAclOutput
	LockEnabled *s3.GetObjectLockConfigurationOutput
}

type NotImplementedError struct {
	message string
}

//+kubebuilder:rbac:groups=aws.platform.operatorhello.com,resources=s3s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aws.platform.operatorhello.com,resources=s3s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aws.platform.operatorhello.com,resources=s3s/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the S3 object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *S3Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	s3ServiceInstance := &awsv1alpha1.S3{}
	reconcileState := S3BucketReconcilerState{Exists: false}
	s3FinalizerName := "aws.platform.operatorhello.com/s3finalizer"
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		reqLogger.Error(err, "unable to load SDK config")
	}

	stsInstance := sts.NewFromConfig(cfg)
	s3Instance := s3.NewFromConfig(cfg)
	stsInstanceInput := &sts.GetCallerIdentityInput{}
	err = r.Get(ctx, req.NamespacedName, s3ServiceInstance)

	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("S3 resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "Failed to get S3")
		return reconcile.Result{}, err
	}

	stsInstanceOutput, err := stsInstance.GetCallerIdentity(context.TODO(), stsInstanceInput)
	if err != nil {
		reqLogger.Error(err, "failed to get identity")
		return ctrl.Result{}, err
	}
	awsAccount := *stsInstanceOutput.Account
	if s3ServiceInstance.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(s3ServiceInstance, s3FinalizerName) {
			controllerutil.AddFinalizer(s3ServiceInstance, s3FinalizerName)
			if err := r.Update(ctx, s3ServiceInstance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(s3ServiceInstance, s3FinalizerName) {
			if err := r.deleteS3Bucket(s3Instance, &awsAccount, &s3ServiceInstance.Spec.Bucket); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(s3ServiceInstance, s3FinalizerName)
			if err := r.Update(ctx, s3ServiceInstance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	bucketList, err := s3Instance.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	for _, bucket := range bucketList.Buckets {
		if *bucket.Name == s3ServiceInstance.Spec.Bucket {
			reconcileState.Exists = true
			r.updateS3Bucket(&reconcileState, s3ServiceInstance, s3Instance, &awsAccount, &s3ServiceInstance.Spec.Bucket)
			break
		}
	}

	if reconcileState.Exists {
		r.createS3Bucket(s3Instance, s3ServiceInstance)
		return ctrl.Result{}, nil
	}

	s3BucketInput := s3.CreateBucketInput{
		Bucket:                     &s3ServiceInstance.Spec.Bucket,
		ACL:                        s3ServiceInstance.Spec.AccessList,
		CreateBucketConfiguration:  &types.CreateBucketConfiguration{LocationConstraint: s3ServiceInstance.Spec.BucketConfiguration.LocationConstraint},
		ObjectLockEnabledForBucket: s3ServiceInstance.Spec.LockEnabled,
		ObjectOwnership:            s3ServiceInstance.Spec.Ownership,
		GrantFullControl:           &s3ServiceInstance.Spec.GrantFullControl,
	}

	_, err = s3Instance.CreateBucket(context.Background(), &s3BucketInput)
	if err != nil {
		reqLogger.Error(err, "failed to create bucket")
	}
	reqLogger.Info("Bucket successfully created", "Bucket", s3ServiceInstance.Spec.Bucket)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&awsv1alpha1.S3{}).
		Complete(r)
}

func (r *S3Reconciler) deleteS3Bucket(S3Instance *s3.Client, Account *string, BucketName *string) error {
	_, err := S3Instance.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket:              BucketName,
		ExpectedBucketOwner: Account,
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *S3Reconciler) updateS3Bucket(reconcileState *S3BucketReconcilerState, s3ServiceInstance *awsv1alpha1.S3, s3Instance *s3.Client, Account *string, BucketName *string) error {
	return &NotImplementedError{message: "updateS3Bucket"}
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("%s not implemented error", e.message)
}

func (r *S3Reconciler) createS3Bucket(s3Instance *s3.Client, s3ServiceInstance *awsv1alpha1.S3) error {
	_, err := s3Instance.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket:                     &s3ServiceInstance.Spec.Bucket,
		ACL:                        s3ServiceInstance.Spec.AccessList,
		CreateBucketConfiguration:  &types.CreateBucketConfiguration{LocationConstraint: s3ServiceInstance.Spec.BucketConfiguration.LocationConstraint},
		GrantFullControl:           &s3ServiceInstance.Spec.GrantFullControl,
		ObjectLockEnabledForBucket: s3ServiceInstance.Spec.LockEnabled,
		ObjectOwnership:            s3ServiceInstance.Spec.Ownership,
	})
	if err != nil {
		return err
	}
	return nil
}
