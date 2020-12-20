/*
Unlicensed
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Image struct {
	Name       string   `json:"name,omitempty"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
}

type Repository struct {
	URL string `json:"url,omitempty"`
	Ref string `json:"ref,omitempty"`
}

type ProjectSpec struct {
	Image      Image      `json:"image,omitempty"`
	Repository Repository `json:"repository,omitempty"`
}

type ProjectStatus struct {
	LastRevision string `json:"lastRevision,omitempty"`
}

// +kubebuilder:object:root=true

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectSpec   `json:"spec,omitempty"`
	Status ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Project{}, &ProjectList{})
}
