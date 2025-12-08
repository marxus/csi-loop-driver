# Questions & TODOs for Later

## From node.go initial implementation

1. **mkfs flags for different filesystems**
   - Currently NOT using any flags (removed `-F`)
   - User reported it works fine without flags
   - Need to test on actual Linux system - might mkfs prompt for confirmation in non-interactive mode?
   - If we need flags: xfs uses `-f`, btrfs uses `-f`, ext4 uses `-F`, vfat doesn't need it

2. **Idempotency**
   - If Kubelet calls NodePublishVolume twice (e.g., after a crash), we'll try to create the file again
   - Should we check if backing file already exists?
   - Should we check if target is already mounted?
   - What's the correct behavior on retry?

3. **Error handling and cleanup**
   - If mkfs fails, we leave the backing file lying around
   - Should we cleanup on partial failure?
   - What about mount failures?
   - How to handle umount failures in NodeUnpublishVolume?

4. **General approach concerns**
   - Any issues with the naive implementation?
   - What should we prioritize adding first?
   - Performance concerns?
   - Security concerns?

## Understanding: Container tools vs Host kernel

### Can we use mount without nsenter?
**Yes!**
- `mount` is a system call to the kernel, not a container operation
- With `hostPID: true` and Bidirectional mount propagation, mounts are visible to host
- The `mount` binary just needs to exist (can use container's binary)

### What if host doesn't have mkfs tools?
**We install mkfs tools in the CONTAINER:**
- `mkfs.*` are userspace tools that write filesystem structures to files
- They only need basic file I/O, no special kernel support
- The HOST KERNEL must have the filesystem driver for mounting

**Example:**
```
Container has: mkfs.xfs (userspace tool)
Host kernel has: XFS driver (CONFIG_XFS_FS=y)
→ Works! Format with container's mkfs.xfs, mount with kernel's XFS driver
```

**What matters:**
- Host kernel must support the filesystem (kernel module/built-in)
- Container needs mkfs.* tools
- Container needs mount/umount binaries

**Potential issue:** If host kernel doesn't have filesystem support (e.g., no XFS module),
mount will fail even though mkfs succeeds. Should we detect this early or let it fail at mount time?

## Understanding: CSI Node Service Methods

### What is NodeGetCapabilities?
**Purpose:** Tells Kubernetes what optional features this CSI driver supports on the node.

**Our implementation:**
```go
return &csi.NodeGetCapabilitiesResponse{
    Capabilities: []*csi.NodeServiceCapability{},
}, nil
```
Returns empty list - we don't support any optional features.

**Available capabilities we could declare:**
- `STAGE_UNSTAGE_VOLUME` - Two-phase mount (stage first, then publish)
- `GET_VOLUME_STATS` - Report volume usage statistics
- `EXPAND_VOLUME` - Support resizing volumes
- `VOLUME_CONDITION` - Report volume health
- `SINGLE_NODE_MULTI_WRITER` - Multiple pods on same node can write

**For our simple driver:** We don't need any of these - just create → format → mount.

### What is NodeGetInfo?
**Purpose:** Returns information about the specific node where the driver is running.

**Our implementation:**
```go
return &csi.NodeGetInfoResponse{
    NodeId: ns.nodeID,
}, nil
```
Returns the node's ID (from `--nodeid` flag in DaemonSet).

**What else could we return:**
- `MaxVolumesPerNode` - Limit volumes per node
- `AccessibleTopology` - Zone/region info for scheduling

**Why Kubernetes needs these:**
- **NodeGetCapabilities** - Kubelet checks at startup to know available features
- **NodeGetInfo** - Kubelet uses NodeId to track which node has which volumes

**For our driver:** Both are required boilerplate. We return minimum: empty capabilities and just node ID.

### Do we really need unique node_id for ephemeral volumes?

**Question:** Since ephemeral volumes are node-local and never coordinated across nodes, do we need unique node_id values?

**Answer: YES, but not for the reason you might think**

**Why it's required:**

1. **Kubelet Registration (Critical)**
   - When the CSI driver starts, kubelet calls `NodeGetInfo` to register the driver
   - If `NodeGetInfo` is not implemented or returns invalid data, registration fails
   - Without registration, kubelet won't route any CSI calls to your driver

2. **CSINode Object Creation**
   - Kubernetes creates a `CSINode` object for each node to track available CSI drivers
   - The node_id is stored in this object for bookkeeping
   - For ephemeral volumes, this is mainly for tracking, not coordination

3. **NOT Used for Volume Coordination**
   - For ephemeral inline volumes, there's NO controller coordination
   - Each kubelet independently creates/deletes volumes on its own node
   - Volumes are never moved between nodes
   - The unique node_id doesn't affect actual volume operations

**Could you use a constant value?**
- Technically maybe, but DON'T
- The CSI spec says node_id MUST uniquely identify the node
- It might confuse Kubernetes internal tracking
- Provides zero benefit since `spec.nodeName` is already available in DaemonSet
- Could cause issues with CSINode object tracking

**Our implementation (correct):**
```yaml
# In DaemonSet:
env:
- name: NODE_ID
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName  # ✅ Unique per node
```

**Bottom line:** Yes, unique node_id per node is required, but it's for kubelet registration and CSINode object tracking, not for volume coordination (which doesn't exist for ephemeral volumes).

