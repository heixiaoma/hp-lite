package log

import (
	"log"

	daemon "github.com/kardianos/service"
)

var Log daemon.Logger

func Error(v ...interface{}) {
	if Log != nil {
		Log.Error(v)
	} else {
		log.Println(v)
	}
}

func Info(v ...interface{}) {
	if Log != nil {
		Log.Info(v)
	} else {
		log.Println(v)
	}
}

func Errorf(format string, a ...interface{}) {
	if Log != nil {
		Log.Errorf(format, a)
	} else {
		log.Printf(format, a)
	}
}

func Infof(format string, a ...interface{}) {
	if Log != nil {
		Log.Infof(format, a)
	} else {
		log.Printf(format, a)
	}
}
