# Write Go Tests

Write idiomatic Go tests following standard library patterns with table-driven test organization.

## Core Principles

1. **Self-documenting tests** - Test names and structure should explain what's being tested
2. **Table-driven when beneficial** - Use tables for tests with similar structure but different inputs
3. **Minimal comments** - Only document non-obvious behavior or test setup
4. **Clear error messages** - Use "got X, want Y" format in assertions
5. **Package comments only** - No doc comments on individual test functions

## Package-Level Comments

Every `_test.go` file should have a minimal package comment describing what aspects are tested:

```go
// Package driver tests the CSI driver implementation.
package driver

// Node service volume operations tests.
package driver

// Identity service capability tests.
package driver
```

**Format:**
- Simple one-line comment explaining test scope
- No elaborate documentation - tests are code, not prose
- Located before `package` declaration

## When to Use Table-Driven Tests

Use table-driven tests when you have:

### ✅ Good Candidates

1. **Multiple similar test cases** - Same test logic, different inputs/outputs
   ```go
   // Good: Testing same function with different inputs
   TestParseQuantity - (1Gi, 500Mi, 1G, invalid format)
   TestNodePublishVolume - (success case, invalid size, mount failure)
   ```

2. **Success and error paths** - Testing both happy path and error conditions
   ```go
   // Good: Same operation, different outcomes
   TestRunDriver - (valid NODE_ID, missing NODE_ID)
   TestNodeUnpublishVolume - (success, umount failure)
   ```

3. **Variations of behavior** - Same function, different configurations
   ```go
   // Good: Different volume sizes
   TestCreateBackingFile - (1Gi, 10Mi, 100Gi)
   ```

### ❌ Poor Candidates

1. **Unique test logic** - Each test validates completely different behavior
2. **Complex setup** - Test setup is more complex than the test itself
3. **Few cases** - Only 1-2 test cases (just write separate functions)
4. **Different assertions** - Each case needs fundamentally different validations

