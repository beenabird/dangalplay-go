package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	AuthToken  = "jqeGWxRKK7FK5zEk3xCM"
	SecretKey  = "f53d31a4377e4ef31fa0"
	Region     = "IN"
	APIBase    = "https://ottapi.dangalplay.com"
	ConfigPath = "config/dangalplay_config.json"
)

type Config struct {
	UserID string `json:"user_id"`
	Proxy  string `json:"proxy,omitempty"`
}

type CLIArgs struct {
	URL          string
	Output       string
	Proxy        string
	Quality      string
	Season       int
	Episode      int
	EpisodeRange string
	Debug        bool
}

type ContentMetadata struct {
	CatalogID   string                 `json:"catalog_id"`
	ContentID   string                 `json:"content_id"`
	Title       string                 `json:"title"`
	Part        int                    `json:"part"`
	SequenceNo  int                    `json:"sequence_no"`
	SeasonTitle string                 `json:"season_title"`
	ItemCaption string                 `json:"item_caption"`
	Extra       map[string]interface{} `json:"-"`
}

type Downloader struct {
	Args         CLIArgs
	Config       Config
	Client       *http.Client
	ProxyURL     *url.URL
	ContentTitle string
}

func main() {
	args := parseArgs()
	downloader, err := NewDownloader(args)
	if err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}
	err = downloader.Run(args.URL)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}
}

func parseArgs() CLIArgs {
	var args CLIArgs
	flag.StringVar(&args.URL, "url", "", "Content URL")
	flag.StringVar(&args.Output, "o", "downloads", "Output directory")
	flag.StringVar(&args.Proxy, "p", "", "Proxy server URL")
	flag.StringVar(&args.Quality, "q", "720", "Video quality")
	flag.IntVar(&args.Season, "s", 0, "Season number")
	flag.IntVar(&args.Episode, "e", 0, "Episode number")
	flag.StringVar(&args.EpisodeRange, "w", "", "Episode range")
	flag.BoolVar(&args.Debug, "debug", false, "Enable debug mode")
	flag.Parse()
	return args
}

func NewDownloader(args CLIArgs) (*Downloader, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	dl := &Downloader{Args: args, Config: config, Client: &http.Client{}}
	if args.Proxy != "" {
		err := dl.setProxy(args.Proxy)
		if err != nil {
			return nil, err
		}
	}
	return dl, nil
}

func (d *Downloader) setProxy(proxyURL string) error {
	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}
	d.ProxyURL = parsed
	d.Client.Transport = &http.Transport{Proxy: http.ProxyURL(parsed)}
	return nil
}

func loadConfig() (Config, error) {
	content, err := ioutil.ReadFile(ConfigPath)
	if os.IsNotExist(err) {
		return createConfig()
	}
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(content, &config)
	return config, err
}

func createConfig() (Config, error) {
	fmt.Print("Enter DangalPlay USER_ID: ")
	reader := bufio.NewReader(os.Stdin)
	userID, _ := reader.ReadString('\n')
	config := Config{UserID: strings.TrimSpace(userID)}
	fmt.Print("Add proxy? (y/n): ")
	proxyChoice, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(proxyChoice)) == "y" {
		fmt.Print("Proxy URL: ")
		proxyURL, _ := reader.ReadString('\n')
		config.Proxy = strings.TrimSpace(proxyURL)
	}
	os.MkdirAll(filepath.Dir(ConfigPath), 0755)
	data, _ := json.MarshalIndent(config, "", "  ")
	err := ioutil.WriteFile(ConfigPath, data, 0644)
	return config, err
}

func (d *Downloader) Run(contentURL string) error {
	log.Printf("Starting download for URL: %s", contentURL)
	metadata, seriesTitle, err := d.getContentMetadata(contentURL)
	if err != nil {
		return err
	}
	if len(metadata) > 1 || d.isSeries(metadata) {
		return d.handleSeries(metadata, seriesTitle)
	}
	return d.handleMovie(metadata[0])
}

func (d *Downloader) isSeries(metadata []ContentMetadata) bool {
	return len(metadata) > 1 || (len(metadata) > 0 && metadata[0].Part > 0)
}

func extractYear(itemCaption string) string {
	parts := strings.Split(itemCaption, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) == 4 {
			if _, err := strconv.Atoi(part); err == nil {
				return part
			}
		}
	}
	return ""
}

