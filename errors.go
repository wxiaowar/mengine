package mengine

type Error interface {
	Code() int32
	Msg() string
	Detail() string
}
