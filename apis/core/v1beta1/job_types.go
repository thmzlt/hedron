/*
Unlicensed
*/

package v1beta1

import (
	k8s_corev1 "k8s.io/api/core/v1"
	k8s_metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobSpec struct {
	ProjectRef k8s_corev1.LocalObjectReference `json:"projectRef,omitempty"`
	Revision   string                          `json:"revision,omitempty"`
}

type JobStatus struct {
	Output string `json:"output,omitempty"`
}

// +kubebuilder:object:root=true

type Job struct {
	k8s_metav1.TypeMeta   `json:",inline"`
	k8s_metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec,omitempty"`
	Status JobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type JobList struct {
	k8s_metav1.TypeMeta `json:",inline"`
	k8s_metav1.ListMeta `json:"metadata,omitempty"`
	Items               []Job `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Job{}, &JobList{})
}