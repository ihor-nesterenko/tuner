package tuner

type internalError struct {
	message string
}

func (i internalError) Error() string {
	return i.message
}