**Sources:**
- [Ephemeral Local Volumes Documentation](https://kubernetes-csi.github.io/docs/ephemeral-local-volumes.html)
- [CSI Specification](https://github.com/container-storage-interface/spec/blob/master/spec.md)
- [KEP-596: Ephemeral Inline CSI volumes](https://github.com/kubernetes/enhancements/blob/master/keps/sig-storage/596-csi-inline-volumes/README.md)

## Architecture Question: Do we need Controller service?

**For ephemeral inline volumes:**
- Identity service: **REQUIRED** (all CSI drivers need this)
- Node service: **REQUIRED** (where actual work happens)
- Controller service: **NOT NEEDED** (only for dynamic provisioning)

**Decision: REMOVED**
- Deleted `controller.go` entirely
- Removed `CONTROLLER_SERVICE` from `GetPluginCapabilities` in identity.go
- Removed controller registration from driver.go
- Simpler code, clearer intent (we're ephemeral-only)

## Architecture Question: Do we need TCP endpoint support?

**Question:** Does `parseEndpoint()` need to support both Unix sockets and TCP?

**Research findings:**
- **Kubelet only uses Unix sockets**: Kubelet discovers and communicates with CSI drivers exclusively via Unix Domain Sockets through the kubelet plugin registration mechanism
- **TCP is only for testing**: While the CSI spec technically allows TCP endpoints (like `dns:///my-machine:9000`), this is **only used for testing with csi-sanity** - not for production Kubernetes deployments
- **No real-world TCP usage**: No examples of production CSI drivers using TCP endpoints in Kubernetes exist
- **Performance**: Unix sockets are ~20% faster than TCP for local IPC (102,187 ns/op vs 127,188 ns/op for 100k requests)
- **Security**: Unix file permissions apply to Unix sockets, providing better access control

**Sources:**
- [Testing of CSI drivers | Kubernetes](https://kubernetes.io/blog/2020/01/08/testing-of-csi-drivers/)
- [Developing a CSI Driver for Kubernetes](https://kubernetes-csi.github.io/docs/developing.html)

**Decision: SIMPLIFIED**
- Removed `parseEndpoint()` function entirely
- Hard-coded to use "unix" scheme only
- Simply strip "unix://" prefix if present
- Since we're building for Kubernetes (not a test harness), TCP support adds unnecessary complexity with zero practical benefit

## Architecture Question: Do we need ServiceAccount and RBAC for ephemeral-only drivers?

**Question:** Does an ephemeral-only CSI driver DaemonSet need a ServiceAccount, ClusterRole, and ClusterRoleBinding?

**Short Answer: NO**

**What our driver does:**
- **Node-local operations only**: Creates backing files, formats them, mounts loop devices
- **No Kubernetes API calls**: Doesn't query nodes, create PVs, watch PVCs, manage secrets, etc.
- **Direct kubelet interaction**: Kubelet calls the driver via Unix socket at `/csi/csi.sock`
- **Gets node ID from environment**: Uses `spec.nodeName` injected as NODE_ID env var

**What requires RBAC permissions:**
- ❌ **Controller sidecars** - We don't have any (no external-provisioner, external-attacher, external-resizer, external-snapshotter)
- ❌ **Reading Secrets/ConfigMaps** - Our driver doesn't need these from Kubernetes API
- ❌ **Querying node info** - Our driver gets nodeID from environment variable, doesn't query Kubernetes API
- ❌ **Creating/updating PVs/PVCs** - Ephemeral volumes don't use PersistentVolume or PersistentVolumeClaim objects
- ❌ **Volume attachment/detachment** - Not needed for ephemeral inline volumes
- ❌ **CSINode updates** - Handled by node-driver-registrar sidecar (which also doesn't need RBAC)

**What the DaemonSet actually needs (security, not RBAC):**
- ✅ **Privileged mode** - To mount volumes and create loop devices
- ✅ **hostPID: true** - For mount visibility to host
- ✅ **Bidirectional mount propagation** - So mounts are visible to host kubelet
- ✅ **Access to host paths** - `/var/lib/kubelet/pods`, `/var/lib/csi-loop`, `/dev`
- ✅ **SYS_ADMIN capability** - For mount operations

**Pod ServiceAccount permissions for users:**
For CSI ephemeral inline volumes, anyone who can create a pod can mount volumes using a CSI driver configured for ephemeral use. No special RBAC permissions are required for the pod's ServiceAccount beyond what's needed to create the pod itself.

**Current implementation:**
We have a ClusterRole that grants `get` on `nodes` resources, but our driver never calls the Kubernetes API, so this permission is unused.

**Decision: CAN BE REMOVED**
- ServiceAccount can be removed (DaemonSet will use default ServiceAccount)
- ClusterRole can be removed (no Kubernetes API access needed)
- ClusterRoleBinding can be removed (no role to bind)

The driver works purely through:
1. Kubelet plugin registration mechanism (via socket and registration directory)
2. gRPC calls from kubelet to the driver
3. Local filesystem and mount operations

**Testing recommendation:**
Delete rbac.yaml and serviceaccount.yaml and test deployment. If the driver registers and mounts volumes successfully, no RBAC is needed.

**Sources:**
- [Ephemeral Local Volumes Documentation](https://kubernetes-csi.github.io/docs/ephemeral-local-volumes.html)
- [CSI Ephemeral Inline Volumes](https://kubernetes.io/blog/2020/01/21/csi-ephemeral-inline-volumes/)
- [KEP-596: Ephemeral Inline CSI volumes](https://github.com/kubernetes/enhancements/blob/master/keps/sig-storage/596-csi-inline-volumes/README.md)
- [Deploying a CSI Driver on Kubernetes](https://kubernetes-csi.github.io/docs/deploying.html)

## Security Question: Can we avoid privileged mode for the CSI driver?

**Question:** Does the CSI driver DaemonSet need to run with `privileged: true`, or can we use more restrictive capabilities?

**Short Answer: Depends on Kubernetes version**

### What Operations Require Privileges

Our CSI driver needs to perform the following privileged operations:
1. **Mount filesystems** - Using `mount -o loop` command
2. **Create loop devices** - Access to `/dev` for loop device creation
3. **Bidirectional mount propagation** - So mounts are visible to host kubelet and pods
4. **Format filesystems** - Using `mkfs.btrfs` (less privileged, mainly needs file I/O)

### The Traditional Requirement: privileged: true

**Before Kubernetes 1.27:**
- Bidirectional mount propagation required `privileged: true` containers
- This was a Kubernetes restriction, not a Linux kernel limitation
- `privileged: true` grants:
  - ❌ ALL Linux capabilities (CAP_SYS_ADMIN + many others)
  - ❌ Access to ALL host devices
  - ❌ Disables AppArmor, SELinux, seccomp profiles
  - ❌ Effectively full root access to the host

**Configuration (Kubernetes < 1.27):**
```yaml
securityContext:
  privileged: true
  # No need to add capabilities - privileged gives everything
```

### Modern Approach: CAP_SYS_ADMIN Only (Kubernetes 1.27+)

**Kubernetes PR #117812** (merged in 1.27) allows bidirectional mount propagation with just `CAP_SYS_ADMIN` capability, without requiring full privileged mode.

**What CAP_SYS_ADMIN provides:**
- ✅ Mount/unmount filesystems
- ✅ Create loop devices
- ✅ Bidirectional mount propagation (1.27+)
- ✅ Everything our CSI driver needs
- ✅ More restricted than `privileged: true`

**Configuration (Kubernetes >= 1.27):**
```yaml
securityContext:
  privileged: false
  allowPrivilegeEscalation: true  # Required for mount operations
  capabilities:
    add:
      - SYS_ADMIN
```

### Security Comparison

| Mode | Capabilities | Device Access | Security Profiles | Risk Level |
|------|--------------|---------------|-------------------|------------|
| `privileged: true` | ALL | ALL devices | Disabled | 🔴 Very High |
| `CAP_SYS_ADMIN` | SYS_ADMIN only | Limited | Active | 🟡 High |
| No privileges | None | None | Active | 🟢 Low |

### Important Security Note

⚠️ **CAP_SYS_ADMIN is still very powerful** - it's sometimes called "the new root" because:
- Grants numerous wide-ranging powers
- Can be leveraged to gain other capabilities
- Can potentially escalate to full root privileges
- Should only be used when absolutely necessary

However, it's **significantly more restrictive** than `privileged: true`.

### Alternative Approaches (Not Applicable to Loop Devices)

**For FUSE-based CSI drivers only:**
The meta-fuse-csi-plugin pattern allows:
- CSI driver pod runs with CAP_SYS_ADMIN (handles mount operations)
- User pods run without any special privileges
- CSI driver performs all privileged operations on behalf of users

**Not applicable to our driver because:**
- We use loop devices + traditional filesystems (btrfs), not FUSE
- Loop device creation requires CAP_SYS_ADMIN in the mounting process
- Mount operations happen during NodePublishVolume in the CSI driver

### Our Implementation Decision

**Current implementation:**
```yaml
securityContext:
  privileged: true
  capabilities:
    add:
      - SYS_ADMIN  # Redundant when privileged: true
```

**Recommended implementation:**

**Option 1: Support Kubernetes 1.27+ only (more secure)**
```yaml
securityContext:
  privileged: false
  allowPrivilegeEscalation: true
  capabilities:
    add:
      - SYS_ADMIN
```

**Option 2: Support older Kubernetes versions (broader compatibility)**
```yaml
securityContext:
  privileged: true
  # Remove redundant capabilities section
```

**Trade-off:**
- Option 1: Better security, requires Kubernetes 1.27+ (released April 2023)
- Option 2: Broader compatibility, less secure

### Verification

To test if your cluster supports CAP_SYS_ADMIN without privileged mode:
1. Check Kubernetes version: `kubectl version --short`
2. If >= 1.27, use Option 1 (CAP_SYS_ADMIN only)
3. If < 1.27, use Option 2 (privileged: true)

### What We Cannot Avoid

Regardless of approach, the CSI driver **must have CAP_SYS_ADMIN** because:
- Linux kernel requires it for mount/umount operations
- Loop device creation requires it
- No way to perform these operations without this capability
- This is a fundamental Linux security boundary, not a Kubernetes restriction

**Sources:**
- [Allow bidirectional mount propagation with SYS_ADMIN capability - PR #117812](https://github.com/kubernetes/kubernetes/pull/117812)
- [Limit privileged access for CSI driver - Issue #94400](https://github.com/kubernetes/kubernetes/issues/94400)
- [CAP_SYS_ADMIN: the new root](https://lwn.net/Articles/486306/)
- [Deploying a CSI Driver on Kubernetes](https://kubernetes-csi.github.io/docs/deploying.html)
- [meta-fuse-csi-plugin](https://github.com/pfnet-research/meta-fuse-csi-plugin)
- [Digital Ocean CSI Driver Issue #378](https://github.com/digitalocean/csi-digitalocean/issues/378)

## Notes
- Starting with simplest possible implementation
- Will address these as we test and iterate
