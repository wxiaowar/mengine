package mengine

type Router interface {
	Rout(path string) (h HFunc, b bool)
}
