// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1r8_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/diki/pkg/provider/gardener"
	"github.com/gardener/diki/pkg/provider/gardener/ruleset/disak8sstig/v1r8"
	dikirule "github.com/gardener/diki/pkg/rule"
)

var _ = Describe("#242422", func() {
	var (
		fakeClient client.Client
		ctx        = context.TODO()
		namespace  = "foo"

		ksDeployment *appsv1.Deployment
		target       = gardener.NewTarget("cluster", "seed", "name", "kube-apiserver", "namespace", namespace, "kind", "deployment")
	)

	BeforeEach(func() {
		fakeClient = fakeclient.NewClientBuilder().Build()
		ksDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-apiserver",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "kube-apiserver",
								Command: []string{},
								Args:    []string{},
							},
						},
					},
				},
			},
		}
	})

	It("should error when kube-apiserver is not found", func() {
		rule := &v1r8.Rule242422{Logger: testLogger, Client: fakeClient, Namespace: namespace}

		ruleResult, err := rule.Run(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(ruleResult.CheckResults).To(Equal([]dikirule.CheckResult{
			{
				Status:  dikirule.Errored,
				Message: "deployments.apps \"kube-apiserver\" not found",
				Target:  target,
			},
		},
		))
	})

	DescribeTable("Run cases",
		func(container corev1.Container, expectedCheckResults []dikirule.CheckResult, errorMatcher gomegatypes.GomegaMatcher) {
			ksDeployment.Spec.Template.Spec.Containers = []corev1.Container{container}
			Expect(fakeClient.Create(ctx, ksDeployment)).To(Succeed())

			rule := &v1r8.Rule242422{Logger: testLogger, Client: fakeClient, Namespace: namespace}
			ruleResult, err := rule.Run(ctx)
			Expect(err).To(errorMatcher)

			Expect(ruleResult.CheckResults).To(Equal(expectedCheckResults))
		},

		Entry("should fail when tls-cert-file and tls-private-key-file are not set",
			corev1.Container{Name: "kube-apiserver", Command: []string{"--flag1=value1", "--flag2=value2"}},
			[]dikirule.CheckResult{
				{Status: dikirule.Failed, Message: "Option tls-cert-file has not been set.", Target: target},
				{Status: dikirule.Failed, Message: "Option tls-private-key-file has not been set.", Target: target}},
			BeNil()),
		Entry("should pass when tls-cert-file and tls-private-key-file are set",
			corev1.Container{Name: "kube-apiserver", Command: []string{"--tls-cert-file=set", "--tls-private-key-file=set"}},
			[]dikirule.CheckResult{
				{Status: dikirule.Passed, Message: "Option tls-cert-file set.", Target: target},
				{Status: dikirule.Passed, Message: "Option tls-private-key-file set.", Target: target}},
			BeNil()),
		Entry("should fail when tls-cert-file and tls-private-key-file are empty",
			corev1.Container{Name: "kube-apiserver", Command: []string{"--tls-cert-file", "--tls-private-key-file"}},
			[]dikirule.CheckResult{
				{Status: dikirule.Failed, Message: "Option tls-cert-file is empty.", Target: target},
				{Status: dikirule.Failed, Message: "Option tls-private-key-file is empty.", Target: target}},
			BeNil()),
		Entry("should return correct different check results",
			corev1.Container{Name: "kube-apiserver", Command: []string{"--tls-cert-file=set", "--tls-private-key-file"}},
			[]dikirule.CheckResult{
				{Status: dikirule.Passed, Message: "Option tls-cert-file set.", Target: target},
				{Status: dikirule.Failed, Message: "Option tls-private-key-file is empty.", Target: target}},
			BeNil()),
		Entry("should warn when tls-cert-file and tls-private-key-file are set more than once",
			corev1.Container{Name: "kube-apiserver", Command: []string{"--tls-cert-file=set1", "--tls-private-key-file=set1"}, Args: []string{"--tls-cert-file=set2", "--tls-private-key-file=set2"}},
			[]dikirule.CheckResult{
				{Status: dikirule.Warning, Message: "Option tls-cert-file has been set more than once in container command.", Target: target},
				{Status: dikirule.Warning, Message: "Option tls-private-key-file has been set more than once in container command.", Target: target}},
			BeNil()),
		Entry("should error when deployment does not have container 'kube-apiserver'",
			corev1.Container{Name: "not-kube-apiserver", Command: []string{"--tls-cert-file=true"}},
			[]dikirule.CheckResult{{Status: dikirule.Errored, Message: "deployment: kube-apiserver does not contain container: kube-apiserver", Target: target}},
			BeNil()),
	)
})
