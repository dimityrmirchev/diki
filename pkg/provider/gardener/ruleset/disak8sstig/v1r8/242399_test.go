// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1r8_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Masterminds/semver"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	kubernetesgardener "github.com/gardener/gardener/pkg/client/kubernetes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	manualfake "k8s.io/client-go/rest/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/diki/pkg/kubernetes/pod"
	fakepod "github.com/gardener/diki/pkg/kubernetes/pod/fake"
	"github.com/gardener/diki/pkg/provider/gardener"
	"github.com/gardener/diki/pkg/provider/gardener/ruleset/disak8sstig/v1r8"
	"github.com/gardener/diki/pkg/rule"
	dikirule "github.com/gardener/diki/pkg/rule"
)

var _ = Describe("#242399", func() {
	var (
		fakeControlPlaneClient client.Client
		fakeClusterClient      client.Client
		fakeClusterRESTClient  rest.Interface
		fakeClusterPodContext  pod.PodContext
		clusterVersion122      *semver.Version
		clusterVersion126      *semver.Version
		ctx                    = context.TODO()
		workers                *extensionsv1alpha1.Worker
		namespace              = "foo"
	)

	BeforeEach(func() {
		v1r8.Generator = &FakeRandString{CurrentChar: 'a'}
		fakeControlPlaneClient = fakeclient.NewClientBuilder().WithScheme(kubernetesgardener.SeedScheme).Build()
		fakeClusterClient = fakeclient.NewClientBuilder().WithScheme(kubernetesgardener.ShootScheme).Build()

		workers = &extensionsv1alpha1.Worker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker1",
				Namespace: namespace,
			},
			Spec: extensionsv1alpha1.WorkerSpec{
				Pools: []extensionsv1alpha1.WorkerPool{
					{
						Name: "pool1",
					},
					{
						Name: "pool2",
					},
					{
						Name: "pool3",
					},
				},
			},
		}

		Expect(fakeControlPlaneClient.Create(ctx, workers)).To(Succeed())

		node1 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
				Labels: map[string]string{
					"worker.gardener.cloud/pool": "pool1",
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		Expect(fakeClusterClient.Create(ctx, node1)).To(Succeed())

		node2 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node2",
				Labels: map[string]string{
					"worker.gardener.cloud/pool": "pool2",
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		Expect(fakeClusterClient.Create(ctx, node2)).To(Succeed())

		node3 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node3",
				Labels: map[string]string{
					"worker.gardener.cloud/pool": "pool3",
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		}
		Expect(fakeClusterClient.Create(ctx, node3)).To(Succeed())

		node4 := &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node4",
				Labels: map[string]string{
					"worker.gardener.cloud/pool": "pool2",
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		Expect(fakeClusterClient.Create(ctx, node4)).To(Succeed())

		fakeClusterRESTClient = &manualfake.RESTClient{
			GroupVersion:         schema.GroupVersion{Group: "", Version: "v1"},
			NegotiatedSerializer: scheme.Codecs,
			Client: manualfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
				switch req.URL.String() {
				case "https://localhost/nodes/node1/proxy/configz":
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(dynamicKubeletConfigAllowedNodeConfig)))}, nil
				case "https://localhost/nodes/node2/proxy/configz":
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(dynamicKubeletConfigNotAllowedNodeConfig)))}, nil
				case "https://localhost/nodes/node4/proxy/configz":
					return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(dynamicKubeletConfigNotSetNodeConfig)))}, nil
				default:
					return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(&bytes.Buffer{})}, nil
				}
			}),
		}

		clusterVersion122 = semver.MustParse("1.22.0")
		clusterVersion126 = semver.MustParse("1.26.0")
	})

	It("should return skipped checkResults when cluster Kubernetes version is >= v1.26", func() {
		expectedCheckResults := []dikirule.CheckResult{
			rule.SkippedCheckResult("Option featureGates.DynamicKubeletConfig removed in Kubernetes v1.26.", gardener.NewTarget("cluster", "shoot", "details", "Cluster uses Kubernetes 1.26.0.")),
		}
		rule := &v1r8.Rule242399{
			ClusterVersion: clusterVersion126,
		}

		ruleResult, err := rule.Run(ctx)
		Expect(err).To(BeNil())

		Expect(ruleResult.CheckResults).To(Equal(expectedCheckResults))
	})

	DescribeTable("Run cases",
		func(executeReturnString [][]string, executeReturnError [][]error, expectedCheckResults []dikirule.CheckResult) {
			alwaysExpectedCheckResults := []dikirule.CheckResult{
				rule.PassedCheckResult("Option featureGates.DynamicKubeletConfig set to allowed value.", gardener.NewTarget("cluster", "shoot", "kind", "node", "name", "node1")),
				rule.FailedCheckResult("Option featureGates.DynamicKubeletConfig set to not allowed value.", gardener.NewTarget("cluster", "shoot", "kind", "node", "name", "node2")),
				rule.WarningCheckResult("Node is not in Ready state.", gardener.NewTarget("cluster", "shoot", "kind", "node", "name", "node3")),
				rule.PassedCheckResult("Option featureGates.DynamicKubeletConfig not set.", gardener.NewTarget("cluster", "shoot", "kind", "node", "name", "node4")),
			}
			expectedCheckResults = append(expectedCheckResults, alwaysExpectedCheckResults...)
			fakeClusterPodContext = fakepod.NewFakeSimplePodContext(executeReturnString, executeReturnError)
			rule := &v1r8.Rule242399{
				Logger:                  testLogger,
				ControlPlaneClient:      fakeControlPlaneClient,
				ControlPlaneNamespace:   namespace,
				ClusterClient:           fakeClusterClient,
				ClusterVersion:          clusterVersion122,
				ClusterCoreV1RESTClient: fakeClusterRESTClient,
				ClusterPodContext:       fakeClusterPodContext,
			}

			ruleResult, err := rule.Run(ctx)
			Expect(err).To(BeNil())

			Expect(ruleResult.CheckResults).To(Equal(expectedCheckResults))
		},

		Entry("should return correct checkResults when execute errors, and one node has feature-gates kubelet flag set",
			[][]string{{""}, {"--feature-gates=DynamicKubeletConfig=true"}},
			[][]error{{fmt.Errorf("command stderr output: sh: 1: -c: not found")}, {nil}},
			[]dikirule.CheckResult{
				rule.ErroredCheckResult("command stderr output: sh: 1: -c: not found", gardener.NewTarget("cluster", "shoot", "kind", "pod", "namespace", "kube-system", "name", "diki-node-files-aaaaaaaaaa")),
				rule.FailedCheckResult("Use of deprecated kubelet config flag feature-gates.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool2")),
				rule.WarningCheckResult("There are no nodes in Ready state for worker group.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool3")),
			}),
		Entry("should return correct checkResults when nodes have featureGates.DynamicKubeletConfig set",
			[][]string{{"--not-feature-gates=DynamicKubeletConfig=true --config=./config", dynamicKubeletConfigAllowedConfig}, {"--not-feature-gates=DynamicKubeletConfig=true --config=./config", dynamicKubeletConfigNotAllowedConfig}},
			[][]error{{nil, nil}, {nil, nil}},
			[]dikirule.CheckResult{
				rule.PassedCheckResult("Option featureGates.DynamicKubeletConfig set to allowed value.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool1")),
				rule.FailedCheckResult("Option featureGates.DynamicKubeletConfig set to not allowed value.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool2")),
				rule.WarningCheckResult("There are no nodes in Ready state for worker group.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool3")),
			}),
		Entry("should return correct checkResults when nodes do not have featureGates.DynamicKubeletConfig set",
			[][]string{{"--not-feature-gates=DynamicKubeletConfig=true --config=./config", dynamicKubeletConfigNotSetConfig}, {"--not-feature-gates=DynamicKubeletConfig=true, --config=./config", dynamicKubeletConfigNotSetConfig}},
			[][]error{{nil, nil}, {nil, nil}},
			[]dikirule.CheckResult{
				rule.PassedCheckResult("Option featureGates.DynamicKubeletConfig not set.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool1")),
				rule.PassedCheckResult("Option featureGates.DynamicKubeletConfig not set.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool2")),
				rule.WarningCheckResult("There are no nodes in Ready state for worker group.", gardener.NewTarget("cluster", "seed", "kind", "workerGroup", "name", "pool3")),
			}),
	)
})

const (
	dynamicKubeletConfigAllowedConfig = `featureGates:
  DynamicKubeletConfig: false
`
	dynamicKubeletConfigNotAllowedConfig = `featureGates:
  DynamicKubeletConfig: true
`
	dynamicKubeletConfigNotSetConfig = `maxPods: 100
`
	dynamicKubeletConfigAllowedNodeConfig    = `{"kubeletconfig":{"featureGates":{"DynamicKubeletConfig":false}}}`
	dynamicKubeletConfigNotAllowedNodeConfig = `{"kubeletconfig":{"featureGates":{"DynamicKubeletConfig":true}}}`
	dynamicKubeletConfigNotSetNodeConfig     = `{"kubeletconfig":{"authentication":{"webhook":{"enabled":true,"cacheTTL":"2m0s"}}}}`
)
