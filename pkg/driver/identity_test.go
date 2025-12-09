// Identity service capability tests.
package driver

import (
	"context"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdentityServer_GetPluginInfo(t *testing.T) {
	ids := &IdentityServer{}

	resp, err := ids.GetPluginInfo(context.Background(), &csi.GetPluginInfoRequest{})

	require.NoError(t, err)
	assert.Equal(t, "loop.csi.k8s.io", resp.Name)
	assert.Equal(t, "v1.0.0", resp.VendorVersion)
}

func TestIdentityServer_GetPluginCapabilities(t *testing.T) {
	ids := &IdentityServer{}

	resp, err := ids.GetPluginCapabilities(context.Background(), &csi.GetPluginCapabilitiesRequest{})

	require.NoError(t, err)
	assert.Empty(t, resp.Capabilities, "ephemeral-only driver should have no plugin capabilities")
}

func TestIdentityServer_Probe(t *testing.T) {
	ids := &IdentityServer{}

	resp, err := ids.Probe(context.Background(), &csi.ProbeRequest{})

	require.NoError(t, err)
	assert.NotNil(t, resp)
}
