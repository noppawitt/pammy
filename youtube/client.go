package youtube

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/tidwall/gjson"
)

var ErrNotFound = errors.New("not found")

type Client struct {
	*youtube.Client

	httpClient *http.Client
}

type VideoInfo struct {
	ID       string
	Title    string
	Duration time.Duration
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

func (c *Client) GetSuggestedVideos(id string) ([]VideoInfo, error) {
	resp, err := c.httpClient.Get("https://www.youtube.com/watch?v=" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	infos, err := ExtractSuggestedVideos(resp.Body)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func (c *Client) GetAudioStreamURL(video *youtube.Video) (string, error) {
	format := c.filterAudioChannel(video.Formats)
	if format == nil {
		return "", errors.New("no audio format")
	}

	var (
		streamURL string
		err       error
		lastErr   error
	)

	for i := 0; i < 3; i++ {
		streamURL, err = c.GetStreamURL(video, format)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := c.httpClient.Head(streamURL)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("invalid response code: %d", resp.StatusCode)
		}

		time.Sleep(200 * time.Millisecond)
	}

	return streamURL, lastErr
}

var opusItags = [...]int{249, 250, 251}

func (c *Client) filterAudioChannel(formats youtube.FormatList) *youtube.Format {
	for _, itag := range opusItags {
		format := formats.FindByItag(itag)
		if format != nil {
			return format
		}
	}

	return nil
}

func extractJSONData(r io.Reader) ([]byte, error) {
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

	return []byte(s), nil
}

func ExtractSearchResult(r io.Reader) ([]VideoInfo, error) {
	data, err := extractJSONData(r)
	if err != nil {
		return nil, err
	}

	var infos []VideoInfo

	results := gjson.GetBytes(data, "contents.twoColumnSearchResultsRenderer.primaryContents.sectionListRenderer.contents.0.itemSectionRenderer.contents.#.videoRenderer")
	results.ForEach(func(key, value gjson.Result) bool {
		info := VideoInfo{
			ID:       value.Get("videoId").Str,
			Title:    value.Get("title.runs.0.text").Str,
			Duration: durationFromLengthText(value.Get("lengthText.simpleText").Str),
		}

		infos = append(infos, info)
		return true
	})

	return infos, nil
}

func ExtractSuggestedVideos(r io.Reader) ([]VideoInfo, error) {
	data, err := extractJSONData(r)
	if err != nil {
		return nil, err
	}

	var infos []VideoInfo

	results := gjson.GetBytes(data, "contents.twoColumnWatchNextResults.secondaryResults.secondaryResults.results.#.compactVideoRenderer")
	results.ForEach(func(key, value gjson.Result) bool {
		info := VideoInfo{
			ID:       value.Get("videoId").Str,
			Title:    value.Get("title.simpleText").Str,
			Duration: durationFromLengthText(value.Get("lengthText.simpleText").Str),
		}

		infos = append(infos, info)
		return true
	})

	return infos, nil
}

func durationFromLengthText(text string) time.Duration {
	parts := strings.Split(text, ":")

	if len(parts) > 3 {
		return 0
	}

	var d time.Duration
	units := [3]time.Duration{time.Second, time.Minute, time.Hour}
	for i := 0; i < len(parts); i++ {
		n, err := strconv.Atoi(parts[len(parts)-1-i])
		if err != nil {
			return 0
		}

		d += time.Duration(n) * units[i]
	}

	return d
}
