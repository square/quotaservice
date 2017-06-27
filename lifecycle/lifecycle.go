// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package lifecycle

// The status type
type Status int

const (
	Stopped Status = iota
	Started
)

func (s Status) String() string {
	switch s {
	case Started:
		return "Started"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}
