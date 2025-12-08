package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
)

type IdentityServer struct {
}

// GetPluginInfo returns metadata about the plugin
func (ids *IdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.V(5).Infof("GetPluginInfo called")

	return &csi.GetPluginInfoResponse{
		Name:          "loop.csi.k8s.io",
		VendorVersion: "v1.0.0",
	}, nil
}

// GetPluginCapabilities returns the capabilities of the plugin
func (ids *IdentityServer) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	klog.V(5).Infof("GetPluginCapabilities called")

	// No capabilities needed for ephemeral inline volumes
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{},
	}, nil
}

// Probe checks if the plugin is ready
func (ids *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	klog.V(5).Infof("Probe called")
	return &csi.ProbeResponse{}, nil
}
