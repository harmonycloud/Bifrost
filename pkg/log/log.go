package log

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"os"
)

var logFile *os.File
var log *logrus.Logger

// CmdAdd and so on are different kinds of log entries for different packages
var (
	CmdAdd *logrus.Entry
	CmdDel *logrus.Entry
	Alloc  *logrus.Entry
	Etcd   *logrus.Entry
	IPPool *logrus.Entry
	Kube   *logrus.Entry
	Rest   *logrus.Entry
)

// Init open the log file and start logging
func Init(logPath, logLevel string) {
	var err error

	log = logrus.New()

	log.Level, err = logrus.ParseLevel(logLevel)
	if err != nil {
		log.Level = logrus.InfoLevel
		log.Info("Failed to parse log level, using info level")
	}

	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = logFile
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	log.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})

	CmdAdd = log.WithField("Selector", "CMDADD")
	CmdDel = log.WithField("Selector", "CMDDEL")
	Alloc = log.WithField("Selector", "ALLOC")
	Etcd = log.WithField("Selector", "ETCD")
	IPPool = log.WithField("Selector", "IPPOOL")
	Kube = log.WithField("Selector", "KUBE")
	Rest = log.WithField("Selector", "REST")

	initRotate(logPath)
}

// Close close the log file descriptor
func Close() {
	logFile.Close()
}

func initRotate(logPath string) {
	rotateConfig, err := os.Create("/etc/logrotate.d/hcipam")
	if err != nil {
		log.Info("Failed to create log rotate file")
	}
	defer rotateConfig.Close()

	rotateConfig.WriteString(logPath + " {\n" +
		"daily\n" +
		"rotate 7\n" +
		"missingok\n" +
		"notifempty\n" +
		"dateext\n" +
		"}")

}
