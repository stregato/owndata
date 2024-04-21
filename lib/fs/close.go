package fs

func (fs *FS) Close() {
	fs.stopUploadJob()
}
