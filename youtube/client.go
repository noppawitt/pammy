package youtube

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/kkdai/youtube/v2"
	"github.com/tidwall/gjson"
)

var ErrNotFound = errors.New("not found")

type Client struct {
	*youtube.Client

	httpClient *http.Client
}

type VideoInfo struct {
	ID         string
	Title      string
	LengthText string
}

func NewClient() *Client {
	httpClient := http.DefaultClient
	return &Client{
		Client: &youtube.Client{
			HTTPClient: httpClient,
		},

		httpClient: httpClient,
	}
}

func (c *Client) SearchOne(query string) (VideoInfo, error) {
	results, err := c.Search(query)
	if err != nil {
		return VideoInfo{}, err
	}

	if len(results) == 0 {
		return VideoInfo{}, ErrNotFound
	}

	return results[0], nil
}

func (c *Client) Search(query string) ([]VideoInfo, error) {
	resp, err := c.httpClient.Get("https://www.youtube.com/results?search_query=" + url.QueryEscape(query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	infos, err := ExtractSearchResult(resp.Body)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func ExtractSearchResult(r io.Reader) ([]VideoInfo, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	s := string(data)

	startKeyword := "ytInitialData = "
	startIdx := strings.Index(s, startKeyword)
	if startIdx < 0 {
		return nil, errors.New("cannot find data starting point")
	}

	s = s[startIdx+len(startKeyword):]
	endIdx := strings.Index(s, ";</script>")
	if endIdx < 0 {
		return nil, errors.New("cannot find data ending point")
	}
	s = s[:endIdx]

	var infos []VideoInfo

	results := gjson.Get(s, "contents.twoColumnSearchResultsRenderer.primaryContents.sectionListRenderer.contents.0.itemSectionRenderer.contents.#.videoRenderer")
	results.ForEach(func(key, value gjson.Result) bool {
		info := VideoInfo{
			ID:         value.Get("videoId").Str,
			Title:      value.Get("title.runs.0.text").Str,
			LengthText: value.Get("lengthText.simpleText").Str,
		}

		infos = append(infos, info)
		return true
	})

	return infos, nil
}
