/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SimpleSpec defines the desired state
type SimpleSpec struct {
	// +kubebuilder:validation:MinLength=1
	// Message is the string to print
	Message string `json:"message"`
}

// SimpleStatus defines the observed state
type SimpleStatus struct {
	// +optional
	// Replied indicates that we’ve seen and logged the Message
	Replied bool `json:"replied,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Simple is the Schema for the simples API
type Simple struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of Simple
	// +required
	Spec SimpleSpec `json:"spec"`

	// status defines the observed state of Simple
	// +optional
	Status SimpleStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// SimpleList contains a list of Simple
type SimpleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Simple `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Simple{}, &SimpleList{})
}
