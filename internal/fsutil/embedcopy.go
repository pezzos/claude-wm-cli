package fsutil

import "embed"

// CopyTreeFromEmbed copies recursively the subtree srcRoot from an embed.FS to dst on disk.
func CopyTreeFromEmbed(src embed.FS, srcRoot, dst string) error {
	return copyTree(src, srcRoot, dst)
}
