//go:build linux
// +build linux

package fs

import (
	"context"
	"os"
	"path"
	"syscall"

	"bazil.org/fuse"
	fsb "bazil.org/fuse/fs"
)

// Define a struct for the whole filesystem
type FuseFS struct {
	fs *FileSystem
}

// Root method that gets the root node for the filesystem
func (f *FuseFS) Root() (fsb.Node, error) {
	return &Dir{f.fs, ""}, nil
}

// Define a struct for a directory (the root directory in this case)
type Dir struct {
	f    *FileSystem
	name string
}

// Attr fills in the attributes for the directory
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0755
	return nil
}

// Lookup looks up a specific entry in the directory
func (d *Dir) Lookup(ctx context.Context, name string) (fsb.Node, error) {
	file, err := d.f.Stat(path.Join(d.name, name))
	if err == nil {
		return &FuseFile{
			file: file,
			f:    *d.f,
		}, nil
	}
	if os.IsNotExist(err) {
		return nil, fuse.Errno(syscall.ENOENT)
	}
	return nil, err
}

// ReadDirAll returns all the entries of the directory
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {

	ls, err := d.f.List(d.name, ListOptions{})
	if err != nil {
		return nil, err
	}

	var dirent []fuse.Dirent
	names := map[string]bool{}
	for _, l := range ls {
		if l.Name == "" {
			continue
		}
		if names[l.Name] {
			continue
		}
		var typ fuse.DirentType
		if l.IsDir {
			typ = fuse.DT_Dir
		} else {
			typ = fuse.DT_File
		}

		dirent = append(dirent, fuse.Dirent{
			Inode: l.ID.Uint64(),
			Name:  l.Name,
			Type:  typ,
		})
		names[l.Name] = true
	}

	return dirent, nil
}

// Implement Create method for creating files
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fsb.Node, fsb.Handle, error) {
	file, err := d.f.PutData(path.Join(d.name, req.Name), []byte{}, PutOptions{})
	if err != nil {
		return nil, nil, err
	}
	f := &FuseFile{
		file: file,
		f:    *d.f,
	}
	resp.Node = fuse.NodeID(file.ID)
	resp.Attr = fuse.Attr{
		Inode: file.ID.Uint64(),
		Mode:  req.Mode,
	}
	return f, f, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	err := d.f.Delete(path.Join(d.name, req.Name))
	if err != nil {
		return fuse.Errno(syscall.ENOENT)
	}
	return nil
}

func (d *Dir) Rmdir(ctx context.Context, req *fuse.RemoveRequest) error {
	return fuse.Errno(syscall.EPERM)
}

func (d *Dir) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fsb.Node) error {
	newDirPath := newDir.(*Dir).name
	oldPath := path.Join(d.name, req.OldName)
	newPath := path.Join(newDirPath, req.NewName)

	_, err := d.f.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil

}

// File represents a file in the filesystem
type FuseFile struct {
	file File
	f    FileSystem
}

// Attr sets the attributes for the file
func (f *FuseFile) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.file.ID.Uint64()
	a.Mode = 0666
	a.Size = uint64(f.file.Size)
	return nil
}

// ReadAll reads the data from the file
func (f *FuseFile) ReadAll(ctx context.Context) ([]byte, error) {
	name := path.Join(f.file.Dir, f.file.Name)
	return f.f.GetData(name, GetOptions{})
}

func (f *FuseFile) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	name := path.Join(f.file.Dir, f.file.Name)
	file, err := f.f.PutData(name, req.Data, PutOptions{ID: f.file.ID})
	if err != nil {
		return err
	}
	resp.Size = file.Size
	return nil
}

func (f *FuseFile) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	return nil
}
