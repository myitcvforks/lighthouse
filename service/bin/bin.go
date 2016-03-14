package bin

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nwidger/lighthouse/service"
)

type Service struct {
	ProjectID int
	Service   *service.Service
}

type Bin struct {
	Default      bool       `json:"default"`
	ID           int        `json:"id"`
	Name         string     `json:"name"`
	Position     int        `json:"position"`
	ProjectID    int        `json:"project_id"`
	Query        string     `json:"query"`
	Shared       bool       `json:"shared"`
	TicketsCount int        `json:"tickets_count"`
	UpdatedAt    *time.Time `json:"updated_at"`
	UserID       int        `json:"user_id"`
	Global       bool       `json:"global"`
}

type Bins []*Bin

type BinCreate struct {
	Default bool   `json:"default"`
	Name    string `json:"name"`
	Query   string `json:"query"`
}

type BinUpdate struct {
	Default bool   `json:"default"`
	Name    string `json:"name"`
	Query   string `json:"query"`
}

type binRequest struct {
	Bin interface{} `json:"ticket_bin"`
}

func (br *binRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(br)
}

type binResponse struct {
	Bin *Bin `json:"ticket_bin"`
}

func (tr *binResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(tr)
}

type binsResponse struct {
	Bins []*binResponse `json:"ticket_bins"`
}

func (bsr *binsResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(bsr)
}

func (bsr *binsResponse) bins() Bins {
	bs := make(Bins, 0, len(bsr.Bins))
	for _, b := range bsr.Bins {
		bs = append(bs, b.Bin)
	}

	return bs
}

func (s *Service) basePath() string {
	return "/projects/" + strconv.Itoa(s.ProjectID)
}

func (s *Service) List() (Bins, error) {
	resp, err := s.Service.RoundTrip("GET", s.basePath()+"/bins.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	bsresp := &binsResponse{}
	err = bsresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return bsresp.bins(), nil
}

func (s *Service) Get(id int) (*Bin, error) {
	resp, err := s.Service.RoundTrip("GET", s.basePath()+"/bins/"+strconv.Itoa(id)+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	bresp := &binResponse{}
	err = bresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return bresp.Bin, nil
}

// Only the fields in BinCreate can be set.
func (s *Service) Create(b *Bin) (*Bin, error) {
	breq := &binRequest{
		Bin: &BinCreate{
			Default: b.Default,
			Name:    b.Name,
			Query:   b.Query,
		},
	}

	buf := &bytes.Buffer{}
	err := breq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.Service.RoundTrip("POST", s.basePath()+"/bins.json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	bresp := &binResponse{
		Bin: b,
	}
	err = bresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Only the fields in BinUpdate can be set.
func (s *Service) Update(b *Bin) error {
	breq := &binRequest{
		Bin: &BinUpdate{
			Default: b.Default,
			Name:    b.Name,
			Query:   b.Query,
		},
	}

	buf := &bytes.Buffer{}
	err := breq.Encode(buf)
	if err != nil {
		return err
	}

	resp, err := s.Service.RoundTrip("PUT", s.basePath()+"/bins/"+strconv.Itoa(b.ID)+".json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(id int) error {
	resp, err := s.Service.RoundTrip("DELETE", s.basePath()+"/bins/"+strconv.Itoa(id)+".json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}
