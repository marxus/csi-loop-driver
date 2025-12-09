// Node service volume operations tests.
package driver

import (
	"context"
	"fmt"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/marxus/csi-loop-driver/conf"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeServer_GetInfo(t *testing.T) {
	ns := &NodeServer{NodeId: "test-node-123"}

	resp, err := ns.NodeGetInfo(context.Background(), &csi.NodeGetInfoRequest{})

	require.NoError(t, err)
	assert.Equal(t, "test-node-123", resp.NodeId)
}

func TestNodeServer_GetCapabilities(t *testing.T) {
	ns := &NodeServer{NodeId: "test-node"}

	resp, err := ns.NodeGetCapabilities(context.Background(), &csi.NodeGetCapabilitiesRequest{})

	require.NoError(t, err)
	assert.Empty(t, resp.Capabilities, "ephemeral-only driver should have no node capabilities")
}

func TestNodeServer_PublishVolume(t *testing.T) {
	tests := []struct {
		name            string
		volumeID        string
		size            string
		targetPath      string
		mockCommands    map[string]error
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "successfully publishes 1Gi volume",
			volumeID:   "vol-123",
			size:       "1Gi",
			targetPath: "/mnt/test",
			mockCommands: map[string]error{
				"truncate":    nil,
				"mkfs.btrfs":  nil,
				"mount":       nil,
			},
			wantErr: false,
		},
		{
			name:       "successfully publishes 500Mi volume",
			volumeID:   "vol-456",
			size:       "500Mi",
			targetPath: "/mnt/test2",
			mockCommands: map[string]error{
				"truncate":    nil,
				"mkfs.btrfs":  nil,
				"mount":       nil,
			},
			wantErr: false,
		},
		{
			name:            "fails on invalid size format",
			volumeID:        "vol-789",
			size:            "invalid-size",
			targetPath:      "/mnt/test3",
			wantErr:         true,
			wantErrContains: "invalid size format",
		},
		{
			name:       "fails when truncate fails",
			volumeID:   "vol-fail",
			size:       "1Gi",
			targetPath: "/mnt/fail",
			mockCommands: map[string]error{
				"truncate": fmt.Errorf("disk full"),
			},
			wantErr:         true,
			wantErrContains: "failed to create backing file",
		},
		{
			name:       "fails when mkfs fails",
			volumeID:   "vol-mkfs-fail",
			size:       "1Gi",
			targetPath: "/mnt/mkfs-fail",
			mockCommands: map[string]error{
				"truncate":   nil,
				"mkfs.btrfs": fmt.Errorf("mkfs error"),
			},
			wantErr:         true,
			wantErrContains: "failed to format",
		},
		{
			name:       "fails when mount fails",
			volumeID:   "vol-mount-fail",
			size:       "1Gi",
			targetPath: "/mnt/mount-fail",
			mockCommands: map[string]error{
				"truncate":   nil,
				"mkfs.btrfs": nil,
				"mount":      fmt.Errorf("mount error"),
			},
			wantErr:         true,
			wantErrContains: "failed to mount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			originalRunCommand := conf.RunCommand
			defer func() { conf.RunCommand = originalRunCommand }()

			conf.RunCommand = func(name string, args ...string) error {
				if tt.mockCommands != nil {
					if err, ok := tt.mockCommands[name]; ok {
						return err
					}
				}
				return nil
			}

			// Clean up test files
			backingFile := fmt.Sprintf("%s/%s.img", backingFileDir, tt.volumeID)
			defer conf.FS.Remove(backingFile)
			defer conf.FS.Remove(tt.targetPath)

			ns := &NodeServer{NodeId: "test-node"}
			req := &csi.NodePublishVolumeRequest{
				VolumeId:   tt.volumeID,
				TargetPath: tt.targetPath,
				VolumeContext: map[string]string{
					"size": tt.size,
				},
			}

			resp, err := ns.NodePublishVolume(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)

				// Verify directories were created
				exists, err := afero.DirExists(conf.FS, backingFileDir)
				require.NoError(t, err)
				assert.True(t, exists, "backing file directory should exist")

				exists, err = afero.DirExists(conf.FS, tt.targetPath)
				require.NoError(t, err)
				assert.True(t, exists, "target path should exist")
			}
		})
	}
}

func TestNodeServer_UnpublishVolume(t *testing.T) {
	tests := []struct {
		name         string
		volumeID     string
		targetPath   string
		setupFiles   bool
		mockUmount   error
		wantErr      bool
	}{
		{
			name:       "successfully unpublishes volume",
			volumeID:   "vol-123",
			targetPath: "/mnt/test",
			setupFiles: true,
			mockUmount: nil,
			wantErr:    false,
		},
		{
			name:       "handles umount failure gracefully",
			volumeID:   "vol-456",
			targetPath: "/mnt/test2",
			setupFiles: true,
			mockUmount: fmt.Errorf("not mounted"),
			wantErr:    false, // umount failure is logged, not returned
		},
		{
			name:       "handles missing files gracefully",
			volumeID:   "vol-789",
			targetPath: "/mnt/test3",
			setupFiles: false,
			mockUmount: fmt.Errorf("not mounted"),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			originalRunCommand := conf.RunCommand
			defer func() { conf.RunCommand = originalRunCommand }()

			conf.RunCommand = func(name string, args ...string) error {
				if name == "umount" {
					return tt.mockUmount
				}
				return nil
			}

			// Setup test files if needed
			backingFile := fmt.Sprintf("%s/%s.img", backingFileDir, tt.volumeID)
			if tt.setupFiles {
				conf.FS.MkdirAll(backingFileDir, 0755)
				afero.WriteFile(conf.FS, backingFile, []byte("fake-image"), 0644)
				conf.FS.MkdirAll(tt.targetPath, 0755)
			}

			ns := &NodeServer{NodeId: "test-node"}
			req := &csi.NodeUnpublishVolumeRequest{
				VolumeId:   tt.volumeID,
				TargetPath: tt.targetPath,
			}

			resp, err := ns.NodeUnpublishVolume(context.Background(), req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)

				// Verify files were removed
				exists, _ := afero.Exists(conf.FS, backingFile)
				assert.False(t, exists, "backing file should be removed")

				exists, _ = afero.Exists(conf.FS, tt.targetPath)
				assert.False(t, exists, "target path should be removed")
			}
		})
	}
}

func TestNodeServer_UnimplementedMethods(t *testing.T) {
	ns := &NodeServer{NodeId: "test-node"}

	t.Run("NodeStageVolume returns not implemented", func(t *testing.T) {
		_, err := ns.NodeStageVolume(context.Background(), &csi.NodeStageVolumeRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("NodeUnstageVolume returns not implemented", func(t *testing.T) {
		_, err := ns.NodeUnstageVolume(context.Background(), &csi.NodeUnstageVolumeRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("NodeGetVolumeStats returns not implemented", func(t *testing.T) {
		_, err := ns.NodeGetVolumeStats(context.Background(), &csi.NodeGetVolumeStatsRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("NodeExpandVolume returns not implemented", func(t *testing.T) {
		_, err := ns.NodeExpandVolume(context.Background(), &csi.NodeExpandVolumeRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})
}
