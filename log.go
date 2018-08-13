package mengine

type ELog interface {
	Debug(args ...interface{})
	Debugf(formate string, args ...interface{})
	Info(args ...interface{})
	Infof(formate string, args ...interface{})
	Notice(args ...interface{})
	Noticef(formate string, args ...interface{})
	Warning(args ...interface{})
	Warningf(formate string, args ...interface{})
	Error(args ...interface{})
	Errorf(formate string, args ...interface{})
}
