## github-webhook-handler

github-webhook-handler is parse webhook from github and dispatch event to your event handler for GitHub Apps.

## Install

```
go get -u github.com/wreulicke/github-webhook-handler
```

## Usage 

```go
	var appID int64 = 0 // your app id
	bs, err := ioutil.ReadFile("your-key.pem")
	if err != nil {
		return nil
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(bs)
	if err != nil {
		return nil
	}
	r := http.NewServeMux()
	webhookHandler := webhook.New()
	appsTransport := ghinstallation.NewAppsTransportFromPrivateKey(http.DefaultTransport, appID, key),
	webhookHandler.On("issues", func(i *webhook.Installation, e *github.IssuesEvent) error {
		itr := ghinstallation.NewFromAppsTransport(appsTransport, i.Id)
		client := github.NewClient(&http.Client{Transport: itr})
		commentBody := e.Issue.GetTitle() + ":" + e.Issue.GetState()
		comment := &github.IssueComment{Body: &commentBody}
		_, _, err := client.Issues.CreateComment(context.Background(), *e.Repo.Owner.Login, *e.Repo.Name, *e.Issue.Number, comment)
		return err
	})
	r.Handle("/", webhookHandler)
	server := http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      r,
		Addr:         ":8080",
	}
	log.Printf("Server is started. Go to http://localhost:%d", 8080)
	if err := server.ListenAndServe(); err != nil {
		return err
	}
```

You can see fully example in [here](./example/)

## LICENSE 

MIT