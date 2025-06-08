package pkgerror

// Code is an enum representing the standard business error code that can occur in the application.
type Code int

const (
	Generic Code = iota
)

func codeMessage() map[Code]string {
	return map[Code]string{
		Generic: "Error",
	}
}

func (c Code) String() string {
	message, ok := codeMessage()[c]

	if ok {
		return message
	}

	return "unknown error"
}
