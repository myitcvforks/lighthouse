Golang Lighthouse API
=====================

[![GoDoc](https://godoc.org/github.com/nwidger/lighthouse?status.svg)](https://godoc.org/github.com/nwidger/lighthouse)

A Golang client library for interacting with the
[Lighthouse API](http://lighthouseapp.com/api).

## Installation

```
go get -u github.com/nwidger/lighthouse
```

## Usage

``` go
import "github.com/nwidger/lighthouse"

// Create an *http.Client which will authenticate with your Lighthouse
// API token.
client := lighthouse.NewClient("your-api-token")

// Or create an *http.Client which will authenticate with your
// Lighthouse email/password.
client := lighthouse.NewClientBasicAuth("your-email", "your-password")

// Create a *lighthouse.Service with your Lighthouse account and client.
// 'https://your-account-name.lighthouseapp.com'.
s := lighthouse.NewService("your-account-name", client)

// Create a service for interacting with each resource type in your
// account.

// Some resources are project specific.
projectID := 123456

// http://help.lighthouseapp.com/kb/api/ticket-bins
binsService := bins.NewService(s, projectID)

// http://help.lighthouseapp.com/kb/api/changesets
changesetService := changesets.NewService(s, projectID)

// http://help.lighthouseapp.com/kb/api/messages
messagesService := messages.NewService(s, projectID)

// http://help.lighthouseapp.com/kb/api/milestones
milestonesService := milestones.NewService(s, projectID)

// http://help.lighthouseapp.com/kb/api/projects
projectsService := projects.NewService(s)

// http://help.lighthouseapp.com/kb/api/tickets
ticketsService := tickets.NewService(s, projectID)

// http://help.lighthouseapp.com/kb/api/users-and-membership
profilesService := profiles.NewService(s)
tokensService := tokens.NewService(s)
usersService := users.NewService(s)

// Call List(), Get(), New(), Create(), Update(), Delete(),
// etc. methods on service.
```

See [GoDoc reference](https://godoc.org/github.com/nwidger/lighthouse)
for more details on each service type.
