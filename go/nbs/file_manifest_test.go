// Copyright 2016 Attic Labs, Inc. All rights reserved.
// Licensed under the Apache License, version 2.0:
// http://www.apache.org/licenses/LICENSE-2.0

package nbs

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/attic-labs/noms/go/constants"
	"github.com/attic-labs/noms/go/hash"
	"github.com/attic-labs/testify/assert"
)

func makeFileManifestTempDir(t *testing.T) fileManifest {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	return fileManifest{dir}
}

func TestFileManifestParseIfExists(t *testing.T) {
	assert := assert.New(t)
	fm := makeFileManifestTempDir(t)
	defer os.RemoveAll(fm.dir)

	exists, vers, root, tableSpecs := fm.ParseIfExists(nil)
	assert.False(exists)

	// Simulate another process writing a manifest (with an old Noms version).
	newRoot := hash.FromData([]byte("new root"))
	tableName := hash.FromData([]byte("table1"))
	b, err := clobberManifest(fm.dir, strings.Join([]string{StorageVersion, "0", newRoot.String(), tableName.String(), "0"}, ":"))
	assert.NoError(err, string(b))

	// ParseIfExists should now reflect the manifest written above.
	exists, vers, root, tableSpecs = fm.ParseIfExists(nil)
	assert.True(exists)
	assert.Equal("0", vers)
	assert.Equal(newRoot, root)
	if assert.Len(tableSpecs, 1) {
		assert.Equal(tableName.String(), tableSpecs[0].name.String())
		assert.Equal(uint32(0), tableSpecs[0].chunkCount)
	}
}

func TestFileManifestParseIfExistsHoldsLock(t *testing.T) {
	assert := assert.New(t)
	fm := makeFileManifestTempDir(t)
	defer os.RemoveAll(fm.dir)

	// Simulate another process writing a manifest.
	newRoot := hash.FromData([]byte("new root"))
	tableName := hash.FromData([]byte("table1"))
	b, err := clobberManifest(fm.dir, strings.Join([]string{StorageVersion, constants.NomsVersion, newRoot.String(), tableName.String(), "0"}, ":"))
	assert.NoError(err, string(b))

	// ParseIfExists should now reflect the manifest written above.
	exists, vers, root, tableSpecs := fm.ParseIfExists(func() {
		// This should fail to get the lock, and therefore _not_ clobber the manifest.
		badRoot := hash.FromData([]byte("bad root"))
		b, err := tryClobberManifest(fm.dir, strings.Join([]string{StorageVersion, "0", badRoot.String(), tableName.String(), "0"}, ":"))
		assert.NoError(err, string(b))
	})

	assert.True(exists)
	assert.Equal(constants.NomsVersion, vers)
	assert.Equal(newRoot, root)
	if assert.Len(tableSpecs, 1) {
		assert.Equal(tableName.String(), tableSpecs[0].name.String())
		assert.Equal(uint32(0), tableSpecs[0].chunkCount)
	}
}

func TestFileManifestUpdate(t *testing.T) {
	assert := assert.New(t)
	fm := makeFileManifestTempDir(t)
	defer os.RemoveAll(fm.dir)

	// Simulate another process having already put old Noms data in dir/.
	b, err := clobberManifest(fm.dir, strings.Join([]string{StorageVersion, "0", hash.Hash{}.String()}, ":"))
	assert.NoError(err, string(b))

	assert.Panics(func() { fm.Update(nil, hash.Hash{}, hash.Hash{}, nil) })
}

func TestFileManifestUpdateWinRace(t *testing.T) {
	assert := assert.New(t)
	fm := makeFileManifestTempDir(t)
	defer os.RemoveAll(fm.dir)

	newRoot2 := hash.FromData([]byte("new root 2"))
	actual, tableSpecs := fm.Update(nil, hash.Hash{}, newRoot2, func() {
		// This should fail to get the lock, and therefore _not_ clobber the manifest. So the Update should succeed.
		newRoot := hash.FromData([]byte("new root"))
		b, err := tryClobberManifest(fm.dir, strings.Join([]string{StorageVersion, constants.NomsVersion, newRoot.String()}, ":"))
		assert.NoError(err, string(b))
	})
	assert.Equal(newRoot2, actual)
	assert.Nil(tableSpecs)
}

func TestFileManifestUpdateRootOptimisticLockFail(t *testing.T) {
	assert := assert.New(t)
	fm := makeFileManifestTempDir(t)
	defer os.RemoveAll(fm.dir)

	tableName := hash.FromData([]byte("table1"))
	newRoot := hash.FromData([]byte("new root"))
	b, err := tryClobberManifest(fm.dir, strings.Join([]string{StorageVersion, constants.NomsVersion, newRoot.String(), tableName.String(), "3"}, ":"))
	assert.NoError(err, string(b))

	newRoot2 := hash.FromData([]byte("new root 2"))
	actual, tableSpecs := fm.Update(nil, hash.Hash{}, newRoot2, nil)
	assert.Equal(newRoot, actual)
	if assert.Len(tableSpecs, 1) {
		assert.Equal(tableName.String(), tableSpecs[0].name.String())
		assert.Equal(uint32(3), tableSpecs[0].chunkCount)
	}
	actual, tableSpecs = fm.Update(nil, actual, newRoot2, nil)
}

// tryClobberManifest simulates another process trying to access dir/manifestFileName concurrently. To avoid deadlock, it does a non-blocking lock of dir/lockFileName. If it can get the lock, it clobbers the manifest.
func tryClobberManifest(dir, contents string) ([]byte, error) {
	return runClobber(dir, contents, false)
}

// clobberManifest simulates another process writing dir/manifestFileName concurrently. It takes the lock, so it's up to the caller to avoid deadlock.
func clobberManifest(dir, contents string) ([]byte, error) {
	return runClobber(dir, contents, true)
}

func runClobber(dir, contents string, takeLock bool) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1)
	clobber := filepath.Join(filepath.Dir(filename), "test/manifest_clobber.go")
	mkPath := func(f string) string {
		return filepath.Join(dir, f)
	}
	args := []string{"run", clobber}
	if takeLock {
		args = append(args, "--take-lock")
	}
	args = append(args, mkPath(lockFileName), mkPath(manifestFileName), contents)

	c := exec.Command("go", args...)
	return c.CombinedOutput()
}
