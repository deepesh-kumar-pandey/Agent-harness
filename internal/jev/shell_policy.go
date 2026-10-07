package jev

type ShellPolicy struct{}

func (p ShellPolicy) Evaluate(action Action) DecisionResult {
	if action.Tool != "shell" {
		return DecisionResult{
			Decision: Allow,
			Reason:   "action is not a shell action",
		}
	}

	return DecisionResult{
		Decision: Confirm,
		Reason:   "shell action requires confirmation",
	}
}
