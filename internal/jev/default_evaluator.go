package jev

type DefaultEvaluator struct {
	policies []Policy
}

func NewDefaultEvaluator() *DefaultEvaluator {
	return &DefaultEvaluator{
		policies: []Policy{
			BasicPolicy{},
			ShellPolicy{},
		},
	}
}

func (e DefaultEvaluator) Evaluate(action Action) DecisionResult {
	var result DecisionResult
	result.Decision = Allow

	for _, policy := range e.policies {
		policyResult := policy.Evaluate(action)

		if policyResult.Decision == Deny {
			return policyResult
		}

		if policyResult.Decision == Confirm {
			result = policyResult
		}
	}

	return result
}
