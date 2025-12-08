package main

import (
	"github.com/marxus/csi-loop-driver/pkg/driver"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)

	if err := driver.RunDriver(); err != nil {
		klog.Fatalf("Failed to run driver: %v", err)
	}
}
