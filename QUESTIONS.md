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

## Notes
- Starting with simplest possible implementation
- Will address these as we test and iterate
