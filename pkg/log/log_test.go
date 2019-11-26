package log

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	log = logrus.New()
	log.Level = logrus.DebugLevel
	log.Out = os.Stdout

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

	CmdAdd.Errorf("test error msg")
}
