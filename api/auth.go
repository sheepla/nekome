package api

import (
	"context"
	"fmt"

	"github.com/dghubble/oauth1"
	twauth "github.com/dghubble/oauth1/twitter"
	"github.com/g8rswimmer/go-twitter/v2"
)

// Auth : アプリケーション認証を行う
func (a *API) Auth(client *oauth1.Token) (*User, error) {
	ct := getConsumerToken(client)
	config := oauth1.Config{
		ConsumerKey:    ct.Token,
		ConsumerSecret: ct.TokenSecret,
		CallbackURL:    "oob",
		Endpoint:       twauth.AuthorizeEndpoint,
	}

	requestToken, _, err := config.RequestToken()
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}

	authURL, err := config.AuthorizationURL(requestToken)
	if err != nil {
		return nil, fmt.Errorf("failed to issue authentication URL: %w", err)
	}

	fmt.Println("🐈 Go to the following URL to authenticate the application and enter the PIN that is displayed")
	fmt.Println("-----")
	fmt.Println(authURL.String())
	fmt.Print("PIN: ")

	var verifier string

	_, err = fmt.Scanf("%s", &verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to read PIN: %w", err)
	}

	accessToken, accessSecret, err := config.AccessToken(requestToken, "", verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain token: %w", err)
	}

	newToken := oauth1.NewToken(accessToken, accessSecret)

	user, err := a.authUserLookup(client, newToken)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain authenticated user: %w", err)
	}

	return &User{
		UserName: user.UserName,
		ID:       user.ID,
		Token:    newToken,
	}, nil
}

// authUserLookup : トークンに紐づいたユーザの情報を取得
func (a *API) authUserLookup(ct, ut *oauth1.Token) (*twitter.UserObj, error) {
	client := newClient(ct, ut)

	opts := twitter.UserLookupOpts{}
	res, err := client.AuthUserLookup(context.Background(), opts)

	if e := checkError(err); e != nil {
		return nil, e
	}

	if e := checkPartialError(res.Raw.Errors); res.Raw.Users[0] == nil && e != nil {
		return nil, e
	}

	return res.Raw.Users[0], nil
}
