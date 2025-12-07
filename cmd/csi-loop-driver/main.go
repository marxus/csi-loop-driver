package main

import (
	"flag"

	"github.com/k8s-loop-volume/csi-loop-driver/pkg/driver"
	"k8s.io/klog/v2"
)

var (
	endpoint = flag.String("endpoint", "unix:///csi/csi.sock", "CSI endpoint")
	nodeID   = flag.String("nodeid", "", "Node ID")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	if *nodeID == "" {
		klog.Fatal("Node ID is required (use --nodeid flag)")
	}

	klog.Infof("Starting CSI Loop Volume Driver")
	klog.Infof("Node ID: %s", *nodeID)
	klog.Infof("Endpoint: %s", *endpoint)

	// Create and run driver
	drv, err := driver.NewDriver(*nodeID, *endpoint)
	if err != nil {
		klog.Fatalf("Failed to create driver: %v", err)
	}

	if err := drv.Run(); err != nil {
		klog.Fatalf("Failed to run driver: %v", err)
	}
}
