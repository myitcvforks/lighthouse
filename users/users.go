// Package users provides access to users via the Lighthouse API.
// http://help.lighthouseapp.com/kb/api/users-and-membership.
package users

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
	"github.com/nwidger/lighthouse/projects"
)

type Service struct {
	basePath string
	s        *lighthouse.Service
}

func NewService(s *lighthouse.Service) *Service {
	return &Service{
		basePath: s.BasePath + "/users",
		s:        s,
	}
}

type ActiveTicket struct {
	Number    int
	Title     string
	URL       string
	UpdatedAt time.Time
}

func (at *ActiveTicket) MarshalJSON() ([]byte, error) {
	number, title, url, updatedAt := 0, "", "", float64(0)
	if at != nil {
		number, title, url, updatedAt = at.Number, at.Title, at.URL, float64(at.UpdatedAt.Unix())
	}

	arr := []interface{}{number, title, url, updatedAt}
	return json.Marshal(&arr)
}

func (at *ActiveTicket) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	if at == nil {
		at = &ActiveTicket{}
	}

	at.Number = 0
	at.Title = ""
	at.URL = ""
	at.UpdatedAt = time.Time{}

	arr := []interface{}{}
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	if len(arr) != 4 {
		return fmt.Errorf("ActiveTicket.UnmarshalJSON: length is %d, expected 4", len(arr))
	}

	number, ok := arr[0].(float64)
	if !ok {
		return fmt.Errorf("ActiveTicket.UnmarshalJSON: first element not an int")
	}
	at.Number = int(number)

	title, ok := arr[1].(string)
	if !ok {
		return fmt.Errorf("ActiveTicket.UnmarshalJSON: second element not a string")
	}
	at.Title = title

	url, ok := arr[2].(string)
	if !ok {
		return fmt.Errorf("ActiveTicket.UnmarshalJSON: third element not a string")
	}
	at.URL = url

	updatedAt, ok := arr[3].(float64)
	if !ok {
		return fmt.Errorf("ActiveTicket.UnmarshalJSON: fourth element not an int")
	}
	at.UpdatedAt = time.Unix(int64(updatedAt), 0)

	return nil
}

type ActiveTickets []*ActiveTicket

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

type User struct {
	ID            int           `json:"id"`
	Job           string        `json:"job"`
	Name          string        `json:"name"`
	Website       string        `json:"website"`
	ActiveTickets ActiveTickets `json:"active_tickets"`
}

type UserUpdate struct {
	ID      int    `json:"id"`
	Job     string `json:"job"`
	Name    string `json:"name"`
	Website string `json:"website"`
}

type userRequest struct {
	User interface{} `json:"user"`
}

func (ur *userRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(ur)
}

type userResponse struct {
	User *User `json:"user"`
}

func (ur *userResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(ur)
}

func (s *Service) Get(idOrName string) (*User, error) {
	id, err := lighthouse.ID(idOrName)
	if err == nil {
		return s.GetByID(id)
	}
	return s.GetByName(idOrName)
}

func (s *Service) GetByID(id int) (*User, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+"/"+strconv.Itoa(id)+".json", nil)
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

func (s *Service) GetByName(name string) (*User, error) {
	seen := map[int]struct{}{}
	projectService := projects.NewService(s.s)
	ps, err := projectService.List()
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(name)
	for _, p := range ps {
		ms, err := projectService.MembershipsByID(p.ID)
		if err != nil {
			return nil, err
		}
		for _, m := range ms {
			if _, ok := seen[m.User.ID]; ok {
				continue
			}
			seen[m.User.ID] = struct{}{}
			fullName := strings.ToLower(m.User.Name)
			firstName := fullName
			idx := strings.Index(fullName, " ")
			if idx != -1 {
				firstName = fullName[:idx]
			}
			switch {
			case fullName == lower, firstName == lower:
				return s.GetByID(m.User.ID)
			}

		}
	}
	return nil, fmt.Errorf("no such user %q", name)
}

// Only the fields in UserUpdate can be set.
func (s *Service) Update(u *User) error {
	ureq := &userRequest{
		User: &UserUpdate{
			ID:      u.ID,
			Job:     u.Job,
			Name:    u.Name,
			Website: u.Website,
		},
	}

	buf := &bytes.Buffer{}
	err := ureq.Encode(buf)
	if err != nil {
		return err
	}

	resp, err := s.s.RoundTrip("PUT", s.basePath+"/"+strconv.Itoa(u.ID)+".json", buf)
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

func (s *Service) Memberships(idOrName string) (Memberships, error) {
	id, err := lighthouse.ID(idOrName)
	if err == nil {
		return s.MembershipsByID(id)
	}
	return s.MembershipsByName(idOrName)
}

func (s *Service) MembershipsByID(id int) (Memberships, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+"/"+strconv.Itoa(id)+"/memberships.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	usresp := &membershipsResponse{}
	err = usresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return usresp.memberships(), nil
}

func (s *Service) MembershipsByName(name string) (Memberships, error) {
	u, err := s.GetByName(name)
	if err != nil {
		return nil, err
	}
	return s.MembershipsByID(u.ID)
}
