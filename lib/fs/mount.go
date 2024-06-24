package fs

import (
	"log"
	"os"

	"bazil.org/fuse"
	fsb "bazil.org/fuse/fs"
)

func (fs *FileSystem) Mount(mountPoint string) error {
	os.MkdirAll(mountPoint, 0755)

	fuse.Debug = func(msg interface{}) {
		log.Printf("FUSE Debug: %v", msg)
	}

	c, err := fuse.Mount(
		mountPoint,
		fuse.FSName("miofs"),
		fuse.Subtype(fs.S.ID),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	err = fsb.Serve(c, &FuseFS{fs: fs})
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileSystem) Unmount(mountPoint string) error {
	return fuse.Unmount(mountPoint)
}
