package service

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	StatusUnprocessableEntity = 422
)

var (
	DefaultTransport         = http.DefaultTransport
	DefaultInsecureTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
)

// Transport wraps another http.RoundTripper and ensures the outgoing
// request is authenticated with an API token
type Transport struct {
	// API token to use for authentication.  If set this is used
	// instead of Username/Password.
	Token string
	// If true, send API token in '_token' URL parameter
	TokenAsParameter bool

	// Username and password to use for authentication
	Username, Password string

	// Transport specifies the mechanism by which individual HTTP
	// requests are made.  If Transport is nil and Insecure is
	// false, DefaultTransport is used.  If Transport is nil and
	// Insecure is true, DefaultInsecureTransport is used.
	Insecure  bool
	Transport http.RoundTripper
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	if t.Insecure {
		return DefaultInsecureTransport
	}
	return DefaultTransport
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.Token) > 0 {
		if t.TokenAsParameter {
			req.URL.Query().Set("_token", t.Token)
		} else {
			req.Header.Set("X-LighthouseToken", t.Token)
		}
	} else if len(t.Username) > 0 && len(t.Password) > 0 {
		req.SetBasicAuth(t.Username, t.Password)
	}

	if len(req.Header.Get("Content-Type")) == 0 {
		switch filepath.Ext(req.URL.Path) {
		case ".json":
			req.Header.Set("Content-Type", "application/json")
		case ".xml":
			req.Header.Set("Content-Type", "application/xml")
		}
	}

	return t.transport().RoundTrip(req)
}

type Service struct {
	URL    string
	Client *http.Client
}

func makeURL(account string) string {
	return fmt.Sprintf("https://%s.lighthouseapp.com", account)
}

func New(account, token string, rt http.RoundTripper) (*Service, error) {
	return &Service{
		URL: makeURL(account),
		Client: &http.Client{
			Transport: &Transport{
				Token:     token,
				Transport: rt,
			},
		},
	}, nil
}

func NewPublic(account string, rt http.RoundTripper) (*Service, error) {
	return &Service{
		URL: makeURL(account),
		Client: &http.Client{
			Transport: &Transport{
				Transport: rt,
			},
		},
	}, nil
}

func NewBasicAuth(account, username, password string, rt http.RoundTripper) (*Service, error) {
	return &Service{
		URL: makeURL(account),
		Client: &http.Client{
			Transport: &Transport{
				Username:  username,
				Password:  password,
				Transport: rt,
			},
		},
	}, nil
}

func (s *Service) RoundTrip(method, path string, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(path, "http") {
		path = s.URL + path
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
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

	eu.Field = arr[0]
	eu.Message = arr[1]

	return nil
}

func (ve ErrUnprocessable) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

type ErrUnprocessables []ErrUnprocessable

func (ves ErrUnprocessables) Error() string {
	msg := ""
	for i, ve := range ves {
		if i > 0 {
			msg += ", "
		}
		msg += ve.Error()
	}
	return msg
}

type ErrInvalidResponse struct {
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

func newErrInvalidResponse(resp *http.Response) error {
	var err error

	defer resp.Body.Close()

	eir := &ErrInvalidResponse{
		Resp: resp,
	}

	if resp.StatusCode != StatusUnprocessableEntity {
		eir.BodyContents, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	} else {
		dec := json.NewDecoder(resp.Body)
		eir.Unprocessables = ErrUnprocessables{}

		err = dec.Decode(&eir.Unprocessables)
		if err != nil {
			return err
		}
	}

	return eir
}

func (eir *ErrInvalidResponse) Error() string {
	if eir.Unprocessables != nil {
		return eir.Unprocessables.Error()
	}

	return fmt.Sprintf("expected %d %s response, received %s",
		eir.ExpectedCode, http.StatusText(eir.ExpectedCode), eir.Resp.Status)
}

func CheckResponse(resp *http.Response, expected int) error {
	if resp.StatusCode != expected {
		return newErrInvalidResponse(resp)
	}
	return nil
}
