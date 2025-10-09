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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gatewayv1alpha1 "github.com/agentic-layer/ai-gateway-operator/api/v1alpha1"
)

var _ = Describe("AiGateway Webhook", func() {
	var (
		obj       *gatewayv1alpha1.AiGateway
		oldObj    *gatewayv1alpha1.AiGateway
		validator AiGatewayCustomValidator
		defaulter AiGatewayCustomDefaulter
	)

	BeforeEach(func() {
		obj = &gatewayv1alpha1.AiGateway{}
		oldObj = &gatewayv1alpha1.AiGateway{}
		validator = AiGatewayCustomValidator{}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		defaulter = AiGatewayCustomDefaulter{}
		Expect(defaulter).NotTo(BeNil(), "Expected defaulter to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
		// TODO (user): Add any setup logic common to all tests
	})

	AfterEach(func() {
		// TODO (user): Add any teardown logic common to all tests
	})

	Context("When creating AiGateway under Defaulting Webhook", func() {
		It("Should apply default port when port is not specified", func() {
			By("simulating a scenario where port is not set")
			obj.Spec.Port = 0
			By("calling the Default method to apply defaults")
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			By("checking that the default port is set")
			Expect(obj.Spec.Port).To(Equal(int32(4000)))
		})

		It("Should not override port when it is already set", func() {
			By("setting a custom port")
			obj.Spec.Port = 8080
			By("calling the Default method")
			Expect(defaulter.Default(ctx, obj)).To(Succeed())
			By("checking that the custom port is preserved")
			Expect(obj.Spec.Port).To(Equal(int32(8080)))
		})
	})

	Context("When creating or updating AiGateway under Validating Webhook", func() {
		It("Should deny creation if port is zero or negative", func() {
			By("creating an AiGateway with port 0")
			obj.Spec.Port = 0
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: "openai"},
			}
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be positive"))

			By("creating an AiGateway with negative port")
			obj.Spec.Port = -1
			_, err = validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("port must be positive"))
		})

		It("Should deny creation if no AI models are specified", func() {
			By("creating an AiGateway without AI models")
			obj.Spec.Port = 4000
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{}
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no AI models specified"))
		})

		It("Should deny creation if AI model name is empty", func() {
			By("creating an AiGateway with empty model name")
			obj.Spec.Port = 4000
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "", Provider: "openai"},
			}
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("AI model name cannot be empty"))
		})

		It("Should deny creation if AI model provider is empty", func() {
			By("creating an AiGateway with empty provider")
			obj.Spec.Port = 4000
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: ""},
			}
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("AI model provider cannot be empty"))
		})

		It("Should admit creation if all required fields are valid", func() {
			By("creating a valid AiGateway")
			obj.Spec.Port = 4000
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: "openai"},
				{Name: "claude-3-opus", Provider: "anthropic"},
			}
			_, err := validator.ValidateCreate(ctx, obj)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should validate updates correctly", func() {
			By("updating an AiGateway with valid values")
			oldObj.Spec.Port = 4000
			oldObj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: "openai"},
			}
			obj.Spec.Port = 8080
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: "openai"},
				{Name: "claude-3-opus", Provider: "anthropic"},
			}
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should deny update if validation fails", func() {
			By("updating an AiGateway with invalid model")
			oldObj.Spec.Port = 4000
			oldObj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: "openai"},
			}
			obj.Spec.Port = 4000
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "", Provider: "openai"},
			}
			_, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("AI model name cannot be empty"))
		})

		It("Should allow deletion without validation errors", func() {
			By("deleting an AiGateway")
			obj.Spec.Port = 4000
			obj.Spec.AiModels = []gatewayv1alpha1.AiModel{
				{Name: "gpt-4", Provider: "openai"},
			}
			_, err := validator.ValidateDelete(ctx, obj)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
