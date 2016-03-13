Golang Lighthouse API
=====================

[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/nwidger/lighthouse)

A Golang client library for interacting with the
[Lighthouse API](http://lighthouseapp.com/api).

## Installation

```
go get -u github.com/nwidger/lighthouse
```

## Usage

``` go
// create a Lighthouse service instance by passing in your account
// name and an API token.
s, err := lighthouse.NewService("your-account-name", "your-api-token", nil)
if err != nil {
	log.Fatal(err)
}

// OR create a Lighthouse service instance with
// your Lighthouse username/password.
s, err := lighthouse.NewBasicAuthService("your-account-name", "your-username", "your-password", nil)
if err != nil {
	log.Fatal(err)
}

// OR create a Lighthouse service instance for interacting
// with public Lighthouse projects.
s, err := lighthouse.NewPublicService("public-account-name", nil)
if err != nil {
	log.Fatal(err)
}

// use Lighthouse service instance to create service for interacting
// with a specific resource type
projectID := 123456

// http://help.lighthouseapp.com/kb/api/ticket-bins
binService, err := s.BinService(s, projectID)

// http://help.lighthouseapp.com/kb/api/changesets
changesetService, err := s.ChangesetService(s, projectID)

// http://help.lighthouseapp.com/kb/api/messages
messageService, err := s.MessageService(s, projectID)

// http://help.lighthouseapp.com/kb/api/milestones
milestoneService, err := s.MilestoneService(s, projectID)

// http://help.lighthouseapp.com/kb/api/projects
projectService, err := s.ProjectService(s)

// http://help.lighthouseapp.com/kb/api/tickets
ticketService, err := s.TicketService(s, projectID)

// http://help.lighthouseapp.com/kb/api/users-and-membership
profileService, err := s.ProfileService(s)
tokenService, err := s.TokenService(s)
userService, err := s.UserService(s)

// call List(), Get(), New(), Update(), Delete(), etc. methods on
// service.

// See GoDoc for more details on methods available for each resource
//type.
```
