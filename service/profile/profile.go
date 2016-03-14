package profile

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/nwidger/lighthouse/service"
)

type Service struct {
	Service *service.Service
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
	resp, err := s.Service.RoundTrip("GET", "/profile.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
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
