package shepard

func (s *ShepardBot) setOrg(orgName string) error {
	org, _, err := s.gClient.Organizations.Get(s.ctx, orgName)
	if err != nil {
		return err
	}
	s.org = org
	return nil
}
