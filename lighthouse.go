// Package lighthouse provides access to the Lighthouse API.
// http://help.lighthouseapp.com/kb/api
package lighthouse

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

const (
	StatusUnprocessableEntity = 422
)

// Transport wraps another http.RoundTripper and ensures the outgoing
// request is properly authenticated
type Transport struct {
	// API token to use for authentication.  If set this is used
	// instead of Email/Password.
	Token string
	// If Token is set and TokenAsBasicAuth is true, send API
	// token in Authorization header using Basic Authentication
	// with the API token as the username and 'x' as the password.
	TokenAsBasicAuth bool
	// If Token is set, TokenAsBasicAuth is false and
	// TokenAsParameter is true, send API token in '_token' URL
	// parameter.
	TokenAsParameter bool

	// Email and password to use for authentication.
	Email, Password string

	// Base specifies the mechanism by which individual HTTP
	// requests are made.  If Base is nil, http.DefaultTransport
	// is used.
	Base http.RoundTripper
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := cloneRequest(req) // per http.RoundTripper contract

	if len(t.Token) > 0 {
		if t.TokenAsBasicAuth {
			req2.SetBasicAuth(t.Token, "x")
		} else if t.TokenAsParameter {
			values := req2.URL.Query()
			values.Set("_token", t.Token)
			req2.URL.RawQuery = values.Encode()
		} else {
			req2.Header.Set("X-LighthouseToken", t.Token)
		}
	} else if len(t.Email) > 0 && len(t.Password) > 0 {
		req2.SetBasicAuth(t.Email, t.Password)
	}

	return t.base().RoundTrip(req2)
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

func NewClient(token string) *http.Client {
	return &http.Client{
		Transport: &Transport{
			Token: token,
		},
	}
}

func NewClientBasicAuth(email, password string) *http.Client {
	return &http.Client{
		Transport: &Transport{
			Email:    email,
			Password: password,
		},
	}
}

type Service struct {
	BasePath string
	Client   *http.Client
}

func BasePath(account string) string {
	return fmt.Sprintf("https://%s.lighthouseapp.com", account)
}

func NewService(account string, client *http.Client) *Service {
	return &Service{
		BasePath: BasePath(account),
		Client:   client,
	}
}

type Plan struct {
	Plan  string `xml:"plan"`
	Free  bool   `xml:"free"`
	Users int    `xml:"users"`

	//
	Projects int `xml:"projects"`
	Storage  int `xml:"storage"`
}

type planResponse struct {
	XMLName xml.Name `xml:"hash"`
	*Plan
}

func (pr *planResponse) decode(r io.Reader) error {
	dec := xml.NewDecoder(r)
	return dec.Decode(pr)
}

// Get account plan details.  Undocumented, see
// http://help.lighthouseapp.com/discussions/api-developers/1100-check-if-using-free-plan.
func (s *Service) Plan() (*Plan, error) {
	// using XML because JSON endpoint returns 406 Not Acceptable
	resp, err := s.RoundTrip("GET", s.BasePath+"/plan.xml", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	presp := &planResponse{}
	err = presp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return presp.Plan, nil
}

func (s *Service) RoundTrip(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		switch filepath.Ext(req.URL.Path) {
		case ".json":
			req.Header.Set("Content-Type", "application/json")
		case ".xml":
			req.Header.Set("Content-Type", "application/xml")
		}
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type ErrUnprocessable struct {
	Field   string
	Message string
}

func (eu *ErrUnprocessable) MarshalJSON() ([]byte, error) {
	field, message := "", ""
	if eu != nil {
		field, message = eu.Field, eu.Message
	}

	arr := []string{field, message}
	return json.Marshal(&arr)
}

func (eu *ErrUnprocessable) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	if eu == nil {
		eu = &ErrUnprocessable{}
	}

	eu.Field = ""
	eu.Message = ""

	arr := []string{}
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	if len(arr) != 2 {
		return fmt.Errorf("ErrUnprocessable.UnmarshalJSON: length is %d, expected 2", len(arr))
	}

	eu.Field, eu.Message = arr[0], arr[1]

	return nil
}

func (eu *ErrUnprocessable) Error() string {
	return fmt.Sprintf("%s: %s", eu.Field, eu.Message)
}

type ErrUnprocessables []*ErrUnprocessable

func (eus ErrUnprocessables) Error() string {
	msg := ""
	for i, ve := range eus {
		if i > 0 {
			msg += ", "
		}
		msg += ve.Error()
	}
	return msg
}

type ErrUnexpectedResponse struct {
	// The expected StatusCode
	ExpectedCode int

	// Resp.Body will always be closed.
	Resp *http.Response

	// BodyContents will contain the contents of Resp.Body if
	// Unprocessables is nil.
	BodyContents []byte

	// Unprocessables will not be nil if Resp.StatusCode was 422
	// StatusUnprocessableEntity.
	Unprocessables ErrUnprocessables
}

func newErrUnexpectedResponse(resp *http.Response, expected int) error {
	var err error

	defer resp.Body.Close()

	eur := &ErrUnexpectedResponse{
		ExpectedCode: expected,
		Resp:         resp,
	}

	if resp.StatusCode != StatusUnprocessableEntity {
		eur.BodyContents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	} else {
		dec := json.NewDecoder(resp.Body)
		eur.Unprocessables = ErrUnprocessables{}

		err = dec.Decode(&eur.Unprocessables)
		if err != nil {
			return err
		}
	}

	return eur
}

func (eir *ErrUnexpectedResponse) Error() string {
	if eir.Unprocessables != nil {
		return eir.Unprocessables.Error()
	}

	return fmt.Sprintf("expected %d %s response, received %s",
		eir.ExpectedCode, http.StatusText(eir.ExpectedCode), eir.Resp.Status)
}

func CheckResponse(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return newErrUnexpectedResponse(resp, expected)
	}
	return nil
}
