package service

func (s *Service) GetGoogleOAuthURL() string {
	return s.google.AuthCodeURL()
}
