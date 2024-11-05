// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package retry_test

import (
	"context"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/diki/pkg/rule"
	"github.com/gardener/diki/pkg/rule/retry"
)

var counter = 0
var _ rule.Rule = &simpleRule{}

type simpleRule struct{}

func (r *simpleRule) ID() string {
	return "1"
}

func (r *simpleRule) Name() string {
	return "Simple rule"
}

func (r *simpleRule) Run(_ context.Context) (rule.RuleResult, error) {
	counter++
	if counter >= 2 {
		return rule.Result(r, rule.ErroredCheckResult("bar", rule.NewTarget())), nil
	}
	return rule.Result(r, rule.ErroredCheckResult("foo", rule.NewTarget())), nil
}

var _ = Describe("retryablerule", func() {
	Describe("#RetryableRule", func() {
		var (
			trueRetryCondition = func(_ rule.RuleResult) bool {
				return true
			}
			falseRetryCondition = func(_ rule.RuleResult) bool {
				return false
			}
			simpleRetryCondition = func(ruleResult rule.RuleResult) bool {
				for _, checkResult := range ruleResult.CheckResults {
					if checkResult.Status == rule.Errored {
						if checkResult.Message == "foo" {
							return true
						}
					}
				}
				return false
			}
			ctx = context.TODO()
		)
		BeforeEach(func() {
			counter = 0
		})

		DescribeTable("Run cases", func(retryCondition func(ruleResult rule.RuleResult) bool, maxRetries, expectedCounter int) {
			sr := &simpleRule{}
			rr := retry.New(
				retry.WithBaseRule(sr),
				retry.WithMaxRetries(maxRetries),
				retry.WithRetryCondition(retryCondition),
				retry.WithLogger(testLogger),
			)

			_, err := rr.Run(ctx)

			Expect(err).To(BeNil())
			Expect(counter).To(Equal(expectedCounter))
		},
			Entry("should hit maxRetry when retry condition is always met", trueRetryCondition, 2, 3),
			Entry("should not retry when retry condition is not met", falseRetryCondition, 7, 1),
			Entry("should retry until retry condition is not met", simpleRetryCondition, 7, 2),
		)
	})

	Describe("#RetryConditionFromRegex", func() {
		var (
			fooRegex       = regexp.MustCompile(`(?i)(foo)`)
			barRegex       = regexp.MustCompile(`(?i)(bar)`)
			fooCheckResult rule.CheckResult
			barCheckResult rule.CheckResult
			simpleRule     simpleRule
		)

		BeforeEach(func() {
			fooCheckResult = rule.CheckResult{
				Status:  rule.Errored,
				Message: "foo",
				Target:  rule.NewTarget(),
			}
			barCheckResult = rule.CheckResult{
				Status:  rule.Errored,
				Message: "bar",
				Target:  rule.NewTarget(),
			}
		})

		It("should create retry condition from a single regex", func() {
			rc := retry.RetryConditionFromRegex(*fooRegex)

			result := rc(rule.Result(&simpleRule, fooCheckResult))
			Expect(result).To(Equal(true))

			result = rc(rule.Result(&simpleRule, barCheckResult))
			Expect(result).To(Equal(false))

			fooCheckResult.Status = rule.Passed
			result = rc(rule.Result(&simpleRule, fooCheckResult))
			Expect(result).To(Equal(false))
		})
		It("should create retry condition from multiple regexes", func() {
			rc := retry.RetryConditionFromRegex(*fooRegex, *barRegex)

			result := rc(rule.Result(&simpleRule, fooCheckResult))
			Expect(result).To(Equal(true))

			result = rc(rule.Result(&simpleRule, barCheckResult))
			Expect(result).To(Equal(true))

			fooCheckResult.Status = rule.Passed
			result = rc(rule.Result(&simpleRule, fooCheckResult))
			Expect(result).To(Equal(false))
		})
	})
})
