package message

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

type Comment struct {
	AllAttachmentsCount int        `json:"all_attachments_count"`
	AttachmentsCount    int        `json:"attachments_count"`
	Body                string     `json:"body"`
	BodyHTML            string     `json:"body_html"`
	CommentsCount       int        `json:"comments_count"`
	CreatedAt           *time.Time `json:"created_at"`
	ID                  int        `json:"id"`
	Integer             int        `json:"integer"`
	MilestoneID         int        `json:"milestone_id"`
	ParentID            int        `json:"parent_id"`
	Permalink           string     `json:"permalink"`
	ProjectID           int        `json:"project_id"`
	Title               string     `json:"title"`
	Token               string     `json:"token"`
	UpdatedAt           *time.Time `json:"updated_at"`
	UserID              int        `json:"user_id"`
	UserName            string     `json:"user_name"`
	URL                 string     `json:"url"`
}

type Comments []*Comment

type CommentCreate struct {
	Body  string `json:"body"`
	Title string `json:"title"`
}

type commentRequest struct {
	Comment interface{} `json:"comment"`
}

func (cr *commentRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(cr)
}

type Message struct {
	AllAttachmentsCount int        `json:"all_attachments_count"`
	AttachmentsCount    int        `json:"attachments_count"`
	Body                string     `json:"body"`
	BodyHTML            string     `json:"body_html"`
	CommentsCount       int        `json:"comments_count"`
	CreatedAt           *time.Time `json:"created_at"`
	ID                  int        `json:"id"`
	Integer             int        `json:"integer"`
	MilestoneID         int        `json:"milestone_id"`
	ParentID            int        `json:"parent_id"`
	Permalink           string     `json:"permalink"`
	ProjectID           int        `json:"project_id"`
	Title               string     `json:"title"`
	Token               string     `json:"token"`
	UpdatedAt           *time.Time `json:"updated_at"`
	UserID              int        `json:"user_id"`
	UserName            string     `json:"user_name"`
	URL                 string     `json:"url"`
	Comments            Comments   `json:"comments"`
}

type Messages []*Message

type MessageCreate struct {
	Body  string `json:"body"`
	Title string `json:"title"`
}

type MessageUpdate struct {
	Body  string `json:"body"`
	Title string `json:"title"`
}

type messageRequest struct {
	Message interface{} `json:"message"`
}

func (mr *messageRequest) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(mr)
}

type messageResponse struct {
	Message *Message `json:"message"`
}

func (mr *messageResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(mr)
}

func (mr *messageResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(mr)
}

type messagesResponse struct {
	Messages []*messageResponse `json:"messages"`
}

func (msr *messagesResponse) encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(msr)
}

func (msr *messagesResponse) decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(msr)
}

func (msr *messagesResponse) messages() Messages {
	ms := make(Messages, 0, len(msr.Messages))
	for _, m := range msr.Messages {
		ms = append(ms, m.Message)
	}

	return ms
}

func (s *Service) basePath() string {
	return "/projects/" + strconv.Itoa(s.ProjectID)
}

func (s *Service) List() (Messages, error) {
	resp, err := s.Service.RoundTrip("GET", s.basePath()+"/messages.json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	msresp := &messagesResponse{}
	err = msresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return msresp.messages(), nil
}

func (s *Service) New() (*Message, error) {
	return s.get("new")
}

// Only the fields in MessageUpdate can be set.
func (s *Service) Update(m *Message) error {
	mreq := &messageRequest{
		Message: &MessageUpdate{
			Body:  m.Body,
			Title: m.Title,
		},
	}

	buf := &bytes.Buffer{}
	err := mreq.Encode(buf)
	if err != nil {
		return err
	}

	resp, err := s.Service.RoundTrip("PUT", s.basePath()+"/messages/"+strconv.Itoa(m.ID)+".json", buf)
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

func (s *Service) Get(id int) (*Message, error) {
	return s.get(strconv.Itoa(id))
}

func (s *Service) get(id string) (*Message, error) {
	resp, err := s.Service.RoundTrip("GET", s.basePath()+"/messages/"+id+".json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}

	mresp := &messageResponse{}
	err = mresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return mresp.Message, nil
}

// Only the fields in MessageCreate can be set.
func (s *Service) Create(m *Message) (*Message, error) {
	mreq := &messageRequest{
		Message: &MessageCreate{
			Body:  m.Body,
			Title: m.Title,
		},
	}

	buf := &bytes.Buffer{}
	err := mreq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.Service.RoundTrip("POST", s.basePath()+"/messages.json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	mresp := &messageResponse{
		Message: m,
	}
	err = mresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Only the fields in CommentCreate can be set.
func (s *Service) CreateComment(id int, c *Comment) (*Message, error) {
	creq := &commentRequest{
		Comment: &CommentCreate{
			Body:  c.Body,
			Title: c.Title,
		},
	}

	buf := &bytes.Buffer{}
	err := creq.Encode(buf)
	if err != nil {
		return nil, err
	}

	resp, err := s.Service.RoundTrip("POST", s.basePath()+"/messages/"+strconv.Itoa(id)+"/comments.json", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = service.CheckResponse(resp, http.StatusCreated)
	if err != nil {
		return nil, err
	}

	m := &Message{}
	mresp := &messageResponse{
		Message: m,
	}
	err = mresp.decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (s *Service) Delete(id int) error {
	resp, err := s.Service.RoundTrip("DELETE", s.basePath()+"/messages/"+strconv.Itoa(id)+".json", nil)
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
