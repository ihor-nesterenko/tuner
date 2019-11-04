package tuner

type Reader interface {
	Read(target interface{}) error
}
