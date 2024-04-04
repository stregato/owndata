package fs

type GetOptions struct {
	Async chan string
}

func (f *FS) GetData(path string, options GetOptions) ([]byte, error) {
	return nil, nil
}

func (f *FS) GetFile(path, dest string, options GetOptions) ([]byte, error) {
	return nil, nil
}
