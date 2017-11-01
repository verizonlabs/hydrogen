package plan

// Idle plan does nothing.
type IdlePlan struct{}

func (i IdlePlan) Execute() error    { return nil }
func (i IdlePlan) Update() error     { return nil }
func (i IdlePlan) Status() PlanState { return nil }
func (i IdlePlan) Type() PlanType    { return Idle }

func NewIdlePlan() Plan {
	return IdlePlan{}
}
