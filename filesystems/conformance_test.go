package filesystems_test

import (
	"github.com/stefanlester/skywalker/filesystems"
	"github.com/stefanlester/skywalker/filesystems/miniofilesystem"
	"github.com/stefanlester/skywalker/filesystems/s3filesystem"
	"github.com/stefanlester/skywalker/filesystems/sftpfilesystem"
	"github.com/stefanlester/skywalker/filesystems/webdavfilesystem"
)

// These compile-time assertions guard against the pointer-receiver regression
// that broke wiring once before: every backend implements filesystems.FS with
// pointer receivers, so only the pointer type satisfies the interface. If any
// backend's method set drifts, the package will fail to compile.
var (
	_ filesystems.FS = (*miniofilesystem.Minio)(nil)
	_ filesystems.FS = (*s3filesystem.S3)(nil)
	_ filesystems.FS = (*sftpfilesystem.SFTP)(nil)
	_ filesystems.FS = (*webdavfilesystem.WebDAV)(nil)
)
