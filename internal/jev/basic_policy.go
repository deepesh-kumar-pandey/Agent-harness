package jev

type BasicPolicy struct{}

func (p BasicPolicy) Evaluate(action Action) DecisionResult {
	if action.Tool == "" {
		return DecisionResult{
			Decision: Deny,
			Reason:   "action has no tool",
		}
	}

	return DecisionResult{
		Decision: Allow,
		Reason:   "action is allowed by default policy",
	}
}
