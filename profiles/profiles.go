// Package profiles provides access to profiles via the Lighthouse
// API.  http://help.lighthouseapp.com/kb/api/users-and-membership.
package profiles

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/nwidger/lighthouse"
)

type Service struct {
	basePath string
	s        *lighthouse.Service
}

func NewService(s *lighthouse.Service) *Service {
	return &Service{
		basePath: s.BasePath + "/profile",
		s:        s,
	}
}

type User struct {
	ID      int    `json:"id"`
	Job     string `json:"job"`
	Name    string `json:"name"`
	Website string `json:"website"`
}

type userResponse struct {
	User *User `json:"user"`
}

func (ur *userResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(ur)
}
func (s *Service) Get() (*User, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	uresp := &userResponse{}
	err = uresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return uresp.User, nil
}
