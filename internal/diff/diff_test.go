package diff

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffTrees_NewFiles(t *testing.T) {
	// File exists in A but not in B (new)
	fsA := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("content1")},
		"file2.txt": &fstest.MapFile{Data: []byte("content2")},
	}
	fsB := fstest.MapFS{
		"file2.txt": &fstest.MapFile{Data: []byte("content2")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "file1.txt", Type: ChangeNew},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_DeletedFiles(t *testing.T) {
	// File exists in B but not in A (deleted)
	fsA := fstest.MapFS{
		"file2.txt": &fstest.MapFile{Data: []byte("content2")},
	}
	fsB := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("content1")},
		"file2.txt": &fstest.MapFile{Data: []byte("content2")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "file1.txt", Type: ChangeDel},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_ModifiedFiles(t *testing.T) {
	// File exists in both but with different content
	fsA := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("modified content")},
		"file2.txt": &fstest.MapFile{Data: []byte("unchanged content")},
	}
	fsB := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("original content")},
		"file2.txt": &fstest.MapFile{Data: []byte("unchanged content")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "file1.txt", Type: ChangeMod},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_UnchangedFiles(t *testing.T) {
	// Files exist in both with same content - no changes expected
	fsA := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("same content")},
		"file2.txt": &fstest.MapFile{Data: []byte("also same")},
	}
	fsB := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("same content")},
		"file2.txt": &fstest.MapFile{Data: []byte("also same")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	assert.Empty(t, changes)
}

func TestDiffTrees_MixedChanges(t *testing.T) {
	// Multiple types of changes in one comparison
	fsA := fstest.MapFS{
		"new.txt":      &fstest.MapFile{Data: []byte("new file")},
		"modified.txt": &fstest.MapFile{Data: []byte("modified in A")},
		"unchanged.txt": &fstest.MapFile{Data: []byte("same content")},
	}
	fsB := fstest.MapFS{
		"deleted.txt":   &fstest.MapFile{Data: []byte("will be deleted")},
		"modified.txt":  &fstest.MapFile{Data: []byte("original in B")},
		"unchanged.txt": &fstest.MapFile{Data: []byte("same content")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "deleted.txt", Type: ChangeDel},
		{Path: "modified.txt", Type: ChangeMod},
		{Path: "new.txt", Type: ChangeNew},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_NestedFiles(t *testing.T) {
	// Test files in subdirectories
	fsA := fstest.MapFS{
		"root.txt":           &fstest.MapFile{Data: []byte("root file")},
		"subdir/nested.txt":  &fstest.MapFile{Data: []byte("nested file")},
		"subdir/new.txt":     &fstest.MapFile{Data: []byte("new nested")},
	}
	fsB := fstest.MapFS{
		"root.txt":           &fstest.MapFile{Data: []byte("root file")},
		"subdir/nested.txt":  &fstest.MapFile{Data: []byte("different nested")},
		"subdir/deleted.txt": &fstest.MapFile{Data: []byte("to delete")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "subdir/deleted.txt", Type: ChangeDel},
		{Path: "subdir/nested.txt", Type: ChangeMod},
		{Path: "subdir/new.txt", Type: ChangeNew},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_DirectoriesIgnored(t *testing.T) {
	// Directories should be ignored, only files compared
	fsA := fstest.MapFS{
		"file.txt":      &fstest.MapFile{Data: []byte("content")},
		"dir1/sub.txt":  &fstest.MapFile{Data: []byte("sub content")},
	}
	fsB := fstest.MapFS{
		"file.txt":      &fstest.MapFile{Data: []byte("content")},
		"dir2/sub.txt":  &fstest.MapFile{Data: []byte("sub content")},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "dir1/sub.txt", Type: ChangeNew},
		{Path: "dir2/sub.txt", Type: ChangeDel},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_WithSubRoots(t *testing.T) {
	// Test with specific root directories within filesystems
	fsA := fstest.MapFS{
		"root/file1.txt": &fstest.MapFile{Data: []byte("content1")},
		"root/file2.txt": &fstest.MapFile{Data: []byte("content2")},
		"other/ignore.txt": &fstest.MapFile{Data: []byte("ignored")},
	}
	fsB := fstest.MapFS{
		"root/file2.txt": &fstest.MapFile{Data: []byte("content2")},
		"root/file3.txt": &fstest.MapFile{Data: []byte("content3")},
		"other/ignore.txt": &fstest.MapFile{Data: []byte("ignored")},
	}

	changes, err := DiffTrees(fsA, "root", fsB, "root")
	require.NoError(t, err)

	expected := []Change{
		{Path: "file1.txt", Type: ChangeNew},
		{Path: "file3.txt", Type: ChangeDel},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_EmptyFilesystems(t *testing.T) {
	// Test with empty filesystems
	fsA := fstest.MapFS{}
	fsB := fstest.MapFS{}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	assert.Empty(t, changes)
}

func TestDiffTrees_EmptyVsPopulated(t *testing.T) {
	// Test empty vs populated filesystems
	fsA := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: []byte("content1")},
		"file2.txt": &fstest.MapFile{Data: []byte("content2")},
	}
	fsB := fstest.MapFS{}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "file1.txt", Type: ChangeNew},
		{Path: "file2.txt", Type: ChangeNew},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_SortingBehavior(t *testing.T) {
	// Test that results are properly sorted by path
	fsA := fstest.MapFS{
		"z_file.txt": &fstest.MapFile{Data: []byte("z content")},
		"a_file.txt": &fstest.MapFile{Data: []byte("a content")},
		"m_file.txt": &fstest.MapFile{Data: []byte("m content")},
	}
	fsB := fstest.MapFS{}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	// Verify alphabetical sorting
	assert.Equal(t, "a_file.txt", changes[0].Path)
	assert.Equal(t, "m_file.txt", changes[1].Path)
	assert.Equal(t, "z_file.txt", changes[2].Path)
}

func TestDiffTrees_IdenticalContent(t *testing.T) {
	// Test files with identical content but different names
	identicalContent := []byte("identical content")
	
	fsA := fstest.MapFS{
		"file1.txt": &fstest.MapFile{Data: identicalContent},
		"file2.txt": &fstest.MapFile{Data: identicalContent},
	}
	fsB := fstest.MapFS{
		"file2.txt": &fstest.MapFile{Data: identicalContent},
		"file3.txt": &fstest.MapFile{Data: identicalContent},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "file1.txt", Type: ChangeNew},
		{Path: "file3.txt", Type: ChangeDel},
	}

	assert.Equal(t, expected, changes)
}

func TestDiffTrees_BinaryContent(t *testing.T) {
	// Test with binary content to ensure hash comparison works
	binaryA := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	binaryB := []byte{0x00, 0x01, 0x02, 0xFF, 0xFD} // Last byte different
	
	fsA := fstest.MapFS{
		"binary.bin": &fstest.MapFile{Data: binaryA},
	}
	fsB := fstest.MapFS{
		"binary.bin": &fstest.MapFile{Data: binaryB},
	}

	changes, err := DiffTrees(fsA, ".", fsB, ".")
	require.NoError(t, err)

	expected := []Change{
		{Path: "binary.bin", Type: ChangeMod},
	}

	assert.Equal(t, expected, changes)
}