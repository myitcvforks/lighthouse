// Package changesets provides access to a project's changesets via
// the Lighthouse API.
// http://help.lighthouseapp.com/kb/api/changesets.
package changesets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nwidger/lighthouse"
)

type Service struct {
	basePath string
	s        *lighthouse.Service
}

func NewService(s *lighthouse.Service, projectID int) *Service {
	return &Service{
		basePath: s.BasePath + "/projects/" + strconv.Itoa(projectID) + "/changesets",
		s:        s,
	}
}

type Change struct {
	Operation string
	Path      string
}

func (c *Change) MarshalJSON() ([]byte, error) {
	operation, path := "", ""
	if c != nil {
		operation, path = c.Operation, c.Path
	}

	arr := []string{operation}
	if len(path) > 0 {
		arr = append(arr, path)
	}
	return json.Marshal(&arr)
}

func (c *Change) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	if c == nil {
		c = &Change{}
	}

	c.Operation = ""
	c.Path = ""

	arr := []string{}
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	// if a changeset is extremely long, Lighthouse appears to
	// truncate the list of changes which can result in it
	// returning a 1-element change array containing only the
	// operation [<op>] rather than the normal 2-element change
	// array [<op>, <path>] containing both the operation and
	// path.  We'll support both, with a 1-element change array
	// resulting in a Change with an empty path.
	if len(arr) < 1 || len(arr) > 2 {
		return fmt.Errorf("Change.UnmarshalJSON: length is %d, expected 1 or 2", len(arr))
	}

	c.Operation = arr[0]
	if len(arr) == 2 {
		c.Path = arr[1]
	}

	return nil
}

type Changes []*Change

type Changeset struct {
	Body      string     `json:"body"`
	BodyHTML  string     `json:"body_html"`
	ChangedAt *time.Time `json:"changed_at"`
	Changes   Changes    `json:"changes"`
	Committer string     `json:"committer"`
	ProjectID int        `json:"project_id"`
	Revision  string     `json:"revision"`
	TicketID  int        `json:"ticket_id"`
	Title     string     `json:"title"`
	UserID    int        `json:"user_id"`
}

type Changesets []*Changeset

type ChangesetCreate struct {
	Body      string     `json:"body"`
	BodyHTML  string     `json:"body_html"`
	ChangedAt *time.Time `json:"changed_at"`
	Changes   Changes    `json:"changes"`
	Revision  string     `json:"revision"`
	Title     string     `json:"title"`
	UserID    int        `json:"user_id"`
}

type changesetRequest struct {
	Changeset interface{} `json:"changeset"`
}

func (cr *changesetRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(cr)
}

type changesetResponse struct {
	Changeset *Changeset `json:"changeset"`
}

func (cr *changesetResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(cr)
}

type changesetsResponse struct {
	ChangesetResponse []*changesetResponse `json:"changesets"`
}

func (csr *changesetsResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(csr)
}

func (csr *changesetsResponse) changesets() Changesets {
	cs := make(Changesets, 0, len(csr.ChangesetResponse))
	for _, c := range csr.ChangesetResponse {
		cs = append(cs, c.Changeset)
	}

	return cs
}

type ListOptions struct {
	// Undocumented.  If non-zero, the page to return.
	Page int
}

func (s *Service) List(opts *ListOptions) (Changesets, error) {
	path := s.basePath + ".json"
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
		path = u.String()
	}

	resp, err := s.s.RoundTrip("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	csresp := &changesetsResponse{}
	err = csresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return csresp.changesets(), nil
}

// ListAll repeatedly calls List and returns all pages.  ListAll
// ignores opts.Page.
func (s *Service) ListAll(opts *ListOptions) (Changesets, error) {
	realOpts := ListOptions{}
	if opts != nil {
		realOpts = *opts
	}

	cs := Changesets{}

	for realOpts.Page = 1; ; realOpts.Page++ {
		p, err := s.List(&realOpts)
		if err != nil {
			return nil, err
		}
		if len(p) == 0 {
			break
		}

		cs = append(cs, p...)
	}

	return cs, nil
}

func (s *Service) New() (*Changeset, error) {
	return s.Get("new")
}

func (s *Service) Get(revision string) (*Changeset, error) {
	resp, err := s.s.RoundTrip("GET", s.basePath+"/"+revision+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = lighthouse.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	cresp := &changesetResponse{}
	err = cresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return cresp.Changeset, nil
}

// Only the fields in ChangesetCreate can be set.
func (s *Service) Create(c *Changeset) (*Changeset, error) {
	creq := &changesetRequest{
		Changeset: &ChangesetCreate{
			Body:      c.Body,
			BodyHTML:  c.BodyHTML,
			ChangedAt: c.ChangedAt,
			Changes:   c.Changes,
			Revision:  c.Revision,
			Title:     c.Title,
			UserID:    c.UserID,
		},
	}

	buf := &bytes.Buffer{}
	err := creq.Encode(buf)
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

	cresp := &changesetResponse{
		Changeset: c,
	}
	err = cresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Service) Delete(revision string) error {
	resp, err := s.s.RoundTrip("DELETE", s.basePath+"/"+revision+".json", nil)
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
