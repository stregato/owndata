package safe

func OpenFS(c *Safe) (*FS, error) {
	fs := &FS{S: c}
	return fs, nil
}
