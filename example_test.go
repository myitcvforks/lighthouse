package lighthouse_test

import (
	"fmt"
	"log"

	"github.com/nwidger/lighthouse"
	"github.com/nwidger/lighthouse/tickets"
)

func ExampleNewService() {
	// Create an *http.Client which will authenticate with your Lighthouse
	// API token.
	client := lighthouse.NewClient("your-api-token")

	// Create a *lighthouse.Service with your Lighthouse account and client.
	// 'https://your-account-name.lighthouseapp.com'.
	s := lighthouse.NewService("your-account-name", client)

	// Create a *tickets.Service instance for interacting with
	// project 123456's tickets.
	// http://help.lighthouseapp.com/kb/api/tickets
	ticketsService := tickets.NewService(s, 123456)

	// Search the project's tickets.
	// http://help.lighthouseapp.com/kb/getting-started/how-do-i-search-for-tickets
	ts, err := ticketsService.List(&tickets.ListOptions{
		Query: `responsible:me milestone:v1.2 tagged:bug sort:updated`,
		Limit: tickets.MaxLimit,
		Page:  1,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range ts {
		fmt.Println(t.Number, t.Title, t.Tags, t.Priority, t.State)
	}
}
