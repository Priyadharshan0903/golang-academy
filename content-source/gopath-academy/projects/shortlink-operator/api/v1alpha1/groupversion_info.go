// Package v1alpha1 contains the ShortLink API: a Kubernetes custom
// resource that asks for a short link in the shortener service.
//
// +kubebuilder:object:generate=true
// +groupName=shortener.example.com
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion identifies this API: shortener.example.com/v1alpha1.
	GroupVersion = schema.GroupVersion{Group: "shortener.example.com", Version: "v1alpha1"}

	// SchemeBuilder registers the types with a runtime.Scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to a scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
