package gocket

type (
	Logger interface {
		Info(args ...interface{})
		Debug(args ...interface{})
		Error(args ...interface{})
		Fatal(args ...interface{})
	}
)
