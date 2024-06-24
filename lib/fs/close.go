package fs

func (fs *FileSystem) Close() {
	fs.stopUploadJob()
}
