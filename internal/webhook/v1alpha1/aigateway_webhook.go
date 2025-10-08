/*
Copyright 2025 Agentic Layer.

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
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	gatewayv1alpha1 "github.com/agentic-layer/ai-gateway-operator/api/v1alpha1"
)

// nolint:unused
// log is for logging in this package.
var aigatewaylog = logf.Log.WithName("aigateway-resource")

// SetupAiGatewayWebhookWithManager registers the webhook for AiGateway in the manager.
func SetupAiGatewayWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&gatewayv1alpha1.AiGateway{}).
		WithValidator(&AiGatewayCustomValidator{}).
		WithDefaulter(&AiGatewayCustomDefaulter{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-gateway-agentic-layer-ai-v1alpha1-aigateway,mutating=true,failurePolicy=fail,sideEffects=None,groups=gateway.agentic-layer.ai,resources=aigateways,verbs=create;update,versions=v1alpha1,name=aigateway-v1alpha1.kb.io,admissionReviewVersions=v1

// AiGatewayCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind AiGateway when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type AiGatewayCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &AiGatewayCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind AiGateway.
func (d *AiGatewayCustomDefaulter) Default(_ context.Context, obj runtime.Object) error {
	aiGateway, ok := obj.(*gatewayv1alpha1.AiGateway)

	if !ok {
		return fmt.Errorf("expected an AiGateway object but got %T", obj)
	}
	aigatewaylog.Info("Defaulting for AiGateway", "name", aiGateway.GetName())

	const DefaultPort = 4000
	if aiGateway.Spec.Port == 0 {
		aiGateway.Spec.Port = DefaultPort
	}

	return nil
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-gateway-agentic-layer-ai-v1alpha1-aigateway,mutating=false,failurePolicy=fail,sideEffects=None,groups=gateway.agentic-layer.ai,resources=aigateways,verbs=create;update,versions=v1alpha1,name=vaigateway-v1alpha1.kb.io,admissionReviewVersions=v1

// AiGatewayCustomValidator struct is responsible for validating the AiGateway resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type AiGatewayCustomValidator struct {
}

var _ webhook.CustomValidator = &AiGatewayCustomValidator{}

// Use a constant for the separator to avoid magic strings.
const providerModelSeparator = "/"

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type AiGateway.
func (v *AiGatewayCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	aiGateway, ok := obj.(*gatewayv1alpha1.AiGateway)
	if !ok {
		// This error is for the webhook runtime, not the user.
		return nil, fmt.Errorf("expected a AiGateway object but got %T", obj)
	}
	aigatewaylog.Info("Validation for AiGateway upon creation", "name", aiGateway.GetName())
	return v.validateAiGatewaySpec(aiGateway)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type AiGateway.
func (v *AiGatewayCustomValidator) ValidateUpdate(_ context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	aiGateway, ok := newObj.(*gatewayv1alpha1.AiGateway)
	if !ok {
		return nil, fmt.Errorf("expected a AiGateway object for the newObj but got %T", newObj)
	}
	aigatewaylog.Info("Validation for AiGateway upon update", "name", aiGateway.GetName())
	return v.validateAiGatewaySpec(aiGateway)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type AiGateway.
func (v *AiGatewayCustomValidator) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	aiGateway, ok := obj.(*gatewayv1alpha1.AiGateway)
	if !ok {
		return nil, fmt.Errorf("expected a AiGateway object but got %T", obj)
	}
	aigatewaylog.Info("Validation for AiGateway upon deletion", "name", aiGateway.GetName())

	return nil, nil
}

// validateAiGatewaySpec contains the core validation logic for the AiGateway spec.
// It's called by both ValidateCreate and ValidateUpdate.
func (v *AiGatewayCustomValidator) validateAiGatewaySpec(aiGateway *gatewayv1alpha1.AiGateway) (admission.Warnings, error) {
	// Validate port is positive
	if aiGateway.Spec.Port <= 0 {
		return nil, fmt.Errorf("aiGateway port must be positive, got: %d", aiGateway.Spec.Port)
	}

	// Validate at least one AI model is specified
	if len(aiGateway.Spec.AiModels) == 0 {
		return nil, errors.New("no AI models specified in AiGateway")
	}

	// Validate AI model names
	for _, model := range aiGateway.Spec.AiModels {
		if model.Name == "" {
			return nil, errors.New("AI model name cannot be empty")
		}

		provider, modelName, found := strings.Cut(model.Name, providerModelSeparator)

		// Handle malformed names like "gpt-4" (no separator) or "/gpt-4" (empty provider).
		if !found || provider == "" || modelName == "" {
			return nil, fmt.Errorf("model %q is malformed; format must be 'provider/model-name'", model.Name)
		}

		// LiteLLM supports a vast number of providers, so we don't validate against a specific list.
		// The controller will handle provider-specific configuration and the LiteLLM proxy will
		// validate the actual model availability at runtime.
	}

	return nil, nil
}
