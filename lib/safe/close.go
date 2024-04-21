package safe

func (s *Safe) Close() error {
	s.Store.Close()
	return nil
}
