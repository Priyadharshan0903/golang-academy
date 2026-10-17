package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go tool controller-gen object paths=. crd:crdVersions=v1 output:crd:dir=../../config/crd

// [start types]

// ShortLinkSpec is the desired state: what the user asks for.
type ShortLinkSpec struct {
	// URL is the target of the short link.
	// +kubebuilder:validation:Pattern=`^https?://`
	// +kubebuilder:validation:MaxLength=2048
	URL string `json:"url"`

	// Code is the short code. If empty, the ShortLink's name is used.
	// +kubebuilder:validation:Pattern=`^[A-Za-z0-9_-]{3,32}$`
	// +optional
	Code string `json:"code,omitempty"`
}

// ShortLinkStatus is the observed state: what the operator did.
type ShortLinkStatus struct {
	// Code is the code the operator created in the shortener.
	// +optional
	Code string `json:"code,omitempty"`

	// ShortURL is the public short URL.
	// +optional
	ShortURL string `json:"shortURL,omitempty"`

	// ObservedGeneration is the .metadata.generation the status describes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions describe the current state, most importantly "Ready".
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ShortLink asks the shortener service for a short link.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=sl
// +kubebuilder:printcolumn:name="Code",type=string,JSONPath=`.status.code`
// +kubebuilder:printcolumn:name="Short URL",type=string,JSONPath=`.status.shortURL`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type ShortLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ShortLinkSpec   `json:"spec"`
	Status ShortLinkStatus `json:"status,omitempty"`
}

// [end types]

// ShortLinkList is a list of ShortLinks.
//
// +kubebuilder:object:root=true
type ShortLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShortLink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShortLink{}, &ShortLinkList{})
}
