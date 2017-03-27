package response

const (
	ACCEPTED = "Accepted"
	LAUNCHED = "Launched"
	FAILED   = "Failed"
	KILLED   = "Killed"
	NOTFOUND = "Not Found"
	QUEUED   = "Queued"
	UPDATE   = "Updated"
)

type Deploy struct {
	Status   string
	TaskName string
	Message  string
}

type Kill struct {
	Status   string
	TaskName string
	Message  string
}
