package jev

type Evaluator interface {
	Evaluate(action Action) DecisionResult
}
