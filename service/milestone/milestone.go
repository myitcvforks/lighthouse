package milestone

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nwidger/lighthouse/service"
)

type Service struct {
	ProjectID int
	Service   *service.Service
}

type Milestone struct {
	AttachmentsCount int        `json:"attachments_count"`
	CompletedAt      *time.Time `json:"completed_at"`
	CreatedAt        *time.Time `json:"created_at"`
	DueOn            *time.Time `json:"due_on"`
	Goals            string     `json:"goals"`
	GoalsHTML        string     `json:"goals_html"`
	ID               int        `json:"id"`
	MaxPoints        int        `json:"max_points"`
	OpenTicketsCount int        `json:"open_tickets_count"`
	Permalink        string     `json:"permalink"`
	PointsClosed     int        `json:"points_closed"`
	PointsOpen       int        `json:"points_open"`
	Position         int        `json:"position"`
	ProjectID        int        `json:"project_id"`
	TicketsCount     int        `json:"tickets_count"`
	Title            string     `json:"title"`
	UpdatedAt        *time.Time `json:"updated_at"`
	URL              string     `json:"url"`
	UserName         string     `json:"user_name"`
}

type Milestones []*Milestone

type MilestoneCreate struct {
	Goals string     `json:"goals"`
	Title string     `json:"title"`
	DueOn *time.Time `json:"due_on"`
}

type MilestoneUpdate struct {
	Goals string     `json:"goals"`
	Title string     `json:"title"`
	DueOn *time.Time `json:"due_on"`
}

type milestoneRequest struct {
	Milestone interface{} `json:"milestone"`
}

func (mr *milestoneRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(mr)
}

type milestoneResponse struct {
	Milestone *Milestone `json:"milestone"`
}

func (mr *milestoneResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(mr)
}

func (mr *milestoneResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(mr)
}

type milestonesResponse struct {
	Milestones []*milestoneResponse `json:"milestones"`
}

func (msr *milestonesResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(msr)
}

func (msr *milestonesResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(msr)
}

func (msr *milestonesResponse) milestones() Milestones {
	ms := make(Milestones, 0, len(msr.Milestones))
	for _, m := range msr.Milestones {
		ms = append(ms, m.Milestone)
	}

	return ms
}

func (s *Service) basePath() string {
	return "/projects/" + strconv.Itoa(s.ProjectID)
}

type ListOptions struct {
	Page int
}

func (s *Service) List(opts *ListOptions) (Milestones, error) {
	path := s.basePath() + "/milestones.json"
	if opts != nil {
		u, err := url.Parse(path)
		if err != nil {
			return nil, err
		}
		values := &url.Values{}
		if opts.Page > 0 {
			values.Set("page", strconv.Itoa(opts.Page))
		}
		u.RawQuery = values.Encode()
		path = u.RequestURI()
	}

	resp, err := s.Service.RoundTrip("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	msresp := &milestonesResponse{}
	err = msresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return msresp.milestones(), nil
}

func (s *Service) New() (*Milestone, error) {
	return s.get("new")
}

// Only the fields in MilestoneUpdate can be set.
func (s *Service) Update(m *Milestone) error {
	mreq := &milestoneRequest{
		Milestone: &MilestoneUpdate{
			Goals: m.Goals,
			Title: m.Title,
			DueOn: m.DueOn,
		},
	}

	buf := &bytes.Buffer{}
	err := mreq.Encode(buf)
	if err != nil {
		return err
	}

	resp, err := s.Service.RoundTrip("PUT", s.basePath()+"/milestones/"+strconv.Itoa(m.ID)+".json", buf)
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

func (s *Service) Get(id int) (*Milestone, error) {
	return s.get(strconv.Itoa(id))
}

func (s *Service) get(id string) (*Milestone, error) {
	resp, err := s.Service.RoundTrip("GET", s.basePath()+"/milestones/"+id+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	mresp := &milestoneResponse{}
	err = mresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return mresp.Milestone, nil
}

// Only the fields in MilestoneCreate can be set.
func (s *Service) Create(m *Milestone) (*Milestone, error) {
	mreq := &milestoneRequest{
		Milestone: &MilestoneCreate{
			Goals: m.Goals,
			Title: m.Title,
			DueOn: m.DueOn,
		},
	}

	buf := &bytes.Buffer{}
	err := mreq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.Service.RoundTrip("POST", s.basePath()+"/milestones.json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	mresp := &milestoneResponse{
		Milestone: m,
	}
	err = mresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Must use basic auth, API token not allowed for this action
func (s *Service) Close(id int) error {
	resp, err := s.Service.RoundTrip("PUT", s.basePath()+"/milestones/"+strconv.Itoa(id)+"/close.json", nil)
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

// Must use basic auth, API token not allowed for this action
func (s *Service) Open(id int) error {
	resp, err := s.Service.RoundTrip("PUT", s.basePath()+"/milestones/"+strconv.Itoa(id)+"/open.json", nil)
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
	resp, err := s.Service.RoundTrip("DELETE", s.basePath()+"/milestones/"+strconv.Itoa(id)+".json", nil)
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
