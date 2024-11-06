package test

import (
	klog2 "github.com/cylScripter/chest/klog"
	"testing"
)

func Test(t *testing.T) {
	klog2.SetLevel(klog2.LevelDebug)
	klog2.SetLogger(klog2.NewZapLogger())
	klog2.Infof("hello")

	log := klog2.NewZapLogger()
	log.Infof("hello")
}
