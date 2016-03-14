package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nwidger/lighthouse/service"
)

type Service struct {
	ProjectID int
	Service   *service.Service
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Tags []*Tag

type TagResponse struct {
	Tag *Tag `json:"tag"`
}

type TagsResponse struct {
	Tags []*TagResponse `json:"tags"`
}

type Attachment struct {
	AttachmentFileProcessing bool       `json:"attachment_file_processing"`
	Code                     string     `json:"code"`
	ContentType              string     `json:"content_type"`
	CreatedAt                *time.Time `json:"created_at"`
	Filename                 string     `json:"filename"`
	Height                   int        `json:"height"`
	ID                       int        `json:"id"`
	ProjectID                int        `json:"project_id"`
	Size                     int        `json:"size"`
	UploaderID               int        `json:"uploader_id"`
	Width                    int        `json:"width"`
	URL                      string     `json:"url"`
}

type Attachments []*Attachment

type AttachmentResponse struct {
	Attachment *Attachment `json:"attachment"`
}

type AttachmentsResponse struct {
	Attachments []*AttachmentResponse `json:"attachments"`
}

type AlphabeticalTag struct {
	Tag   string
	Count int
}

func (at *AlphabeticalTag) MarshalJSON() ([]byte, error) {
	tag, count := "", 0
	if at != nil {
		tag, count = at.Tag, at.Count
	}

	arr := []interface{}{tag, count}
	return json.Marshal(&arr)
}

func (at *AlphabeticalTag) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}

	if at == nil {
		at = &AlphabeticalTag{}
	}

	at.Tag = ""
	at.Count = 0

	arr := []interface{}{}
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err
	}

	if len(arr) != 2 {
		return fmt.Errorf("AlphabeticalTag.UnmarshalJSON: length is %d, expected 2", len(arr))
	}

	tag, ok := arr[0].(string)
	if !ok {
		return fmt.Errorf("AlphabeticalTag.UnmarshalJSON: first element not a string")
	}
	at.Tag = tag

	count, ok := arr[1].(float64)
	if !ok {
		return fmt.Errorf("AlphabeticalTag.UnmarshalJSON: first element not an int")
	}
	at.Count = int(count)

	return nil
}

type AlphabeticalTags []AlphabeticalTag

type DiffableAttributes struct {
	State        string `json:"state,omitempty"`
	Title        string `json:"title,omitempty"`
	AssignedUser int    `json:"assigned_user,omitempty"`
	Milestone    int    `json:"milestone,omitempty"`
	Tag          string `json:"tag,omitempty"`
}

type TicketVersion struct {
	AssignedUserID     int                 `json:"assigned_user_id"`
	AttachmentsCount   int                 `json:"attachments_count"`
	Body               string              `json:"body"`
	BodyHTML           string              `json:"body_html"`
	Closed             bool                `json:"closed"`
	CreatedAt          *time.Time          `json:"created_at"`
	CreatorID          int                 `json:"creator_id"`
	DiffableAttributes *DiffableAttributes `json:"diffable_attributes,omitempty"`
	Importance         int                 `json:"importance"`
	MilestoneID        int                 `json:"milestone_id"`
	MilestoneOrder     int                 `json:"milestone_order"`
	Number             int                 `json:"number"`
	Permalink          string              `json:"permalink"`
	ProjectID          int                 `json:"project_id"`
	RawData            []byte              `json:"raw_data"`
	Spam               bool                `json:"spam"`
	State              string              `json:"state,omitempty"`
	Tag                string              `json:"tag"`
	Title              string              `json:"title"`
	UpdatedAt          *time.Time          `json:"updated_at"`
	UserID             int                 `json:"user_id"`
	Version            int                 `json:"version"`
	WatchersIds        []int               `json:"watchers_ids"`
	UserName           string              `json:"user_name"`
	CreatorName        string              `json:"creator_name"`
	URL                string              `json:"url"`
	Priority           int                 `json:"priority"`
	StateColor         string              `json:"state_color"`
}

type TicketVersions []*TicketVersion

type Ticket struct {
	AssignedUserID   int                   `json:"assigned_user_id"`
	AttachmentsCount int                   `json:"attachments_count"`
	Body             string                `json:"body"`
	BodyHTML         string                `json:"body_html"`
	Closed           bool                  `json:"closed"`
	CreatedAt        *time.Time            `json:"created_at"`
	CreatorID        int                   `json:"creator_id"`
	Importance       int                   `json:"importance"`
	MilestoneDueOn   *time.Time            `json:"milestone_due_on"`
	MilestoneID      int                   `json:"milestone_id"`
	MilestoneOrder   int                   `json:"milestone_order"`
	Number           int                   `json:"number"`
	Permalink        string                `json:"permalink"`
	ProjectID        int                   `json:"project_id"`
	RawData          []byte                `json:"raw_data"`
	Spam             bool                  `json:"spam"`
	State            string                `json:"state,omitempty"`
	Tag              string                `json:"tag"`
	Title            string                `json:"title"`
	UpdatedAt        *time.Time            `json:"updated_at"`
	UserID           int                   `json:"user_id"`
	Version          int                   `json:"version"`
	WatchersIds      []int                 `json:"watchers_ids"`
	UserName         string                `json:"user_name"`
	CreatorName      string                `json:"creator_name"`
	AssignedUserName string                `json:"assigned_user_name"`
	URL              string                `json:"url"`
	MilestoneTitle   string                `json:"milestone_title"`
	Priority         int                   `json:"priority"`
	ImportanceName   string                `json:"importance_name"`
	OriginalBody     string                `json:"original_body"`
	LatestBody       string                `json:"latest_body"`
	OriginalBodyHTML string                `json:"original_body_html"`
	StateColor       string                `json:"state_color"`
	Tags             []*TagResponse        `json:"tags"`
	AlphabeticalTags AlphabeticalTags      `json:"alphabetical_tags"`
	Versions         TicketVersions        `json:"versions"`
	Attachments      []*AttachmentResponse `json:"attachments"`
}

