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

func NewPlanManager() *PlanManager {
	return &PlanManager{
		current: make([]Plan, 0),
		idle:    NewIdlePlan(),
	}
}

func (p *PlanManager) Push(plan Plan) {
	p.current = append(p.current, plan)
	p.isReady <- p.current[0]
}

func (p *PlanManager) Pop() Plan {
	curr := p.idle
	if len(p.current) > 0 {
		curr = p.current[0]
		p.current = p.current[1:len(p.current)]
	}
	return curr
}

func (p *PlanManager) Peek() Plan {
	curr := p.idle
	if len(p.current) > 0 {
		curr = p.current[0]
	}
	return curr
}

func (p *PlanManager) Wait() chan Plan {
	return p.isReady
}
