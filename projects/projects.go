package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nwidger/lighthouse"
)

type Service struct {
	basePath string
	s        *lighthouse.Service
}

func NewService(s *lighthouse.Service) (*Service, error) {
	return &Service{
		basePath: s.BasePath + "/projects",
		s:        s,
	}, nil
}

type Todos struct {
	Projects   bool `json:"projects"`
	Tickets    bool `json:"tickets"`
	Milestones bool `json:"milestones"`
}

type CommaList []string

func (t *CommaList) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	s := ""
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("CommaList.UnmarshalJSON: %v: %v", data, err)
	}

	*t = strings.FieldsFunc(s, func(r rune) bool {
		return r == ','
	})

	return nil
}

func (t *CommaList) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.Join(*t, `,`) + `"`), nil
}

type User struct {
	ID        int    `json:"id"`
	Job       string `json:"job"`
	Name      string `json:"name"`
	Website   string `json:"website"`
	AvatarURL string `json:"avatar_url"`
}

type Membership struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	User    *User  `json:"user"`
	Account string `json:"account"`
}

type Memberships []*Membership

type membershipResponse struct {
	Membership *Membership `json:"membership"`
}

type membershipsResponse struct {
	Memberships []*membershipResponse `json:"memberships"`
}

func (psr *membershipsResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(psr)
}

func (psr *membershipsResponse) memberships() Memberships {
	ps := make(Memberships, 0, len(psr.Memberships))
	for _, p := range psr.Memberships {
		ps = append(ps, p.Membership)
	}

	return ps
}

type Project struct {
	Archived               bool       `json:"archived"`
	ClosedStates           string     `json:"closed_states"`
	CreatedAt              *time.Time `json:"created_at"`
	DefaultAssignedUserID  int        `json:"default_assigned_user_id"`
	DefaultMilestoneID     int        `json:"default_milestone_id"`
	DefaultTicketText      string     `json:"default_ticket_text"`
	Description            string     `json:"description"`
	DescriptionHTML        string     `json:"description_html"`
	EnablePoints           bool       `json:"enable_points"`
	Hidden                 bool       `json:"hidden"`
	ID                     int        `json:"id"`
	License                string     `json:"license"`
	Name                   string     `json:"name"`
	OpenStates             string     `json:"open_states"`
	OpenTicketsCount       int        `json:"open_tickets_count"`
	OssReadonly            bool       `json:"oss_readonly"`
	Permalink              string     `json:"permalink"`
	PointsScale            string     `json:"points_scale"`
	Public                 bool       `json:"public"`
	SendChangesetsToEvents bool       `json:"send_changesets_to_events"`
	TodosCompleted         Todos      `json:"todos_completed"`
	UpdatedAt              string     `json:"updated_at"`
	OpenStatesList         CommaList  `json:"open_states_list"`
	ClosedStatesList       CommaList  `json:"closed_states_list"`
}

type Projects []*Project

type ProjectCreate struct {
	Archived bool   `json:"archived"`
	Name     string `json:"name"`
	Public   bool   `json:"public"`
}

type ProjectUpdate struct {
	Archived bool   `json:"archived"`
	Name     string `json:"name"`
	Public   bool   `json:"public"`
}

type projectRequest struct {
	Project interface{} `json:"project"`
}

func (pr *projectRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(pr)
}

type projectResponse struct {
	Project *Project `json:"project"`
}

func (pr *projectResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(pr)
}

type projectsResponse struct {
	Projects []*projectResponse `json:"projects"`
}

func (psr *projectsResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(psr)
}

func (psr *projectsResponse) projects() Projects {
	ps := make(Projects, 0, len(psr.Projects))
	for _, p := range psr.Projects {
		ps = append(ps, p.Project)
	}

	return ps
}

func (s *Service) List() (Projects, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	psresp := &projectsResponse{}
	err = psresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return psresp.projects(), nil
}

func (s *Service) Get(id int) (*Project, error) {
	return s.get(strconv.Itoa(id))
}

func (s *Service) New() (*Project, error) {
	return s.get("new")
}

func (s *Service) get(id string) (*Project, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+"/"+id+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	presp := &projectResponse{}
	err = presp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return presp.Project, nil
}

// Only the fields in ProjectCreate can be set.
func (s *Service) Create(p *Project) (*Project, error) {
	preq := &projectRequest{
		Project: &ProjectCreate{
			Archived: p.Archived,
			Name:     p.Name,
			Public:   p.Public,
		},
	}

	buf := &bytes.Buffer{}
	err := preq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.s.RoundTrip("POST", s.basePath+".json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	presp := &projectResponse{
		Project: p,
	}
	err = presp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Only the fields in ProjectUpdate can be set.
func (s *Service) Update(p *Project) error {
	preq := &projectRequest{
		Project: &ProjectUpdate{
			Archived: p.Archived,
			Name:     p.Name,
			Public:   p.Public,
		},
	}

	buf := &bytes.Buffer{}
	err := preq.Encode(buf)
	if err != nil {
		return err
	}

	resp, err := s.s.RoundTrip("PUT", s.basePath+"/"+strconv.Itoa(p.ID)+".json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(id int) error {
	resp, err := s.s.RoundTrip("DELETE", s.basePath+"/"+strconv.Itoa(id)+".json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Memberships(id int) (Memberships, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+"/"+strconv.Itoa(id)+"/memberships.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	psresp := &membershipsResponse{}
	err = psresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return psresp.memberships(), nil
}
