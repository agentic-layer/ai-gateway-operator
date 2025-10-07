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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	agenticlayeraiv1alpha1 "github.com/agentic-layer/ai-gateway-operator/api/v1alpha1"
)

const (
	testController = "test-controller"
)

var _ = Describe("AiGatewayClass Webhook", func() {
	var (
		obj       *agenticlayeraiv1alpha1.AiGatewayClass
		oldObj    *agenticlayeraiv1alpha1.AiGatewayClass
		validator AiGatewayClassCustomValidator
	)

	BeforeEach(func() {
		obj = &agenticlayeraiv1alpha1.AiGatewayClass{}
		oldObj = &agenticlayeraiv1alpha1.AiGatewayClass{}
		validator = AiGatewayClassCustomValidator{Client: k8sClient}
		Expect(validator).NotTo(BeNil(), "Expected validator to be initialized")
		Expect(oldObj).NotTo(BeNil(), "Expected oldObj to be initialized")
		Expect(obj).NotTo(BeNil(), "Expected obj to be initialized")
	})

	AfterEach(func() {
		// Clean up any AiGatewayClass resources created during tests
		var classList agenticlayeraiv1alpha1.AiGatewayClassList
		_ = k8sClient.List(ctx, &classList)
		for _, class := range classList.Items {
			_ = k8sClient.Delete(ctx, &class)
		}
	})

	Context("When creating AiGatewayClass under Validating Webhook", func() {
		It("Should allow creation when no default class annotation is set", func() {
			By("Creating a AiGatewayClass without default annotation")
			obj.SetName("test-class-no-default")
			obj.Spec.Controller = testController

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())
		})

		It("Should allow creation when this is the first default class", func() {
			By("Creating a AiGatewayClass with default annotation")
			obj.SetName("test-class-first-default")
			obj.Spec.Controller = testController
			obj.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())
		})

		It("Should deny creation when another default class already exists", func() {
			By("Creating the first default class")
			existingClass := &agenticlayeraiv1alpha1.AiGatewayClass{}
			existingClass.SetName("existing-default-class")
			existingClass.SetNamespace("default")
			existingClass.Spec.Controller = testController
			existingClass.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})
			Expect(k8sClient.Create(ctx, existingClass)).To(Succeed())

			By("Attempting to create a second default class")
			obj.SetName("test-class-second-default")
			obj.Spec.Controller = testController
			obj.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})

			warnings, err := validator.ValidateCreate(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("another AiGatewayClass"))
			Expect(err.Error()).To(ContainSubstring("existing-default-class"))
			Expect(warnings).To(BeNil())

			By("Cleaning up the existing default class")
			Expect(k8sClient.Delete(ctx, existingClass)).To(Succeed())
		})

		It("Should return error when validating wrong object type", func() {
			By("Passing a wrong object type to ValidateCreate")
			wrongObj := &agenticlayeraiv1alpha1.AiGateway{}

			warnings, err := validator.ValidateCreate(ctx, wrongObj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected a AiGatewayClass object"))
			Expect(warnings).To(BeNil())
		})
	})

	Context("When updating AiGatewayClass under Validating Webhook", func() {
		It("Should allow update when no default annotation is involved", func() {
			By("Creating a non-default class")
			obj.SetName("test-class-update-no-default")
			obj.SetNamespace("default")
			obj.Spec.Controller = testController
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			By("Updating the class without default annotation")
			oldObj = obj.DeepCopy()
			obj.Spec.Controller = "updated-controller"

			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})

		It("Should allow update when keeping the same default class", func() {
			By("Creating a default class")
			obj.SetName("test-class-update-same-default")
			obj.SetNamespace("default")
			obj.Spec.Controller = testController
			obj.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			By("Updating the same default class")
			oldObj = obj.DeepCopy()
			obj.Spec.Controller = "updated-controller"

			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})

		It("Should deny update when trying to set default while another class is already default", func() {
			By("Creating the first default class")
			existingClass := &agenticlayeraiv1alpha1.AiGatewayClass{}
			existingClass.SetName("existing-default-class-update")
			existingClass.SetNamespace("default")
			existingClass.Spec.Controller = testController
			existingClass.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})
			Expect(k8sClient.Create(ctx, existingClass)).To(Succeed())

			By("Creating a non-default class")
			obj.SetName("test-class-update-to-default")
			obj.SetNamespace("default")
			obj.Spec.Controller = testController
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			By("Attempting to update it to be default")
			oldObj = obj.DeepCopy()
			obj.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})

			warnings, err := validator.ValidateUpdate(ctx, oldObj, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("another AiGatewayClass"))
			Expect(err.Error()).To(ContainSubstring("existing-default-class-update"))
			Expect(warnings).To(BeNil())

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, existingClass)).To(Succeed())
			Expect(k8sClient.Delete(ctx, obj)).To(Succeed())
		})

		It("Should return error when validating wrong object type", func() {
			By("Passing a wrong object type to ValidateUpdate")
			wrongObj := &agenticlayeraiv1alpha1.AiGateway{}
			wrongOldObj := &agenticlayeraiv1alpha1.AiGateway{}

			warnings, err := validator.ValidateUpdate(ctx, wrongOldObj, wrongObj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected a AiGatewayClass object"))
			Expect(warnings).To(BeNil())
		})
	})

	Context("When deleting AiGatewayClass under Validating Webhook", func() {
		It("Should always allow deletion", func() {
			By("Creating a AiGatewayClass")
			obj.SetName("test-class-delete")
			obj.Spec.Controller = testController
			obj.SetAnnotations(map[string]string{
				DefaultClassAnnotation: "true",
			})

			By("Validating deletion")
			warnings, err := validator.ValidateDelete(ctx, obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())
		})

		It("Should allow deletion even for wrong object type", func() {
			By("Passing a wrong object type to ValidateDelete")
			wrongObj := &agenticlayeraiv1alpha1.AiGateway{}

			warnings, err := validator.ValidateDelete(ctx, wrongObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(BeNil())
		})
	})

})
