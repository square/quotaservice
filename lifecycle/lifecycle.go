package lifecycle

// The status type
type Status int

const (
	Stopped Status = iota
	Started
)

func (this Status) String() string {
	switch this {
	case Started:
		return "Started"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}
