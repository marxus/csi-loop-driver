package driver

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/marxus/csi-loop-driver/conf"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

const backingFileDir = "/var/lib/csi-loop"

type NodeServer struct {
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
	backingFile := fmt.Sprintf("%s/%s.img", backingFileDir, volumeID)
	klog.Infof("Creating backing file: %s", backingFile)

	// Parse Kubernetes quantity format (1Gi, 500Mi) to bytes
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, fmt.Errorf("invalid size format %s: %v", size, err)
	}
	sizeBytes := quantity.Value()
	klog.Infof("Parsed size: %s -> %d bytes", size, sizeBytes)

	// Create the file with truncate
	cmd := exec.Command("truncate", "-s", fmt.Sprintf("%d", sizeBytes), conf.RealPath(backingFile))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create backing file: %v", err)
	}

	// Step 2: Format with mkfs.btrfs
	klog.Infof("Formatting with mkfs.btrfs")
	cmd = exec.Command("mkfs.btrfs", conf.RealPath(backingFile))
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to format: %v", err)
	}

	// Step 3: Mount with loop
	klog.Infof("Mounting to %s", targetPath)
	conf.FS.MkdirAll(targetPath, 0755)

	cmd = exec.Command("mount", "-o", "loop", conf.RealPath(backingFile), conf.RealPath(targetPath))
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
	cmd := exec.Command("umount", conf.RealPath(targetPath))
	if err := cmd.Run(); err != nil {
		klog.Warningf("Failed to unmount (may not be mounted): %v", err)
	}

	// Step 2: Remove backing file
	backingFile := fmt.Sprintf("%s/%s.img", backingFileDir, volumeID)
	conf.FS.Remove(backingFile)

	// Step 3: Remove mount directory
	conf.FS.Remove(targetPath)

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
