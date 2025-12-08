package driver

import (
	"fmt"
	"net"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/marxus/csi-loop-driver/conf"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// RunDriver starts the CSI driver
func RunDriver() error {
	socketAddr := "/csi/csi.sock"

	klog.Infof("Starting CSI driver: nodeID=%s, socketAddr=%s", conf.NodeId, socketAddr)

	conf.FS.Remove(socketAddr)
	listener, err := net.Listen("unix", conf.RealPath(socketAddr))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	csi.RegisterIdentityServer(server, &IdentityServer{})
	csi.RegisterNodeServer(server, &NodeServer{nodeID: conf.NodeId})

	klog.Infof("Starting gRPC server on unix://%s", socketAddr)
	return server.Serve(listener)
}
