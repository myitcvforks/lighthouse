package lighthouse

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	StatusUnprocessableEntity = 422
)

// Transport wraps another http.RoundTripper and ensures the outgoing
// request is properly authenticated
type Transport struct {
	// API token to use for authentication.  If set this is used
	// instead of Username/Password.
	Token string
	// If Token is set and TokenAsParameter is true, send API
	// token in '_token' URL parameter.
	TokenAsParameter bool

	// Username and password to use for authentication.
	Username, Password string

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
	if len(t.Token) > 0 {
		if t.TokenAsParameter {
			values := req.URL.Query()
			values.Set("_token", t.Token)
			req.URL.RawQuery = values.Encode()
		} else {
			req.Header.Set("X-LighthouseToken", t.Token)
		}
	} else if len(t.Username) > 0 && len(t.Password) > 0 {
		req.SetBasicAuth(t.Username, t.Password)
	}

	return t.base().RoundTrip(req)
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
			Username: email,
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

func NewService(account string, client *http.Client) (*Service, error) {
	return &Service{
		BasePath: BasePath(account),
		Client:   client,
	}, nil
}

func (s *Service) RoundTrip(method, path string, body io.Reader) (*http.Response, error) {
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
