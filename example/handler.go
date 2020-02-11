package main

import (
	"context"
	"crypto/rsa"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v29/github"
	webhook "github.com/wreulicke/github-webhook-handler"
)

type Handler struct {
	AppsTransport *ghinstallation.AppsTransport
}

func (h *Handler) Issue(i *webhook.Installation, e *github.IssuesEvent) error {
	itr := ghinstallation.NewFromAppsTransport(h.AppsTransport, i.Id)
	client := github.NewClient(&http.Client{Transport: itr})
	commentBody := e.Issue.GetTitle() + ":" + e.Issue.GetState()
	comment := &github.IssueComment{Body: &commentBody}
	_, _, err := client.Issues.CreateComment(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, *e.Issue.Number, comment)
	return err
}

func (h *Handler) IssueComment(i *webhook.Installation, e *github.IssueCommentEvent) error {
	if *e.Comment.User.Type == "Bot" {
		return nil
	}
	itr := ghinstallation.NewFromAppsTransport(h.AppsTransport, i.Id)
	client := github.NewClient(&http.Client{Transport: itr})
	commentBody := e.Issue.GetTitle() + ":" + e.Issue.GetState() + ":" + e.Comment.GetBody()
	comment := &github.IssueComment{Body: &commentBody}
	_, _, err := client.Issues.CreateComment(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, *e.Issue.Number, comment)
	return err
}

func NewHandler(appID int64, key *rsa.PrivateKey) *webhook.Handlers {
	handler := &Handler{
		AppsTransport: ghinstallation.NewAppsTransportFromPrivateKey(http.DefaultTransport, appID, key),
	}
	webhookHandler := webhook.New()
	webhookHandler.On("issues", handler.Issue)
	webhookHandler.On("issue_comment", handler.IssueComment)
	return webhookHandler
}
