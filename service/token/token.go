package token

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/nwidger/lighthouse/service"
)

type Service struct {
	Service *service.Service
}

type Token struct {
	CreatedAt *time.Time `json:"created_at"`
	Note      string     `json:"note"`
	ProjectID int        `json:"project_id"`
	ReadOnly  bool       `json:"read_only"`
	Token     string     `json:"token"`
	UserID    int        `json:"user_id"`
}

type tokenResponse struct {
	Token *Token `json:"token"`
}

func (pr *tokenResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(pr)
}

func (pr *tokenResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(pr)
}

func (s *Service) Get(tokenStr string) (*Token, error) {
	resp, err := s.Service.RoundTrip("GET", "/tokens/"+tokenStr+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	tresp := &tokenResponse{}
	err = tresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return tresp.Token, nil
}
