package response

const (
	ACCEPTED = "Accepted"
	LAUNCHED = "Launched"
	FAILED   = "Failed"
)

type Deploy struct {
	Status string
	Task   string
}
