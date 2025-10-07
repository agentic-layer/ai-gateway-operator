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

package v1alpha1

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	aigatewayv1alpha1 "github.com/agentic-layer/ai-gateway-operator/api/v1alpha1"
)

const (
	DefaultClassAnnotation = "aigateway.kubernetes.io/is-default-class"
)

// nolint:unused
// log is for logging in this package.
var aiGatewayClassLog = logf.Log.WithName("aigatewayclass-resource")

// SetupAiGatewayClassWebhookWithManager registers the webhook for AiGatewayClass in the manager.
func SetupAiGatewayClassWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&aigatewayv1alpha1.AiGatewayClass{}).
		WithValidator(&AiGatewayClassCustomValidator{Client: mgr.GetClient()}).
		Complete()
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-agentic-layer-ai-v1alpha1-aigatewayclass,mutating=false,failurePolicy=fail,sideEffects=None,groups=agentic-layer.ai,resources=aigatewayclasses,verbs=create;update,versions=v1alpha1,name=vaigatewayclass-v1alpha1.kb.io,admissionReviewVersions=v1

// AiGatewayClassCustomValidator struct is responsible for validating the AiGatewayClass resource
// when it is created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type AiGatewayClassCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &AiGatewayClassCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type AiGatewayClass.
func (v *AiGatewayClassCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	aiGatewayClass, ok := obj.(*aigatewayv1alpha1.AiGatewayClass)
	if !ok {
		return nil, fmt.Errorf("expected a AiGatewayClass object but got %T", obj)
	}
	aiGatewayClassLog.Info("Validation for AiGatewayClass upon creation", "name", aiGatewayClass.GetName())

	return v.validateAiGatewayClass(ctx, aiGatewayClass)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type AiGatewayClass.
func (v *AiGatewayClassCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	aiGatewayClass, ok := newObj.(*aigatewayv1alpha1.AiGatewayClass)
	if !ok {
		return nil, fmt.Errorf("expected a AiGatewayClass object for the newObj but got %T", newObj)
	}
	aiGatewayClassLog.Info("Validation for AiGatewayClass upon update", "name", aiGatewayClass.GetName())

	return v.validateAiGatewayClass(ctx, aiGatewayClass)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type AiGatewayClass.
func (v *AiGatewayClassCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed on delete
	return nil, nil
}

// validateAiGatewayClass performs validation logic for AiGatewayClass resources.
func (v *AiGatewayClassCustomValidator) validateAiGatewayClass(ctx context.Context, aiGatewayClass *aigatewayv1alpha1.AiGatewayClass) (admission.Warnings, error) {
	var allErrs field.ErrorList

	// Check if this AiGatewayClass has the default class annotation set to "true"
	annotations := aiGatewayClass.GetAnnotations()
	if annotations != nil && annotations[DefaultClassAnnotation] == "true" {
		// List all existing AiGatewayClasses resources
		var aiGatewayClassList aigatewayv1alpha1.AiGatewayClassList
		if err := v.Client.List(ctx, &aiGatewayClassList); err != nil {
			return nil, fmt.Errorf("failed to list AiGatewayClass resources: %w", err)
		}

		// Check if any other AiGatewayClass already has the default annotation
		for _, existingClass := range aiGatewayClassList.Items {
			// Skip the current resource being validated
			if existingClass.GetName() == aiGatewayClass.GetName() {
				continue
			}

			existingAnnotations := existingClass.GetAnnotations()
			if existingAnnotations != nil && existingAnnotations[DefaultClassAnnotation] == "true" {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath("metadata", "annotations").Key(DefaultClassAnnotation),
					"true",
					fmt.Sprintf("another AiGatewayClass '%s' already has the default class annotation set to 'true'. Only one AiGatewayClass can be marked as default", existingClass.GetName()),
				))
				break
			}
		}
	}

	if len(allErrs) > 0 {
		return nil, allErrs.ToAggregate()
	}

	return nil, nil
}
