package update

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPlan_Apply(t *testing.T) {
	// Local==Baseline, Upstream≠Baseline → Apply upstream changes
	upstream := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("upstream content")},
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")},
	}
	local := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")}, // Same as baseline
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check upstream changes
	assert.Len(t, plan.UpstreamChanges, 1)
	assert.Equal(t, "file.txt", plan.UpstreamChanges[0].Path)
	assert.Equal(t, Action("modified"), plan.UpstreamChanges[0].Action)
	assert.Contains(t, plan.UpstreamChanges[0].Reason, "upstream")

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActApply, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "fast-forward")
}

func TestBuildPlan_PreserveLocal(t *testing.T) {
	// Upstream==Baseline, Local≠Baseline → Preserve local changes
	upstream := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")}, // Same as baseline
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")},
	}
	local := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("local modification")},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check local changes
	assert.Len(t, plan.LocalChanges, 1)
	assert.Equal(t, "file.txt", plan.LocalChanges[0].Path)
	assert.Equal(t, Action("modified"), plan.LocalChanges[0].Action)
	assert.Contains(t, plan.LocalChanges[0].Reason, "local")

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActPreserveLocal, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "preserve local")
}

func TestBuildPlan_Conflict(t *testing.T) {
	// Upstream≠Baseline AND Local≠Baseline → Conflict
	upstream := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("upstream modification")},
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")},
	}
	local := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("local modification")},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check both upstream and local changes
	assert.Len(t, plan.UpstreamChanges, 1)
	assert.Len(t, plan.LocalChanges, 1)

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActConflict, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "both upstream and local")
}

func TestBuildPlan_Delete(t *testing.T) {
	// Upstream removes file vs Baseline, Local==Baseline → Delete
	upstream := fstest.MapFS{
		// File removed in upstream
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")},
	}
	local := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")}, // Same as baseline
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check upstream changes (deletion)
	assert.Len(t, plan.UpstreamChanges, 1)
	assert.Equal(t, "file.txt", plan.UpstreamChanges[0].Path)
	assert.Equal(t, Action("deleted"), plan.UpstreamChanges[0].Action)

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActDelete, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "upstream deleted")
}

func TestBuildPlan_Keep(t *testing.T) {
	// No changes (Upstream==Baseline==Local) → Keep (should not appear in merge)
	content := []byte("unchanged content")
	upstream := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: content},
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: content},
	}
	local := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: content},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// No changes expected in any section
	assert.Empty(t, plan.UpstreamChanges)
	assert.Empty(t, plan.LocalChanges)
	assert.Empty(t, plan.Merge) // Keep actions are not recorded
}

func TestBuildPlan_NewFromUpstream(t *testing.T) {
	// File added in upstream only → Apply
	upstream := fstest.MapFS{
		"new.txt": &fstest.MapFile{Data: []byte("new upstream file")},
	}
	baseline := fstest.MapFS{
		// File doesn't exist in baseline
	}
	local := fstest.MapFS{
		// File doesn't exist in local
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check upstream changes
	assert.Len(t, plan.UpstreamChanges, 1)
	assert.Equal(t, "new.txt", plan.UpstreamChanges[0].Path)
	assert.Equal(t, Action("new"), plan.UpstreamChanges[0].Action)

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "new.txt", plan.Merge[0].Path)
	assert.Equal(t, ActApply, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "new file from upstream")
}

func TestBuildPlan_UserCreatedFile(t *testing.T) {
	// File exists only in local (user created) → Preserve local
	upstream := fstest.MapFS{
		// File doesn't exist in upstream
	}
	baseline := fstest.MapFS{
		// File doesn't exist in baseline
	}
	local := fstest.MapFS{
		"user.txt": &fstest.MapFile{Data: []byte("user created")},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check local changes
	assert.Len(t, plan.LocalChanges, 1)
	assert.Equal(t, "user.txt", plan.LocalChanges[0].Path)
	assert.Equal(t, Action("new"), plan.LocalChanges[0].Action)

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "user.txt", plan.Merge[0].Path)
	assert.Equal(t, ActPreserveLocal, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "user-created")
}

func TestBuildPlan_ConflictDeletedLocal(t *testing.T) {
	// Local deleted but upstream modified → Conflict
	upstream := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("upstream modified")},
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")},
	}
	local := fstest.MapFS{
		// File deleted in local
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check changes
	assert.Len(t, plan.UpstreamChanges, 1)
	assert.Len(t, plan.LocalChanges, 1)

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActConflict, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "local deleted but upstream modified")
}

func TestBuildPlan_ConflictDeletedUpstream(t *testing.T) {
	// Upstream deleted but local modified → Conflict
	upstream := fstest.MapFS{
		// File deleted in upstream
	}
	baseline := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("baseline content")},
	}
	local := fstest.MapFS{
		"file.txt": &fstest.MapFile{Data: []byte("local modified")},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Check merge decision
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActConflict, plan.Merge[0].Action)
	assert.Contains(t, plan.Merge[0].Reason, "upstream deleted but local modified")
}

