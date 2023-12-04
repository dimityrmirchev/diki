// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package disak8sstig

import (
	kubernetesgardener "github.com/gardener/gardener/pkg/client/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki/pkg/config"
	"github.com/gardener/diki/pkg/rule"
	sharedv1r11 "github.com/gardener/diki/pkg/shared/ruleset/disak8sstig/v1r11"
)

func (r *Ruleset) registerV1R11Rules(ruleOptions map[string]config.RuleOptionsConfig) error { //nolint:unused // TODO: add to FromGenericConfig
	runtimeClient, err := client.New(r.RuntimeConfig, client.Options{})
	if err != nil {
		return err
	}

	_, err = client.New(r.GardenConfig, client.Options{Scheme: kubernetesgardener.GardenScheme})
	if err != nil {
		return err
	}

	const (
		ns                      = "garden"
		etcdMain                = "virtual-garden-etcd-main"
		etcdEvents              = "virtual-garden-etcd-events"
		kcmDeploymentName       = "virtual-garden-kube-controller-manager"
		kcmContainerName        = "kube-controller-manager"
		apiserverDeploymentName = "virtual-garden-kube-apiserver"
		apiserverContainerName  = "kube-apiserver"
		noKubeletsMsg           = "The Virtual Garden cluster does not have any nodes therefore there are no kubelets to check."
		noPodsMsg               = "The Virtual Garden cluster does not have any nodes therefore there cluster does not have any pods."
	)
	rules := []rule.Rule{
		&sharedv1r11.Rule242376{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: kcmDeploymentName,
			ContainerName:  kcmContainerName,
		},
		rule.NewSkipRule(
			sharedv1r11.ID242377,
			"The Kubernetes Scheduler must use TLS 1.2, at a minimum, to protect the confidentiality of sensitive data during electronic dissemination (MEDIUM 242376)",
			"The Virtual Garden cluster does not make use of a Kubernetes Scheduler.",
			rule.Skipped,
		),
		&sharedv1r11.Rule242378{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		&sharedv1r11.Rule242379{
			Client:                runtimeClient,
			Namespace:             ns,
			StatefulSetETCDMain:   etcdMain,
			StatefulSetETCDEvents: etcdEvents,
		},
		&sharedv1r11.Rule242380{
			Client:                runtimeClient,
			Namespace:             ns,
			StatefulSetETCDMain:   etcdMain,
			StatefulSetETCDEvents: etcdEvents,
		},
		&sharedv1r11.Rule242381{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: kcmDeploymentName,
			ContainerName:  kcmContainerName,
		},
		&sharedv1r11.Rule242382{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		rule.NewSkipRule(
			sharedv1r11.ID242383,
			"User-managed resources must be created in dedicated namespaces (HIGH 242383)",
			"By design the Garden cluster provides separate namespaces for user projects and users do not have access to system namespaces.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242384,
			"The Kubernetes Scheduler must have secure binding (MEDIUM 242384)",
			"The Virtual Garden cluster does not make use of a Kubernetes Scheduler.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242385,
			"The Kubernetes Controller Manager must have secure binding (MEDIUM 242385)",
			"The Kubernetes Controller Manager runs in a container which already has limited access to network interfaces. In addition ingress traffic to the Kubernetes Controller Manager is restricted via network policies, making an unintended exposure less likely.",
			rule.Skipped,
		),
		&sharedv1r11.Rule242386{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		rule.NewSkipRule(
			sharedv1r11.ID242387,
			"The Kubernetes Kubelet must have the read-only port flag disabled (HIGH 242387)",
			noKubeletsMsg,
			rule.Skipped,
		),
		&sharedv1r11.Rule242388{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		&sharedv1r11.Rule242389{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		&sharedv1r11.Rule242390{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		rule.NewSkipRule(
			sharedv1r11.ID242391,
			"The Kubernetes Kubelet must have anonymous authentication disabled (HIGH 242391)",
			noKubeletsMsg,
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242392,
			"The Kubernetes kubelet must enable explicit authorization (HIGH 242392)",
			noKubeletsMsg,
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242393,
			"Kubernetes Worker Nodes must not have sshd service running (MEDIUM 242393)",
			"The Virtual Garden cluster does not have any nodes.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242394,
			"Kubernetes Worker Nodes must not have the sshd service enabled (MEDIUM 242394)",
			"The Virtual Garden cluster does not have any nodes.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242395,
			"Kubernetes dashboard must not be enabled (MEDIUM 242395)",
			"The Virtual Garden cluster does not have any nodes therefore it does not deploy a Kubernetes dashboard.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242396,
			"Kubernetes Kubectl cp command must give expected access and results (MEDIUM 242396)",
			"The Virtual Garden cluster does not have any nodes therefore it does not install kubectl.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242397,
			"Kubernetes kubelet static PodPath must not enable static pods (HIGH 242397)",
			"The Virtual Garden cluster does not have any nodes therefore there are no kubelets to check.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			// feature-gates.DynamicAuditing removed in v1.19. ref https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates-removed/
			sharedv1r11.ID242398,
			"Kubernetes DynamicAuditing must not be enabled (MEDIUM 242398)",
			"Option feature-gates.DynamicAuditing was removed in Kubernetes v1.19.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242399,
			"Kubernetes DynamicKubeletConfig must not be enabled (MEDIUM 242399)",
			noKubeletsMsg,
			rule.Skipped,
		),
		&sharedv1r11.Rule242400{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		&sharedv1r11.Rule242402{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
			ContainerName:  apiserverContainerName,
		},
		&sharedv1r11.Rule242403{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: apiserverDeploymentName,
		},
		rule.NewSkipRule(
			sharedv1r11.ID242404,
			"Kubernetes Kubelet must deny hostname override (MEDIUM 242404)",
			noKubeletsMsg,
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242405,
			"Kubernetes manifests must be owned by root (MEDIUM 242405)",
			"Gardener does not deploy any control plane component as systemd processes or static pod.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242406,
			"Kubernetes kubelet configuration file must be owned by root (MEDIUM 242406)",
			noKubeletsMsg,
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242407,
			"The Kubernetes KubeletConfiguration files must have file permissions set to 644 or more restrictive (MEDIUM 242407)",
			noKubeletsMsg,
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242408,
			"The Kubernetes manifest files must have least privileges (MEDIUM 242408)",
			"Gardener does not deploy any control plane component as systemd processes or static pod.",
			rule.Skipped,
		),
		&sharedv1r11.Rule242409{
			Client:         runtimeClient,
			Namespace:      ns,
			DeploymentName: kcmDeploymentName,
			ContainerName:  kcmContainerName,
		},
		rule.NewSkipRule(
			sharedv1r11.ID242410,
			"The Kubernetes API Server must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL) (MEDIUM 242410)",
			"Cannot be tested and should be enforced organizationally. Gardener uses a minimum of known and automatically opened/used/created ports/protocols/services (PPSM stands for Ports, Protocols, Service Management).",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242411,
			"The Kubernetes Scheduler must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL) (MEDIUM 242411)",
			"The Virtual Garden cluster does not make use of a Kubernetes Scheduler.",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242412,
			"The Kubernetes Controllers must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL) (MEDIUM 242412)",
			"Cannot be tested and should be enforced organizationally. Gardener uses a minimum of known and automatically opened/used/created ports/protocols/services (PPSM stands for Ports, Protocols, Service Management).",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242413,
			"The Kubernetes etcd must enforce ports, protocols, and services (PPS) that adhere to the Ports, Protocols, and Services Management Category Assurance List (PPSM CAL) (MEDIUM 242413)",
			"Cannot be tested and should be enforced organizationally. Gardener uses a minimum of known and automatically opened/used/created ports/protocols/services (PPSM stands for Ports, Protocols, Service Management).",
			rule.Skipped,
		),
		rule.NewSkipRule(
			sharedv1r11.ID242414,
			"The Kubernetes cluster must use non-privileged host ports for user pods (MEDIUM 242414)",
			noPodsMsg,
			rule.Skipped,
		),
	}

	for i, r := range rules {
		opt, found := ruleOptions[r.ID()]
		if found && opt.Skip != nil && opt.Skip.Enabled {
			rules[i] = rule.NewSkipRule(r.ID(), r.Name(), opt.Skip.Justification, rule.Accepted)
		}
	}

	return r.AddRules(rules...)
}
