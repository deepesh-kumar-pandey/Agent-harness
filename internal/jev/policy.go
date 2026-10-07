package jev

type Policy interface {
	Evaluate(action Action) DecisionResult
}
