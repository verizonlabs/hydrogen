package plan

type (
	PlanQueue interface {
		Push(Plan)
		Pop() Plan
		Peek() Plan
		Wait() chan Plan
	}

	PlanManager struct {
		current []Plan
		idle    Plan
		isReady chan Plan
	}
)

// NewPlanQueue creates a new queue for plans.
func NewPlanQueue() *PlanManager {
	return &PlanManager{
		current: make([]Plan, 0),
		idle:    NewIdlePlan(),
	}
}

// Push adds a new plan onto the queue.
func (p *PlanManager) Push(plan Plan) {
	p.current = append(p.current, plan)
	p.isReady <- p.current[0]
}

// Pop moves the queue up one, removing first Plan.
func (p *PlanManager) Pop() Plan {
	curr := p.idle
	if len(p.current) > 0 {
		curr = p.current[0]
		p.current = p.current[1:len(p.current)]
	}
	return curr
}

// Peek looks at the next Plan to be executed.
func (p *PlanManager) Peek() Plan {
	curr := p.idle
	if len(p.current) > 0 {
		curr = p.current[0]
	}
	return curr
}

// Wait exposes the channel is the queue is empty to wait for new plans.
func (p *PlanManager) Wait() chan Plan {
	return p.isReady
}
