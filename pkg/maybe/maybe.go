package maybe

type MaybeStatus int

const (
	Nothing MaybeStatus = iota
	Just
)

type Maybe interface {
	Just(interface{}) error
	Get() interface{}
}
