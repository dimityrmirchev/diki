// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1r8

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki/pkg/provider/gardener"
	"github.com/gardener/diki/pkg/provider/gardener/internal/utils"
	"github.com/gardener/diki/pkg/rule"
)

var _ rule.Rule = &Rule242414{}

type Rule242414 struct {
	ControlPlaneClient    client.Client
	ControlPlaneNamespace string
	ClusterClient         client.Client
	Options               *Options242414
	Logger                *slog.Logger
}

type Options242414 struct {
	AcceptedPods []struct {
		PodNamePrefix       string
		NamespaceNamePrefix string
		Justification       string
		Ports               []int32
	}
}

func (r *Rule242414) ID() string {
	return ID242414
}

func (r *Rule242414) Name() string {
	return "Kubernetes cluster must use non-privileged host ports for user pods (MEDIUM 242414)"
}

func (r *Rule242414) Run(ctx context.Context) (rule.RuleResult, error) {
	seedPods, err := utils.GetAllPods(ctx, r.ControlPlaneClient, r.ControlPlaneNamespace, labels.NewSelector(), 300)
	seedTarget := gardener.NewTarget("cluster", "seed")
	shootTarget := gardener.NewTarget("cluster", "shoot")
	if err != nil {
		return rule.SingleCheckResult(r, rule.ErroredCheckResult(err.Error(), seedTarget.With("namespace", r.ControlPlaneNamespace, "kind", "podList"))), nil
	}

	checkResults := r.checkPods(ctx, seedPods, seedTarget)

	shootPods, err := utils.GetAllPods(ctx, r.ClusterClient, "", labels.NewSelector(), 300)
	if err != nil {
		return rule.RuleResult{
			RuleID:       r.ID(),
			RuleName:     r.Name(),
			CheckResults: append(checkResults, rule.ErroredCheckResult(err.Error(), shootTarget.With("kind", "podList"))),
		}, nil
	}

	checkResults = append(checkResults, r.checkPods(ctx, shootPods, shootTarget)...)

	return rule.RuleResult{
		RuleID:       r.ID(),
		RuleName:     r.Name(),
		CheckResults: checkResults,
	}, nil
}

func (r *Rule242414) checkPods(ctx context.Context, pods []corev1.Pod, clusterTarget gardener.Target) []rule.CheckResult {
	checkResults := []rule.CheckResult{}
	for _, pod := range pods {
		target := clusterTarget.With("name", pod.Name, "namespace", pod.Namespace, "kind", "pod")
		for _, container := range pod.Spec.Containers {
			uses := false
			if container.Ports != nil {
				for _, port := range container.Ports {
					if port.HostPort != 0 && port.HostPort < 1024 {
						target = target.With("details", fmt.Sprintf("containerName: %s, port: %d", container.Name, port.HostPort))
						if accepted, justification := r.accepted(pod.Name, pod.Namespace, port.HostPort); accepted {
							msg := "Container accepted to use hostPort < 1024."
							if justification != "" {
								msg = justification
							}
							checkResults = append(checkResults, rule.AcceptedCheckResult(msg, target))
						} else {
							checkResults = append(checkResults, rule.FailedCheckResult("Container may not use hostPort < 1024.", target))
						}
						uses = true
					}
				}
			}
			if !uses {
				checkResults = append(checkResults, rule.PassedCheckResult("Container does not use hostPort < 1024.", target))
			}
		}
	}
	return checkResults
}

func (r *Rule242414) accepted(podName, namespace string, hostPort int32) (bool, string) {
	for _, acceptedPods := range r.Options.AcceptedPods {
		if strings.HasPrefix(podName, acceptedPods.PodNamePrefix) && strings.HasPrefix(namespace, acceptedPods.NamespaceNamePrefix) {
			for _, acceptedHostPort := range acceptedPods.Ports {
				if acceptedHostPort == hostPort {
					return true, acceptedPods.Justification
				}
			}
		}
	}

	return false, ""
}
