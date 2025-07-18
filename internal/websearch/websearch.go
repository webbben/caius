package websearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/webbben/caius/internal/llm"
	"github.com/webbben/caius/internal/utils"
	"github.com/webbben/caius/prompts"
)

var lastBraveAPICall time.Time
var braveAPIRate time.Duration = time.Second + (time.Millisecond * 50)
var braveAPICallMutex sync.Mutex

type searchResult struct {
	Title         string `json:"title"`
	Url           string `json:"url"`
	IsSourceLocal bool   `json:"is_source_local"`
	IsSourceBoth  bool   `json:"is_source_both"`
	Description   string `json:"description"`
	Profile       struct {
		Name     string `json:"name"`
		Url      string `json:"url"`
		LongName string `json:"long_name"`
		Img      string `json:"img"`
	} `json:"profile"`
}

type braveSearchResults struct {
	Type string `json:"type"`
	Web  struct {
		Type    string         `json:"type"`
		Results []searchResult `json:"results"`
	} `json:"web"`
}

// find API key from env variable
func getBraveSearchAPIKey() string {
	val, _ := os.LookupEnv("BRAVE_SEARCH_API_KEY")
	return val
}

func callBraveSearchAPI(searchPhrase string) (braveSearchResults, error) {
	apiKey := getBraveSearchAPIKey()
	if apiKey == "" {
		return braveSearchResults{}, errors.New("failed to get brave search API Key")
	}

	u, err := url.Parse("https://api.search.brave.com/res/v1/web/search")
	if err != nil {
		return braveSearchResults{}, err
	}
	q := u.Query()
	q.Set("q", searchPhrase)
	q.Set("count", "5")
	q.Set("country", "us")
	q.Set("search_lang", "en")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return braveSearchResults{}, utils.WrapError("error on requesting Brave Search API", err)
	}
	req.Header.Set("X-Subscription-Token", apiKey)

	// make sure we aren't calling the API more than once per second
	braveAPICallMutex.Lock()
	if time.Since(lastBraveAPICall) < braveAPIRate {
		time.Sleep(braveAPIRate - time.Since(lastBraveAPICall))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return braveSearchResults{}, utils.WrapError("error on requesting Brave Search API", err)
	}
	defer resp.Body.Close()
	lastBraveAPICall = time.Now()
	braveAPICallMutex.Unlock()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return braveSearchResults{}, utils.WrapError("error on reading Brave Search API response body", err)
	}
	var results braveSearchResults
	err = json.Unmarshal(body, &results)
	if err != nil {
		return results, utils.WrapError("error on unmarshalling Brave Search API response data", err)
	}

	return results, nil
}

type Website struct {
	URL         string
	Title       string
	Description string
	WebsiteName string
}

func WebSearch(searchPhrase string) ([]Website, error) {
	searchResults, err := callBraveSearchAPI(searchPhrase)
	if err != nil {
		return []Website{}, err
	}

	websites := []Website{}
	for _, result := range searchResults.Web.Results {
		websites = append(websites, Website{
			URL:         result.Url,
			Title:       result.Title,
			Description: RemoveAllHTML(result.Description),
			WebsiteName: result.Profile.Name,
		})
	}

	return websites, nil
}

func fetchURL(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Caius/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

func SummarizeWebsite(website Website) (string, error) {
	rawHTML, err := fetchURL(website.URL)
	if err != nil {
		return "", err
	}

	utils.WriteLogs(rawHTML)

	bodyText, err := extractBodyText(rawHTML)
	if err != nil {
		return "", err
	}

	utils.WriteLogs(bodyText)

	utils.Terminal.Lowkey("\n\n" + bodyText)

	prompt := fmt.Sprintf("Website: %s\nTitle: %s\n\n%s", website.WebsiteName, website.Title, bodyText)

	aiSummary, err := llm.GenerateSimpleCompletion(prompts.P_SUMMARIZE_WEBSITE, prompt)

	return aiSummary, nil
}

func SummarizeListOfWebsites(websites []Website) (string, error) {
	summaries := ""
	llm.SetModel(llm.Models.Llama3)
	for i, website := range websites {
		summary, err := SummarizeWebsite(website)
		if err != nil {
			log.Println("error summarizing website:", err)
			continue
		}
		summaries += summary + "\n---\n"
		utils.Terminal.Lowkey(fmt.Sprintf("[%v / %v] Website summarized", i+1, len(websites)))
	}

	fmt.Println()
	utils.Terminal.Lowkey(summaries)
	fmt.Println()

	// summarize all of the summaries
	llm.SetModel(llm.Models.DeepSeek14b)
	return llm.GenerateSimpleCompletion(prompts.P_SUMMARIZE_WEBSITE_LIST, summaries)
}
