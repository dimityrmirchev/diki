// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"slices"
)

// Rule defines what is considered a rule in the context of Diki.
type Rule interface {
	ID() string
	Name() string
	Run(ctx context.Context) (RuleResult, error)
}

// RuleResult contains a Rule identification and the results of a Rule run.
type RuleResult struct {
	RuleID, RuleName string
	CheckResults     []CheckResult
}

// Target is the thing that a check is performed against.
type Target any

// CheckResult contains information about a Rule check. Returned from Rule runs.
type CheckResult struct {
	Status  Status
	Message string
	Target  Target
}

// Status of a CheckResult
type Status string

const (
	// Passed status indicates that a check is satisfied.
	Passed Status = "Passed"
	// Skipped status indicates that a rule is skipped with explanation.
	Skipped Status = "Skipped"
	// Accepted status indicates that a check violation is accepted and justified
	// based on additional configuration.
	Accepted Status = "Accepted"
	// Warning status indicates that there is ambiguity and the check was not performed with confidence.
	Warning Status = "Warning"
	// Failed status indicates that a check reported a violation.
	Failed Status = "Failed"
	// Errored status indicates that an unexpected error occured during check execution.
	Errored Status = "Errored"
	// NotImplemented status indicates that a rule/check is not implemented.
	NotImplemented Status = "Not Implemented"
)

// Statuses returns all supported statuses.
func Statuses() []Status {
	return []Status{Passed, Skipped, Accepted, Warning, Failed, Errored, NotImplemented}
}

var orderedStatuses = []Status{Passed, Skipped, Accepted, Warning, Failed, Errored, NotImplemented}

// Less is used to define the priority of the statuses.
// The ascending order is as follows
// Passed, Skipped, Accepted, Warning, Failed, Errored, Not Implemented
func (a Status) Less(b Status) bool {
	i := slices.IndexFunc(orderedStatuses, func(s Status) bool {
		return a == s
	})

	x := slices.IndexFunc(orderedStatuses, func(s Status) bool {
		return b == s
	})

	return i < x
}

// GetStatusIcon returns the icon of a given Status string
func GetStatusIcon(status Status) rune {
	switch status {
	case Passed:
		return '🟢'
	case Failed, Errored:
		return '🔴'
	case Skipped, Accepted:
		return '🔵'
	case Warning, NotImplemented:
		return '🟠'
	default:
		return '⚪'
	}
}
