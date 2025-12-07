package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
)

// NodeServer implements the CSI Node service
type NodeServer struct {
	driver *Driver
	nodeID string
}

// NodePublishVolume mounts the volume to the target path
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	volumeContext := req.GetVolumeContext()

	size := volumeContext["size"]

	klog.Infof("NodePublishVolume: volumeID=%s, targetPath=%s, size=%s", volumeID, targetPath, size)

	// Step 1: Create backing file
	backingFile := fmt.Sprintf("/var/lib/csi-loop/%s.img", volumeID)
	klog.Infof("Creating backing file: %s", backingFile)

	// Make sure directory exists
	os.MkdirAll("/var/lib/csi-loop", 0755)

	// Create the file with truncate
	cmd := exec.Command("truncate", "-s", size, backingFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create backing file: %v", err)
	}

	// Step 2: Format with mkfs.xfs
	klog.Infof("Formatting with mkfs.xfs")
	cmd = exec.Command("mkfs.xfs", backingFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to format: %v", err)
	}

	// Step 3: Mount with loop
	klog.Infof("Mounting to %s", targetPath)
	os.MkdirAll(targetPath, 0755)

	cmd = exec.Command("mount", "-o", "loop", backingFile, targetPath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to mount: %v", err)
	}

	klog.Infof("Volume %s successfully mounted", volumeID)
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume
func (ns *NodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	klog.Infof("NodeUnpublishVolume: volumeID=%s, targetPath=%s", volumeID, targetPath)

	// Step 1: Unmount
	cmd := exec.Command("umount", targetPath)
	if err := cmd.Run(); err != nil {
		klog.Warningf("Failed to unmount (may not be mounted): %v", err)
	}

	// Step 2: Remove backing file
	backingFile := fmt.Sprintf("/var/lib/csi-loop/%s.img", volumeID)
	os.Remove(backingFile)

	// Step 3: Remove mount directory
	os.Remove(targetPath)

	klog.Infof("Volume %s successfully unpublished", volumeID)
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetCapabilities returns node capabilities
func (ns *NodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{},
	}, nil
}

// NodeGetInfo returns node information
func (ns *NodeServer) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return &csi.NodeGetInfoResponse{
		NodeId: ns.nodeID,
	}, nil
}

// Stubs for unused methods
func (ns *NodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ns *NodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ns *NodeServer) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ns *NodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
