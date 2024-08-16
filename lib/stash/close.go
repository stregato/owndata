package stash

func (s *Stash) Close() error {
	s.Store.Close()
	return nil
}
