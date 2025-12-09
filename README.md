# csi-loop-driver

CSI driver that provides ephemeral node-local volumes backed by loop devices. Volumes are created from files using loop mount, formatted with a filesystem, and mounted to pods. The filesystem type is determined by what's available on the host node.

## How It Works

```
Pod requests volume → CSI driver receives NodePublishVolume
                    ↓
Creates backing file (truncate -s <size> /var/lib/csi-loop/<volume-id>.img)
                    ↓
Formats with filesystem (mkfs.btrfs <backing-file>)
                    ↓
Mounts as loop device (mount -o loop <backing-file> <target-path>)
                    ↓
Pod writes to /data → writes to loop-mounted volume
```

Volumes are ephemeral and deleted when the pod terminates (NodeUnpublishVolume).

**⚠️ Experimental / Prototype Project**

This is a minimal CSI driver demonstrating ephemeral inline volumes with loop devices. Currently hardcoded to btrfs for POC purposes, but designed to support any filesystem type available on the host (ext4, xfs, btrfs, etc.). Not intended for production use.

## Build

```bash
# Clone the repository
git clone https://github.com/marxus/csi-loop-driver.git
cd csi-loop-driver

# Build the Docker image
docker build -t ghcr.io/marxus/csi-loop-driver:latest .

# Push to registry
docker push ghcr.io/marxus/csi-loop-driver:latest
```

## Installation

```bash
# Install or upgrade in the kube-system namespace
helm upgrade --install csi-loop-driver ./charts/csi-loop-driver \
  --namespace kube-system \
  --create-namespace \
  --set image.repository=ghcr.io/marxus/csi-loop-driver \
  --set image.tag=latest
```

This installs:
- CSIDriver resource registering `loop.csi.k8s.io`
- DaemonSet running driver pods on each node
- Required RBAC and ServiceAccount

## Usage

Create a pod with an ephemeral inline CSI volume:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-loop-volume
spec:
  containers:
  - name: test
    image: alpine:3.22
    command: ["/bin/sh", "-c"]
    args:
      - |
        echo "Testing loop volume"
        df -h /data
        echo "Hello from loop volume!" > /data/test.txt
        cat /data/test.txt
        sleep infinity
    volumeMounts:
    - name: loop-volume
      mountPath: /data
  volumes:
  - name: loop-volume
    csi:
      driver: loop.csi.k8s.io
      volumeAttributes:
        size: 1Gi
```

Apply the pod:
```bash
kubectl apply -f test-pod.yaml
```

Volume attributes:
- `size` - Volume size in Kubernetes quantity format (1Gi, 500Mi, etc.)

## Project Status

### Implemented

- ✅ CSI Identity service (GetPluginInfo, GetPluginCapabilities, Probe)
- ✅ CSI Node service (NodePublishVolume, NodeUnpublishVolume, NodeGetInfo, NodeGetCapabilities)
- ✅ Ephemeral inline volume support
- ✅ Loop device mounting
- ✅ Filesystem formatting (btrfs hardcoded for POC)
- ✅ Kubernetes quantity parsing (1Gi, 500Mi)
- ✅ Environment-specific configuration (release, develop, testing)
- ✅ Mockable system commands for testing
- ✅ Comprehensive test coverage (9 tests)
- ✅ Helm chart deployment
- ✅ Multi-arch Docker build

### Future Exploration

- Configurable filesystem type (ext4, xfs, btrfs)
- Persistent volumes (PV/PVC)
- Volume staging/unstaging
- Volume expansion
- Volume statistics
- Snapshot support

## Local Development

### Prerequisites

- Go 1.24.3
- btrfs-progs (or other mkfs.* tools)
- Access to a Kubernetes cluster
- kubectl configured

### Run Locally

```bash
# Development build (uses sandboxed filesystem under ./tmp)
go run ./cmd/csi-loop-driver

# Release build (uses real filesystem)
go build -tags=release -o csi-loop-driver ./cmd/csi-loop-driver
NODE_ID=test-node ./csi-loop-driver
```

**Development mode:**
- Filesystem sandboxed to `./tmp`
- Creates `/csi/csi.sock` socket (within sandbox)
- NodeId defaults to "node-id"

**Release mode:**
- Uses real filesystem paths
- Reads `NODE_ID` from environment variable (required)
- Creates `/csi/csi.sock` Unix socket

## Environment Variables

**Release mode** (required):
- `NODE_ID` - Unique identifier for the node

**Development mode** (defaults):
- NodeId: "node-id"
- Socket: `/csi/csi.sock` (within `./tmp`)

## Package Structure

```
pkg/
├── driver/      - CSI Identity and Node service implementations
└── serve/       - High-level driver startup function

cmd/csi-loop-driver/ - Main entry point

conf/            - Environment-specific configuration
├── release.go   - Production config (real filesystem)
├── develop.go   - Development config (sandboxed filesystem)
└── testing.go   - Test config (in-memory filesystem)

charts/csi-loop-driver/ - Helm chart for deployment
```

## Development

**Testing:**
```bash
go test ./...
```

All tests: 9 tests across 1 package (pkg/driver)

**Build:**
```bash
# Development build
go build -o csi-loop-driver ./cmd/csi-loop-driver

# Release build (static binary)
go build -tags=release -o csi-loop-driver ./cmd/csi-loop-driver
```

**Docker:**
```bash
docker build -t csi-loop-driver:latest .
```

Multi-stage build with support for prebuilt binaries.

## License

MIT
