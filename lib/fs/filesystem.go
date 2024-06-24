package fs

import (
	"path"

	"github.com/stregato/mio/lib/safe"
)

var (
	FSDir            = "fs"
	HeadersDir       = path.Join(FSDir, "headers")
	DataDir          = path.Join(FSDir, "data")
	ConfigPath       = path.Join(FSDir, "config.conf")
	ErrExists        = "ErrExist: filesystem already exists in %s"
	DefaultGroupName = safe.GroupName("usr") // default group name

	MIO_GET_GROUP_NAME = "MIO_GET_GROUP_NAME" // query to get group name
)

type FileSystem struct {
	S *safe.Safe
}
