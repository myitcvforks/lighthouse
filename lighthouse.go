package lighthouse

import (
	"net/http"

	"github.com/nwidger/lighthouse/service"
	"github.com/nwidger/lighthouse/service/bin"
	"github.com/nwidger/lighthouse/service/changeset"
	"github.com/nwidger/lighthouse/service/message"
	"github.com/nwidger/lighthouse/service/milestone"
	"github.com/nwidger/lighthouse/service/profile"
	"github.com/nwidger/lighthouse/service/project"
	"github.com/nwidger/lighthouse/service/ticket"
	"github.com/nwidger/lighthouse/service/token"
	"github.com/nwidger/lighthouse/service/user"
)

func NewService(account, token string, rt http.RoundTripper) (*service.Service, error) {
	return service.New(account, token, rt)
}

func NewPublicService(account string, rt http.RoundTripper) (*service.Service, error) {
	return service.NewPublic(account, rt)
}

func NewBasicAuthService(account, username, password string, rt http.RoundTripper) (*service.Service, error) {
	return service.NewBasicAuth(account, username, password, rt)
}

func BinService(s *service.Service, projectID int) (*bin.Service, error) {
	return &bin.Service{
		ProjectID: projectID,
		Service:   s,
	}, nil
}

func ChangesetService(s *service.Service, projectID int) (*changeset.Service, error) {
	return &changeset.Service{
		ProjectID: projectID,
		Service:   s,
	}, nil
}

func MessageService(s *service.Service, projectID int) (*message.Service, error) {
	return &message.Service{
		ProjectID: projectID,
		Service:   s,
	}, nil
}

func MilestoneService(s *service.Service, projectID int) (*milestone.Service, error) {
	return &milestone.Service{
		ProjectID: projectID,
		Service:   s,
	}, nil
}

func ProfileService(s *service.Service) (*profile.Service, error) {
	return &profile.Service{
		Service: s,
	}, nil
}

func ProjectService(s *service.Service) (*project.Service, error) {
	return &project.Service{
		Service: s,
	}, nil
}

func TicketService(s *service.Service, projectID int) (*ticket.Service, error) {
	return &ticket.Service{
		ProjectID: projectID,
		Service:   s,
	}, nil
}

func TokenService(s *service.Service) (*token.Service, error) {
	return &token.Service{
		Service: s,
	}, nil
}

func UserService(s *service.Service) (*user.Service, error) {
	return &user.Service{
		Service: s,
	}, nil
}
