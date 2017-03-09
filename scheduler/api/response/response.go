package response

const (
	ACCEPTED = "Accepted"
	LAUNCHED = "Launched"
	FAILED   = "Failed"
	KILLED   = "Killed"
	NOTFOUND = "Not Found"
	QUEUED   = "Queued"
)

type Deploy struct {
	Status   string
	TaskName string
}

type Kill struct {
	Status   string
	TaskName string
}
