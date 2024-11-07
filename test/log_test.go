package test

import (
	"errors"
	"fmt"
	klog2 "github.com/cylScripter/chest/log"
	"github.com/cylScripter/chest/rpc"
	"testing"
)

func Test(t *testing.T) {
	klog2.SetLevel(klog2.LevelDebug)
	klog2.Infof("hello")

	err := errors.New("hello")

	fmt.Println(rpc.FromError(err))
}
