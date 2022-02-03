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

package v1alpha1

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S3Spec defines the desired state of S3
type S3Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of S3. Edit s3_types.go to remove/update
	Bucket string `json:"bucket"`
	AccessList types.BucketCannedACL `json:"access_list,omitempty"`
	BucketConfiguration CreateBucketConfiguration `json:"bucket_configuration,omitempty"`
	GrantFullControl string `json:"grant_full_control,omitempty"`
	LockEnabled bool `json:"lock_enabled,omitempty"`
	Ownership types.ObjectOwnership `json:"ownership,omitempty"`
}

type CreateBucketConfiguration struct {
	LocationConstraint types.BucketLocationConstraint `json:"location_constraint,omitempty"`
}

// S3Status defines the observed state of S3
type S3Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Bucket string `json:"bucket,omitempty"`
	AccessList string `json:"access_list,omitempty"`
	BucketConfiguration string `json:"bucket_configuration,omitempty"`
	GrantFullControl string `json:"grant_full_control,omitempty"`
	LockEnabled bool `json:"lock_enabled,omitempty"`
	Ownership string `json:"ownership,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// S3 is the Schema for the s3s API
type S3 struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3Spec   `json:"spec,omitempty"`
	Status S3Status `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// S3List contains a list of S3
type S3List struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3 `json:"items"`
}

func init() {
	SchemeBuilder.Register(&S3{}, &S3List{})
}
