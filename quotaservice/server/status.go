package server

// The status type
type status int

const (
	stopped status = iota
	started
)

func (this status) String() string {
	switch this {
	case started:
		return "Started"
	case stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

