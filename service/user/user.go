package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nwidger/lighthouse/service"
)

type Service struct {
	Service *service.Service
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

type ActiveTickets []ActiveTicket

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

func (pr *membershipResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(pr)
}

func (pr *membershipResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(pr)
}

type membershipsResponse struct {
	Memberships []*membershipResponse `json:"memberships"`
}

func (psr *membershipsResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(psr)
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

func (ur *userResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(ur)
}

func (ur *userResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(ur)
}

func (s *Service) Get(id int) (*User, error) {
	resp, err := s.Service.RoundTrip("GET", "/users/"+strconv.Itoa(id)+".json", nil)
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

	resp, err := s.Service.RoundTrip("PUT", "/users/"+strconv.Itoa(u.ID)+".json", buf)
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

func (s *Service) Memberships(id int) (Memberships, error) {
	resp, err := s.Service.RoundTrip("GET", "/users/"+strconv.Itoa(id)+"/memberships.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
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
