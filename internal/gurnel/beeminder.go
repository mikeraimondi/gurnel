package gurnel

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type beeminderClient struct {
	Token     []byte
	User      string
	c         http.Client
	serverURL string
}

func newBeeminderClient(user string, token []byte) (*beeminderClient, error) {
	if user == "" {
		return nil, fmt.Errorf("user must not be blank")
	}
	if token != nil && len(token) == 0 {
		return nil, fmt.Errorf("token must not be blank")
	}

	return &beeminderClient{
		Token:     bytes.TrimSpace(token),
		User:      user,
		serverURL: "https://www.beeminder.com",
	}, nil
}

func (client *beeminderClient) postDatapoint(
	goal string,
	count int,
	t time.Time,
) error {
	if goal == "" {
		return fmt.Errorf("goal must not be blank")
	}
	if count < 0 {
		return fmt.Errorf("count must be nonnegative")
	}

	postURL, err := url.Parse(client.serverURL)
	if err != nil {
		return fmt.Errorf("internal URL error: %w", err)
	}
	postURL.Path = fmt.Sprintf("api/v1/users/%s/goals/%s/datapoints.json",
		client.User, goal)

	v := url.Values{}
	v.Set("auth_token", string(client.Token))
	v.Set("value", strconv.Itoa(count))
	v.Set("comment", "via Gurnel at "+t.Format("15:04:05 MST"))

	resp, err := client.c.PostForm(postURL.String(), v)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respData, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(respData) == 0 {
			respData = []byte("no further info")
		}
		return fmt.Errorf("server returned %s: %s", resp.Status, respData)
	}
	return nil
}
