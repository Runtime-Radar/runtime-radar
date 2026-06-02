package model

type RegistrationStatus uint8

const (
	RegistrationStatusNone RegistrationStatus = iota
	RegistrationStatusError
	RegistrationStatusOK
)

type Registration struct {
	Base
	TokenHash string
	Status    RegistrationStatus
	Error     string
}

func (rs RegistrationStatus) String() string {
	switch rs {
	case RegistrationStatusNone:
		return "none"
	case RegistrationStatusError:
		return "error"
	case RegistrationStatusOK:
		return "ok"
	default:
		return "unknown"
	}
}
