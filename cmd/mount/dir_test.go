//go:build linux

package mount

import (
	"context"
	"os"
	"testing"
	"time"

	"bazil.org/fuse"
	_ "github.com/rclone/rclone/backend/local"
	"github.com/rclone/rclone/cmd/mountlib"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDirEntryValid checks that entries made by Create are cached by the
// kernel for --attr-timeout, as those found by Lookup are, rather than for
// bazil's default of 1 minute.
func TestDirEntryValid(t *testing.T) {
	ctx := context.Background()
	f, err := fs.NewFs(ctx, t.TempDir())
	require.NoError(t, err)
	VFS := vfs.New(ctx, f, nil)
	defer VFS.Shutdown()
	opt := mountlib.Opt
	opt.AttrTimeout = fs.Duration(3 * time.Second)
	fsys := NewFS(VFS, &opt)
	root, err := fsys.Root()
	require.NoError(t, err)
	dir := root.(*Dir)

	var createResp fuse.CreateResponse
	_, handle, err := dir.Create(ctx, &fuse.CreateRequest{Name: "file", Flags: fuse.OpenFlags(os.O_CREATE | os.O_WRONLY), Mode: 0644}, &createResp)
	require.NoError(t, err)
	require.NoError(t, handle.(*FileHandle).Release(ctx, &fuse.ReleaseRequest{}))
	assert.Equal(t, 3*time.Second, createResp.EntryValid)

	var lookupResp fuse.LookupResponse
	_, err = dir.Lookup(ctx, &fuse.LookupRequest{Name: "file"}, &lookupResp)
	require.NoError(t, err)
	assert.Equal(t, 3*time.Second, lookupResp.EntryValid)
}
