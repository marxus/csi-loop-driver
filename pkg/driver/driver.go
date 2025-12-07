package driver

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

type Driver struct {
	nodeID   string
	endpoint string

	identityServer   *IdentityServer
	controllerServer *ControllerServer
	nodeServer       *NodeServer

	server *grpc.Server
}

// NewDriver creates a new CSI driver
func NewDriver(nodeID, endpoint string) (*Driver, error) {
	klog.Infof("Creating CSI driver: nodeID=%s, endpoint=%s", nodeID, endpoint)

	driver := &Driver{
		nodeID:   nodeID,
		endpoint: endpoint,
	}

	// Create the service servers
	driver.identityServer = &IdentityServer{driver: driver}
	driver.controllerServer = &ControllerServer{driver: driver}
	driver.nodeServer = &NodeServer{driver: driver, nodeID: nodeID}

	return driver, nil
}

// Run starts the CSI driver
func (d *Driver) Run() error {
	// Parse endpoint (e.g., "unix:///csi/csi.sock")
	scheme, addr := parseEndpoint(d.endpoint)

	// Remove existing socket if it exists
	if scheme == "unix" {
		os.Remove(addr)
	}

	// Create listener
	listener, err := net.Listen(scheme, addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create gRPC server
	d.server = grpc.NewServer()

	// Register our services
	csi.RegisterIdentityServer(d.server, d.identityServer)
	csi.RegisterControllerServer(d.server, d.controllerServer)
	csi.RegisterNodeServer(d.server, d.nodeServer)

	klog.Infof("Starting gRPC server on %s://%s", scheme, addr)
	return d.server.Serve(listener)
}

func parseEndpoint(endpoint string) (string, string) {
	// Handle "unix:///path" or "tcp://host:port"
	if strings.HasPrefix(endpoint, "unix://") {
		return "unix", strings.TrimPrefix(endpoint, "unix://")
	}
	if strings.HasPrefix(endpoint, "tcp://") {
		return "tcp", strings.TrimPrefix(endpoint, "tcp://")
	}
	// Default: assume it's a unix socket path
	return "unix", endpoint
}