func (d *Downloader) handleMovie(meta ContentMetadata) error {
	d.ContentTitle = sanitizeFilename(meta.Title)
	manifestURL, err := d.getStreamURL(meta)
	if err != nil {
		return err
	}
	year := extractYear(meta.ItemCaption)
	nameParts := []string{d.ContentTitle}
	if year != "" {
		nameParts = append(nameParts, year)
	}
	nameParts = append(nameParts, fmt.Sprintf("%sp", d.Args.Quality))
	movieDir := filepath.Join(d.Args.Output, "Movies", d.ContentTitle)
	os.MkdirAll(movieDir, 0755)
	return d.downloadContent(manifestURL, strings.Join(nameParts, "."), movieDir)
}

func (d *Downloader) handleSeries(episodes []ContentMetadata, seriesTitle string) error {
	filtered := d.filterEpisodes(episodes)
	if len(filtered) == 0 {
		return fmt.Errorf("no episodes match selection criteria")
	}
	d.ContentTitle = sanitizeFilename(seriesTitle)
	baseSeriesDir := filepath.Join(d.Args.Output, "TV Shows", d.ContentTitle)
	seasonMap := make(map[int][]ContentMetadata)
	for _, ep := range filtered {
		seasonMap[ep.Part] = append(seasonMap[ep.Part], ep)
	}
	for season, episodes := range seasonMap {
		seasonDir := filepath.Join(baseSeriesDir, fmt.Sprintf("Season %02d", season))
		os.MkdirAll(seasonDir, 0755)
		for _, ep := range episodes {
			manifestURL, err := d.getStreamURL(ep)
			if err != nil {
				return err
			}
			filename := fmt.Sprintf("%s.S%02dE%02d.%sp", d.ContentTitle, ep.Part, ep.SequenceNo, d.Args.Quality)
			err = d.downloadContent(manifestURL, filename, seasonDir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Downloader) getContentMetadata(urlStr string) ([]ContentMetadata, string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, "", err
	}
	pathParts := strings.Split(strings.Trim(u.Path, "/"), "/")
	switch pathParts[0] {
	case "movies":
		metadata, err := d.getMovieMetadata(pathParts[1])
		return metadata, "", err
	case "shows":
		return d.getSeriesMetadata(pathParts[1])
	default:
		return nil, "", fmt.Errorf("unsupported content type")
	}
}

func (d *Downloader) getMovieMetadata(movieID string) ([]ContentMetadata, error) {
	url := fmt.Sprintf("%s/catalogs/movies/items/%s.gzip", APIBase, movieID)
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("auth_token", AuthToken)
	q.Add("region", Region)
	req.URL.RawQuery = q.Encode()
	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct{ Data ContentMetadata }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return []ContentMetadata{result.Data}, nil
}

func (d *Downloader) getSeriesMetadata(showID string) ([]ContentMetadata, string, error) {
	url := fmt.Sprintf("%s/catalogs/shows/items/%s.gzip", APIBase, showID)
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("auth_token", AuthToken)
	q.Add("region", Region)
	q.Add("item_language", "hindi")
	req.URL.RawQuery = q.Encode()
	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			Title         string
			Subcategories []struct {
				FriendlyID  string
				EpisodeFlag string
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", err
	}
	var episodes []ContentMetadata
	for _, subcat := range result.Data.Subcategories {
		if subcat.EpisodeFlag == "yes" {
			eps, err := d.getSubcategoryEpisodes(showID, subcat.FriendlyID)
			if err != nil {
				return nil, "", err
			}
			episodes = append(episodes, eps...)
		}
	}
	return episodes, result.Data.Title, nil
}

func (d *Downloader) getSubcategoryEpisodes(showID, friendlyID string) ([]ContentMetadata, error) {
	var episodes []ContentMetadata
	page := 1
	for {
		url := fmt.Sprintf("%s/catalogs/shows/items/%s/subcategories/%s/episodes.gzip", APIBase, showID, friendlyID)
		req, _ := http.NewRequest("GET", url, nil)
		q := req.URL.Query()
		q.Add("auth_token", AuthToken)
		q.Add("region", Region)
		q.Add("order_by", "asc")
		q.Add("page", strconv.Itoa(page))
		q.Add("page_size", "100")
		req.URL.RawQuery = q.Encode()
		resp, err := d.Client.Do(req)
		if err != nil {
			return nil, err
		}
		var result struct {
			Data struct {
				Items   []ContentMetadata
				HasMore bool
			}
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()
		episodes = append(episodes, result.Data.Items...)
		if !result.Data.HasMore {
			break
		}
		page++
	}
	return episodes, nil
}

func (d *Downloader) getStreamURL(metadata ContentMetadata) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	hashInput := fmt.Sprintf("%s%s%s%s%s", metadata.CatalogID, metadata.ContentID, d.Config.UserID, timestamp, SecretKey)
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(hashInput)))
	payload := map[string]interface{}{
		"category":    "movies",
		"catalog_id":  metadata.CatalogID,
		"content_id":  metadata.ContentID,
		"region":      Region,
		"auth_token":  AuthToken,
		"id":          d.Config.UserID,
		"md5":         md5Hash,
		"ts":          timestamp,
	}
	resp, err := d.Client.Post(APIBase+"/v2/users/get_all_details.gzip", "application/json", bytes.NewBuffer(mustMarshal(payload)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			AdaptiveURL string
			HLSURLs    []struct{ URL string }
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Data.AdaptiveURL != "" {
		return result.Data.AdaptiveURL, nil
	}
	if len(result.Data.HLSURLs) > 0 {
		return result.Data.HLSURLs[0].URL, nil
	}
	return "", fmt.Errorf("no stream URLs found")
}

func (d *Downloader) downloadContent(manifestURL, filename, outputDir string) error {
	cmd := exec.Command("N_m3u8DL-RE", manifestURL,
		"--tmp-dir", "tmp",
		"--save-dir", outputDir,
		"--save-name", filename,
		"--log-level", "INFO",
		"--mux-after-done", "format=mp4",
		"--select-video", qualityToResolution(d.Args.Quality),
		"--select-audio", "best",
		"--select-subtitle", "best",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func qualityToResolution(q string) string {
	switch q {
	case "360": return "res=.*x360"
	case "480": return "res=.*x480"
	case "720": return "res=.*x720"
	case "1080": return "res=.*x1080"
	default: return "best"
	}
}

func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[^\w\s.-]`)
	name = reg.ReplaceAllString(name, "")
	name = strings.ReplaceAll(name, " ", ".")
	return strings.Trim(name, ".")
}

func (d *Downloader) filterEpisodes(episodes []ContentMetadata) []ContentMetadata {
	if d.Args.EpisodeRange != "" {
		season, start, end := parseEpisodeRange(d.Args.EpisodeRange)
		return filterByRange(episodes, season, start, end)
	}
	if d.Args.Season > 0 {
		if d.Args.Episode > 0 {
			return filterExactEpisode(episodes, d.Args.Season, d.Args.Episode)
		}
		return filterBySeason(episodes, d.Args.Season)
	}
	return episodes
}

func parseEpisodeRange(rangeStr string) (int, int, int) {
	re := regexp.MustCompile(`S(\d+)E(\d+)(?:-E?(\d+))?$`)
	matches := re.FindStringSubmatch(strings.ToUpper(rangeStr))
	if len(matches) < 3 {
		return 0, 0, 0
	}
	season, _ := strconv.Atoi(matches[1])
	start, _ := strconv.Atoi(matches[2])
	end := start
	if len(matches) > 3 && matches[3] != "" {
		end, _ = strconv.Atoi(matches[3])
	}
	return season, start, end
}

func filterByRange(episodes []ContentMetadata, season, start, end int) []ContentMetadata {
	var filtered []ContentMetadata
	for _, ep := range episodes {
		if ep.Part == season && ep.SequenceNo >= start && ep.SequenceNo <= end {
			filtered = append(filtered, ep)
		}
	}
	return filtered
}

func filterExactEpisode(episodes []ContentMetadata, season, episode int) []ContentMetadata {
	var filtered []ContentMetadata
	for _, ep := range episodes {
		if ep.Part == season && ep.SequenceNo == episode {
			filtered = append(filtered, ep)
		}
	}
	return filtered
}

func filterBySeason(episodes []ContentMetadata, season int) []ContentMetadata {
	var filtered []ContentMetadata
	for _, ep := range episodes {
		if ep.Part == season {
			filtered = append(filtered, ep)
		}
	}
	return filtered
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