func TestBuildPlan_MultipleFiles(t *testing.T) {
	// Test with multiple files showing different merge scenarios
	upstream := fstest.MapFS{
		"apply.txt":    &fstest.MapFile{Data: []byte("upstream changed")},
		"preserve.txt": &fstest.MapFile{Data: []byte("baseline")}, // Same as baseline → preserve local changes
		"conflict.txt": &fstest.MapFile{Data: []byte("upstream changed")},
		"new.txt":      &fstest.MapFile{Data: []byte("new from upstream")},
		"keep.txt":     &fstest.MapFile{Data: []byte("unchanged")},
		// deleted.txt removed from upstream
	}
	baseline := fstest.MapFS{
		"apply.txt":    &fstest.MapFile{Data: []byte("baseline")},
		"preserve.txt": &fstest.MapFile{Data: []byte("baseline")},
		"conflict.txt": &fstest.MapFile{Data: []byte("baseline")},
		"deleted.txt":  &fstest.MapFile{Data: []byte("to delete")},
		"keep.txt":     &fstest.MapFile{Data: []byte("unchanged")},
	}
	local := fstest.MapFS{
		"apply.txt":    &fstest.MapFile{Data: []byte("baseline")}, // Same as baseline → apply
		"preserve.txt": &fstest.MapFile{Data: []byte("local changed")}, // Local changed, upstream unchanged → preserve
		"conflict.txt": &fstest.MapFile{Data: []byte("local changed")}, // Both changed → conflict
		"deleted.txt":  &fstest.MapFile{Data: []byte("to delete")}, // Same as baseline → delete
		"user.txt":     &fstest.MapFile{Data: []byte("user created")}, // User created
		"keep.txt":     &fstest.MapFile{Data: []byte("unchanged")},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Create maps for easier testing
	mergeActions := make(map[string]Action)
	for _, entry := range plan.Merge {
		mergeActions[entry.Path] = entry.Action
	}

	// Verify merge decisions
	assert.Equal(t, ActApply, mergeActions["apply.txt"])
	assert.Equal(t, ActPreserveLocal, mergeActions["preserve.txt"])
	assert.Equal(t, ActConflict, mergeActions["conflict.txt"])
	assert.Equal(t, ActDelete, mergeActions["deleted.txt"])
	assert.Equal(t, ActApply, mergeActions["new.txt"])
	assert.Equal(t, ActPreserveLocal, mergeActions["user.txt"])

	// keep.txt should not appear in merge (ActKeep not recorded)
	_, exists := mergeActions["keep.txt"]
	assert.False(t, exists)
}

func TestBuildPlan_NestedFiles(t *testing.T) {
	// Test with files in subdirectories
	upstream := fstest.MapFS{
		"root.txt":           &fstest.MapFile{Data: []byte("root upstream")},
		"subdir/nested.txt":  &fstest.MapFile{Data: []byte("nested upstream")},
	}
	baseline := fstest.MapFS{
		"root.txt":           &fstest.MapFile{Data: []byte("root baseline")},
		"subdir/nested.txt":  &fstest.MapFile{Data: []byte("nested baseline")},
	}
	local := fstest.MapFS{
		"root.txt":           &fstest.MapFile{Data: []byte("root baseline")}, // Apply upstream
		"subdir/nested.txt":  &fstest.MapFile{Data: []byte("nested local")}, // Conflict
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	// Create map for easier testing
	mergeActions := make(map[string]Action)
	for _, entry := range plan.Merge {
		mergeActions[entry.Path] = entry.Action
	}

	assert.Equal(t, ActApply, mergeActions["root.txt"])
	assert.Equal(t, ActConflict, mergeActions["subdir/nested.txt"])
}

func TestBuildPlan_WithSubRoots(t *testing.T) {
	// Test with specific root directories within filesystems
	upstream := fstest.MapFS{
		"workspace/file.txt": &fstest.MapFile{Data: []byte("upstream")},
		"other/ignore.txt":   &fstest.MapFile{Data: []byte("ignored")},
	}
	baseline := fstest.MapFS{
		"workspace/file.txt": &fstest.MapFile{Data: []byte("baseline")},
		"other/ignore.txt":   &fstest.MapFile{Data: []byte("ignored")},
	}
	local := fstest.MapFS{
		"workspace/file.txt": &fstest.MapFile{Data: []byte("baseline")},
		"other/ignore.txt":   &fstest.MapFile{Data: []byte("ignored")},
	}

	plan, err := BuildPlan(upstream, "workspace", baseline, "workspace", local, "workspace")
	require.NoError(t, err)

	// Should only see changes in workspace, not other
	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "file.txt", plan.Merge[0].Path)
	assert.Equal(t, ActApply, plan.Merge[0].Action)
}

func TestBuildPlan_EmptyFilesystems(t *testing.T) {
	// Test with all empty filesystems
	upstream := fstest.MapFS{}
	baseline := fstest.MapFS{}
	local := fstest.MapFS{}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	assert.Empty(t, plan.UpstreamChanges)
	assert.Empty(t, plan.LocalChanges)
	assert.Empty(t, plan.Merge)
}

func TestBuildPlan_BinaryFiles(t *testing.T) {
	// Test with binary content to ensure hash comparison works
	upstreamBinary := []byte{0x00, 0x01, 0x02, 0xFF}
	baselineBinary := []byte{0x00, 0x01, 0x02, 0xFE}
	localBinary := []byte{0x00, 0x01, 0x02, 0xFE} // Same as baseline
	
	upstream := fstest.MapFS{
		"binary.bin": &fstest.MapFile{Data: upstreamBinary},
	}
	baseline := fstest.MapFS{
		"binary.bin": &fstest.MapFile{Data: baselineBinary},
	}
	local := fstest.MapFS{
		"binary.bin": &fstest.MapFile{Data: localBinary},
	}

	plan, err := BuildPlan(upstream, ".", baseline, ".", local, ".")
	require.NoError(t, err)

	assert.Len(t, plan.Merge, 1)
	assert.Equal(t, "binary.bin", plan.Merge[0].Path)
	assert.Equal(t, ActApply, plan.Merge[0].Action)
}