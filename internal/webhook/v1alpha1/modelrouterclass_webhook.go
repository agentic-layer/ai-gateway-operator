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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	aigatewayv1alpha1 "github.com/agentic-layer/ai-gateway-operator/api/v1alpha1"
)

const (
	DefaultClassAnnotation = "modelrouter.kubernetes.io/is-default-class"
)

// nolint:unused
// log is for logging in this package.
var modelRouterClassLog = logf.Log.WithName("modelrouterclass-resource")

// SetupModelRouterClassWebhookWithManager registers the webhook for ModelRouterClass in the manager.
func SetupModelRouterClassWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&aigatewayv1alpha1.ModelRouterClass{}).
		WithValidator(&ModelRouterClassCustomValidator{}).
		Complete()
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-agentic-layer-ai-v1alpha1-modelrouterclass,mutating=false,failurePolicy=fail,sideEffects=None,groups=agentic-layer.ai,resources=modelrouterclasses,verbs=create;update,versions=v1alpha1,name=vmodelrouterclass-v1alpha1.kb.io,admissionReviewVersions=v1

// ModelRouterClassCustomValidator struct is responsible for validating the ModelRouterClass resource
// when it is created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type ModelRouterClassCustomValidator struct {
	Client client.Client
}

var _ webhook.CustomValidator = &ModelRouterClassCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type ModelRouterClass.
func (v *ModelRouterClassCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	modelRouterClass, ok := obj.(*aigatewayv1alpha1.ModelRouterClass)
	if !ok {
		return nil, fmt.Errorf("expected a ModelRouterClass object but got %T", obj)
	}
	modelRouterClassLog.Info("Validation for ModelRouterClass upon creation", "name", modelRouterClass.GetName())

	return v.validateModelRouterClass(ctx, modelRouterClass)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type ModelRouterClass.
func (v *ModelRouterClassCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	modelRouterClass, ok := newObj.(*aigatewayv1alpha1.ModelRouterClass)
	if !ok {
		return nil, fmt.Errorf("expected a ModelRouterClass object for the newObj but got %T", newObj)
	}
	modelRouterClassLog.Info("Validation for ModelRouterClass upon update", "name", modelRouterClass.GetName())

	return v.validateModelRouterClass(ctx, modelRouterClass)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type ModelRouterClass.
func (v *ModelRouterClassCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No validation needed on delete
	return nil, nil
}

// validateModelRouterClass performs validation logic for ModelRouterClass resources.
func (v *ModelRouterClassCustomValidator) validateModelRouterClass(ctx context.Context, modelRouterClass *aigatewayv1alpha1.ModelRouterClass) (admission.Warnings, error) {
	var allErrs field.ErrorList

	// Check if this ModelRouterClass has the default class annotation set to "true"
	annotations := modelRouterClass.GetAnnotations()
	if annotations != nil && annotations[DefaultClassAnnotation] == "true" {
		// List all existing ModelRouterClasses resources
		var modelRouterClassList aigatewayv1alpha1.ModelRouterClassList
		if err := v.Client.List(ctx, &modelRouterClassList); err != nil {
			return nil, fmt.Errorf("failed to list ModelRouterClass resources: %w", err)
		}

		// Check if any other ModelRouterClass already has the default annotation
		for _, existingClass := range modelRouterClassList.Items {
			// Skip the current resource being validated
			if existingClass.GetName() == modelRouterClass.GetName() {
				continue
			}

			existingAnnotations := existingClass.GetAnnotations()
			if existingAnnotations != nil && existingAnnotations[DefaultClassAnnotation] == "true" {
				allErrs = append(allErrs, field.Invalid(
					field.NewPath("metadata", "annotations").Key(DefaultClassAnnotation),
					"true",
					fmt.Sprintf("another ModelRouterClass '%s' already has the default class annotation set to 'true'. Only one ModelRouterClass can be marked as default", existingClass.GetName()),
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
