// Package serve provides high-level functions for starting the CSI loop driver.
// It handles driver initialization and gRPC server startup.
package serve

import (
	"fmt"
	"net"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/marxus/csi-loop-driver/conf"
	"github.com/marxus/csi-loop-driver/pkg/driver"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

const socketAddress = "/csi/csi.sock"

// StartDriver starts the CSI loop driver server.
// It validates the NODE_ID environment variable, creates the gRPC server,
// registers the CSI services, and starts listening on the Unix socket.
//
// Returns an error if NODE_ID is missing, socket creation fails, or server startup fails.
func StartDriver() error {
	if conf.NodeId == "" {
		return fmt.Errorf("NODE_ID environment variable is required")
	}

	klog.Infof("Starting CSI driver: nodeID=%s, socketAddr=%s", conf.NodeId, socketAddress)

	// Remove existing socket if it exists
	conf.FS.Remove(socketAddress)

	listener, err := net.Listen("unix", conf.RealPath(socketAddress))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	csi.RegisterIdentityServer(server, &driver.IdentityServer{})
	csi.RegisterNodeServer(server, &driver.NodeServer{NodeId: conf.NodeId})

	klog.Infof("Starting gRPC server on unix://%s", socketAddress)
	return server.Serve(listener)
}
