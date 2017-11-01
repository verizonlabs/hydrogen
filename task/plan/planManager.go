package plan

type (
	PlanType uint8

	PlanQueue interface {
		Push(PlanType)
		Pop() PlanType
		Peek() PlanType
	}

	PlanManager struct {
		current []PlanType
	}
)

const (
	Idle      PlanType = 0
	Launch    PlanType = 1
	Update    PlanType = 2
	Kill      PlanType = 3
	Reconcile PlanType = 4
)

func NewPlanManager() *PlanManager {
	return &PlanManager{
		current: make([]PlanType, 0),
	}
}

func (p *PlanManager) Push(planType PlanType) {
	p.current = append(p.current, planType)
}

func (p *PlanManager) Pop() PlanType {
	curr := Idle
	if len(p.current) > 0 {
		curr = p.current[0]
		p.current = p.current[1:len(p.current)]
	}
	return curr
}

func (p *PlanManager) Peek() PlanType {
	curr := Idle
	if len(p.current) > 0 {
		curr = p.current[0]
	}
	return curr
}
