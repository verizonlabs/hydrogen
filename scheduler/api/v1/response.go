package v1

// V1 API statuses.
const (
	ACCEPTED = "Accepted"
	LAUNCHED = "Launched"
	FAILED   = "Failed"
	RUNNING  = "Running"
	KILLED   = "Killed"
	NOTFOUND = "Not Found"
	QUEUED   = "Queued"
	UPDATE   = "Updated"
)

// V1 API response format.
type Response struct {
	Status   string
	TaskName string
	Message  string
}
