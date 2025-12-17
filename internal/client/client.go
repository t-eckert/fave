package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/t-eckert/fave/internal"
)

type Client struct {
	host string
	http *http.Client
}

func New(host string) *Client {
	return &Client{
		host: host,
		http: &http.Client{},
	}
}

func (c *Client) Add(bookmark internal.Bookmark) (int, error) {
	j, err := json.Marshal(bookmark)
	if err != nil {
		return 0, err
	}
	resp, err := c.http.Post(c.host+"/bookmarks", "application/json", bytes.NewReader(j))

	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	id, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(id))
}

func (c *Client) List() (map[string]internal.Bookmark, error) {
	resp, err := c.http.Get(c.host + "/bookmarks")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bookmarks map[string]internal.Bookmark
	err = json.NewDecoder(resp.Body).Decode(&bookmarks)
	if err != nil {
		return nil, err
	}

	return bookmarks, nil
}

func (c *Client) Get(id string) (*internal.Bookmark, error) {
	resp, err := c.http.Get(c.host + "/bookmarks/" + id)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bookmark internal.Bookmark
	err = json.NewDecoder(resp.Body).Decode(&bookmark)
	if err != nil {
		return nil, err
	}

	return &bookmark, nil
}

func (c *Client) Delete(id string) (string, error) {
	req, err := http.NewRequest("DELETE", c.host+"/bookmarks/"+id, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)

	return string(respBody), nil
}
