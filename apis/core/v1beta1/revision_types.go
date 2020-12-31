/*
Unlicensed
*/

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=Pending;Building;Ready;Failed

type State string

type RevisionSpec struct {
	ProjectRef corev1.LocalObjectReference `json:"projectRef,omitempty"`
	Revision   string                      `json:"revision,omitempty"`
}

type RevisionStatus struct {
	State State `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`

type Revision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RevisionSpec   `json:"spec,omitempty"`
	Status RevisionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type RevisionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Revision `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Revision{}, &RevisionList{})
}
