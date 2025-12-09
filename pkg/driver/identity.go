// Package driver implements the CSI Identity and Node services for the loop volume driver.
// It provides ephemeral node-local volumes backed by btrfs-formatted loop devices.
package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
)

// IdentityServer implements the CSI Identity service.
// It provides plugin metadata and capabilities.
type IdentityServer struct {
}

// GetPluginInfo returns metadata about the plugin.
// It provides the plugin name and version for CSI driver registration.
func (ids *IdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.V(5).Infof("GetPluginInfo called")
	return &csi.GetPluginInfoResponse{
		Name:          "loop.csi.k8s.io",
		VendorVersion: "v1.0.0",
	}, nil
}

// GetPluginCapabilities returns the capabilities of the plugin.
// Returns an empty list since ephemeral inline volumes require no special capabilities.
func (ids *IdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.V(5).Infof("GetPluginCapabilities called")
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{},
	}, nil
}

// Probe checks if the plugin is ready to serve requests.
// Always returns success since this driver has no external dependencies.
func (ids *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.V(5).Infof("Probe called")
	return &csi.ProbeResponse{}, nil
}