type Tickets []*Ticket

type TicketCreate struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	State          string `json:"state,omitempty"`
	AssignedUserID int    `json:"assigned_user_id,omitempty"`
	MilestoneID    int    `json:"milestone_id,omitempty"`
	Tag            string `json:"tag"`

	// See:
	// http://help.lighthouseapp.com/discussions/api-developers/196-change-ticket-notifications
	NotifyAll        *bool `json:"notify_all,omitempty"`
	MultipleWatchers []int `json:"multiple_watchers,omitempty"`
}

type TicketUpdate struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	State          string `json:"state,omitempty"`
	AssignedUserID int    `json:"assigned_user_id"`
	MilestoneID    int    `json:"milestone_id"`
	Tag            string `json:"tag"`

	NotifyAll        *bool `json:"notify_all,omitempty"`
	MultipleWatchers []int `json:"multiple_watchers,omitempty"`
}

type ticketRequest struct {
	Ticket interface{} `json:"ticket"`
}

func (mr *ticketRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(mr)
}

type ticketResponse struct {
	Ticket *Ticket `json:"ticket"`
}

func (mr *ticketResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(mr)
}

type ticketsResponse struct {
	Tickets []*ticketResponse `json:"tickets"`
}

func (msr *ticketsResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(msr)
}

func (msr *ticketsResponse) tickets() Tickets {
	ms := make(Tickets, 0, len(msr.Tickets))
	for _, m := range msr.Tickets {
		ms = append(ms, m.Ticket)
	}

	return ms
}

func (s *Service) basePath() string {
	return "/projects/" + strconv.Itoa(s.ProjectID)
}

type ListOptions struct {
	Query string
	Limit int
	Page  int
}

func (s *Service) List(opts *ListOptions) (Tickets, error) {
	path := s.basePath() + "/tickets.json"
	if opts != nil {
		u, err := url.Parse(path)
		if err != nil {
			return nil, err
		}
		values := &url.Values{}
		if len(opts.Query) > 0 {
			values.Set("q", opts.Query)
		}
		if opts.Limit > 0 {
			values.Set("limit", strconv.Itoa(opts.Limit))
		}
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

	tsresp := &ticketsResponse{}
	err = tsresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return tsresp.tickets(), nil
}

// Only the fields in TicketUpdate can be set.
func (s *Service) Update(t *Ticket) error {
	treq := &ticketRequest{
		Ticket: &TicketUpdate{
			Title:          t.Title,
			Body:           t.Body,
			State:          t.State,
			AssignedUserID: t.AssignedUserID,
			MilestoneID:    t.MilestoneID,
			Tag:            t.Tag,
		},
	}

	buf := &bytes.Buffer{}
	err := treq.Encode(buf)
	if err != nil {
		return err
	}

	resp, err := s.Service.RoundTrip("PUT", s.basePath()+"/tickets/"+strconv.Itoa(t.Number)+".json", buf)
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

func (s *Service) New() (*Ticket, error) {
	return s.get("new")
}

func (s *Service) Get(number int) (*Ticket, error) {
	return s.get(strconv.Itoa(number))
}

func (s *Service) get(number string) (*Ticket, error) {
	resp, err := s.Service.RoundTrip("GET", s.basePath()+"/tickets/"+number+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	tresp := &ticketResponse{}
	err = tresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return tresp.Ticket, nil
}

// Only the fields in TicketCreate can be set.
func (s *Service) Create(m *Ticket) (*Ticket, error) {
	treq := &ticketRequest{
		Ticket: &TicketCreate{
			Title:          m.Title,
			Body:           m.Body,
			State:          m.State,
			AssignedUserID: m.AssignedUserID,
			MilestoneID:    m.MilestoneID,
			Tag:            m.Tag,
		},
	}

	buf := &bytes.Buffer{}
	err := treq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.Service.RoundTrip("POST", s.basePath()+"/tickets.json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	tresp := &ticketResponse{
		Ticket: m,
	}
	err = tresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Service) Delete(number int) error {
	resp, err := s.Service.RoundTrip("DELETE", s.basePath()+"/tickets/"+strconv.Itoa(number)+".json", nil)
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

func (s *Service) GetAttachment(a *Attachment) (io.ReadCloser, error) {
	resp, err := s.Service.RoundTrip("GET", a.URL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (s *Service) AddAttachment(t *Ticket, filename string, r io.Reader) error {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	attachmentPart, err := w.CreateFormFile("ticket[attachment][]", filepath.Base(filename))
	if err != nil {
		return err
	}

	_, err = io.Copy(attachmentPart, r)
	if err != nil {
		return err
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="json"`)
	h.Set("Content-Type", "application/json")

	ticketPart, err := w.CreatePart(h)
	if err != nil {
		return err
	}

	treq := &ticketRequest{
		Ticket: &TicketUpdate{
			Title:          t.Title,
			Body:           t.Body,
			State:          t.State,
			AssignedUserID: t.AssignedUserID,
			MilestoneID:    t.MilestoneID,
			Tag:            t.Tag,
		},
	}

	err = treq.Encode(ticketPart)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", s.Service.URL+s.basePath()+"/tickets/"+strconv.Itoa(t.Number)+".json", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := s.Service.Client.Do(req)
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
