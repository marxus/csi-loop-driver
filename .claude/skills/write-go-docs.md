# Write Go Documentation

Generate Go documentation comments following official Go standards for godoc and pkg.go.dev.

## Core Principles

1. **Every exported (capitalized) name must have a doc comment**
2. Doc comments appear directly before declarations with **no blank lines**
3. First sentence is crucial - appears in search results and package listings
4. Use complete sentences starting with the declared name
5. Explain **what** the code does, not **how** it works

## Documentation by Type

### Package Comments

```go
// Package [name] provides [brief description].
package name
```

**Example for this project:**
```go
// Package driver implements the CSI Node, Identity, and Controller services
// for the loop volume driver. It provides ephemeral node-local volumes
// backed by XFS-formatted loop devices.
package driver
```

### Function/Method Comments

```go
// FunctionName [verb phrase describing what it does or returns].
func FunctionName(param Type) (ReturnType, error)
```

**Examples from CSI driver:**
```go
// NodePublishVolume mounts the volume to the target path.
// It creates a backing file, formats it with XFS, and mounts it as a loop device.
//
// Returns an error if volume creation, formatting, or mounting fails.
func (ns *NodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error)

// NewDriver creates a new CSI driver instance with the given node ID and endpoint.
func NewDriver(nodeID, endpoint string) (*Driver, error)
```

### Type Comments

```go
// TypeName [describes what each instance represents].
type TypeName struct
```

**Example:**
```go
// NodeServer implements the CSI Node service.
// It handles volume mounting and unmounting operations on the node.
type NodeServer struct {
    driver *Driver
    nodeID string
}
```

### Simple Pattern for This Project

Since this is a small, focused CSI driver, keep documentation simple:

**Do document:**
- What each exported function does
- Parameters that need explanation (volumeID, targetPath, etc.)
- Error conditions
- CSI-specific behavior

**Don't over-document:**
- Obvious getters/setters
- Standard CSI method signatures (everyone knows what NodePublishVolume is)
- Implementation details (loop device mechanics, mkfs flags)

## Quick Examples

```go
// Run starts the CSI driver and blocks until it exits.
func (d *Driver) Run() error

// parseEndpoint extracts the scheme and address from a CSI endpoint string.
// It handles "unix://path" and "tcp://host:port" formats.
func parseEndpoint(endpoint string) (string, string)
```

## Tools

```bash
# View package documentation
go doc github.com/marxus/csi-loop-driver/pkg/driver

# View specific symbol
go doc github.com/marxus/csi-loop-driver/pkg/driver.NodeServer

# Format code including comments
gofmt -w .
```

## Validation Checklist

- [ ] All exported names have doc comments
- [ ] Comments start with the declared name
- [ ] First sentence is complete
- [ ] No blank lines between comment and declaration
- [ ] Comments explain what, not how
- [ ] Error conditions documented

## References

- [Go Doc Comments](https://tip.golang.org/doc/comment)
- [CSI Spec](https://github.com/container-storage-interface/spec)