## Table-Driven Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name       string           // descriptive test case name
        input      InputType        // test inputs
        wantOutput ExpectedType     // expected outputs
        wantErr    bool            // whether error is expected
        errMsg     string          // expected error message (optional)
    }{
        {
            name:    "descriptive case name",
            input:   someInput,
            wantOutput: expectedOutput,
            wantErr: false,
        },
        {
            name:    "error case description",
            input:   invalidInput,
            wantErr: true,
            errMsg:  "expected error substring",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.wantOutput, result)
            }
        })
    }
}
```

## Field Naming Conventions

Use clear, descriptive field names in test structs:

### Input Fields
- Use descriptive names: `volumeID`, `targetPath`, `size`, `nodeID`
- Not generic: `input`, `data`, `param`

### Expected Output Fields
- Prefix with `want`: `wantErr`, `wantResult`, `wantBackingFile`, `wantMounted`
- Or use descriptive names: `expectedPath`, `shouldExist`

### Setup/Configuration Fields
- Use function closures: `setupFS func() afero.Fs`
- Or descriptive names: `createTestVolume`, `mockNodeID`

**Good examples:**
```go
tests := []struct {
    name         string
    volumeID     string
    size         string
    wantErr      bool
    wantFileSize int64
}{
    // ...
}
```

**Avoid:**
```go
tests := []struct {
    name   string
    input  interface{}  // too generic
    output interface{}  // too generic
    err    bool
}{
    // ...
}
```

## Test Case Names

Use descriptive names that explain the scenario:

**Good:**
- `"successfully creates 1Gi volume"`
- `"fails on invalid size format"`
- `"mounts volume with loop device"`
- `"handles missing backing file"`
- `"unmounts and cleans up volume"`

**Avoid:**
- `"test 1"`, `"case 2"` - not descriptive
- `"works"`, `"fails"` - too vague
- `"TestCase1"` - redundant prefix

## Filesystem Testing Patterns

When testing file operations with `conf.FS`:

### ✅ Do This

1. **Use `conf.FS` directly** - It's already initialized in `conf/testing.go`
2. **Clean up with defer** - Remove files you create, keep folder structure
3. **Test success paths** - Focus on what should work
4. **No filesystem swapping** - Don't replace `conf.FS` with new instances

**Good example:**
```go
func TestCreateBackingFile(t *testing.T) {
    backingFile := "/var/lib/csi-loop/test-volume.img"
    defer conf.FS.Remove(backingFile)

    err := createBackingFile(backingFile, 1024*1024*1024) // 1Gi
    require.NoError(t, err)

    exists, err := afero.Exists(conf.FS, backingFile)
    require.NoError(t, err)
    assert.True(t, exists)

    stat, err := conf.FS.Stat(backingFile)
    require.NoError(t, err)
    assert.Equal(t, int64(1024*1024*1024), stat.Size())
}
```

### ❌ Don't Do This

1. **Don't test infrastructure errors** - No read-only filesystem tests, no missing directory tests
2. **Don't swap filesystems** - Avoid `conf.FS = newFS` pattern
3. **Don't create new filesystems** - Use existing `conf.FS` instead of `afero.NewMemMapFs()`
4. **Don't test error cases that can't happen** - If `conf.FS` has proper setup, don't test missing directories

**Bad example:**
```go
func TestCreateFile_ReadOnlyError(t *testing.T) {
    // Don't test infrastructure failures
    fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
    originalFS := conf.FS
    conf.FS = fs
    defer func() { conf.FS = originalFS }()

    err := CreateFile("/path/file")
    assert.Error(t, err)  // Avoid this pattern
}
```

## Common Patterns

### Success Cases with Cleanup

```go
func TestNodePublishVolume(t *testing.T) {
    volumeID := "test-vol-123"
    backingFile := fmt.Sprintf("%s/%s.img", conf.BackingFileDir, volumeID)
    defer conf.FS.Remove(backingFile)

    // Test volume creation
    err := publishVolume(volumeID, "1Gi", "/mnt/test")
    require.NoError(t, err)

    // Verify backing file exists
    exists, err := afero.Exists(conf.FS, backingFile)
    require.NoError(t, err)
    assert.True(t, exists)
}
```

### Multiple Variations

```go
func TestParseVolumeSize(t *testing.T) {
    tests := []struct {
        name      string
        size      string
        wantBytes int64
        wantErr   bool
    }{
        {
            name:      "parses 1Gi correctly",
            size:      "1Gi",
            wantBytes: 1073741824,
            wantErr:   false,
        },
        {
            name:      "parses 500Mi correctly",
            size:      "500Mi",
            wantBytes: 524288000,
            wantErr:   false,
        },
        {
            name:    "fails on invalid format",
            size:    "invalid",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            bytes, err := parseVolumeSize(tt.size)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.wantBytes, bytes)
            }
        })
    }
}
```

## When NOT to Use Table-Driven Tests

Don't force table-driven tests when:

1. **Single test case** - Just write a normal test function
2. **Completely different test logic** - Each test validates different aspects
3. **Complex per-case setup** - Table becomes harder to read than separate functions
4. **Integration tests** - Often have unique setup/teardown per test

**Example of when to keep separate:**
```go
// Good: Each test validates fundamentally different behavior
func TestNodeServer_PublishesVolume(t *testing.T) { /* ... */ }
func TestNodeServer_UnpublishesVolume(t *testing.T) { /* ... */ }
func TestNodeServer_ReturnsCapabilities(t *testing.T) { /* ... */ }
```

## Test Helpers

Document helper functions that are used across multiple tests:

```go
// createTestVolumeRequest returns a basic NodePublishVolumeRequest for testing.
func createTestVolumeRequest(volumeID, size string) *csi.NodePublishVolumeRequest {
    return &csi.NodePublishVolumeRequest{
        VolumeId:   volumeID,
        TargetPath: "/mnt/test",
        VolumeContext: map[string]string{
            "size": size,
        },
    }
}
```

## CSI Driver Testing Notes

For CSI driver tests specifically:

1. **Mock system commands** - Don't actually run `mount`, `umount`, `mkfs.btrfs` in tests
2. **Use interfaces** - Consider wrapping exec.Command for testability
3. **Test gRPC handlers** - Test NodeServer methods directly, not via gRPC calls
4. **Focus on business logic** - Test path construction, size parsing, error handling

## Validation Checklist

- [ ] Package comment describes what the test file validates
- [ ] No doc comments on individual test functions (test names are self-documenting)
- [ ] Table-driven tests use descriptive field names (`wantErr`, not `err`)
- [ ] Test case names are descriptive (`"handles empty input"`, not `"test1"`)
- [ ] Error cases check error message content with `assert.Contains`
- [ ] Related test cases are grouped in table-driven tests
- [ ] Tests use `t.Run` for subtests to improve output
- [ ] Test logic is simple and easy to understand
- [ ] Helper functions have doc comments explaining their purpose
- [ ] Filesystem tests use `conf.FS` directly (no swapping, no infrastructure errors)
- [ ] File operations include `defer` cleanup for created files

## References

- [Go Wiki: Test Comments](https://go.dev/wiki/TestComments)
- [Go Wiki: Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [testing package](https://pkg.go.dev/testing)
