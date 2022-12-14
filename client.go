package osdb

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/kolo/xmlrpc"
)

const (
	// DefaultOSDBServer is OSDB's API base URL.
	DefaultOSDBServer = "https://api.opensubtitles.org:443/xml-rpc"

	// DefaultUserAgent is the current version of this lib.
	DefaultUserAgent = "osdb-go 0.2"

	// SearchLimit = nax hits per search
	SearchLimit = 100

	// StatusSuccess is the successful response status for API calls.
	StatusSuccess = "200 OK"
)

// Client wraps an XML-RPC client to connect to OSDB.
type Client struct {
	UserAgent string
	Token     string
	Login     string
	Password  string
	Language  string
	*xmlrpc.Client
}

// Movie is a type that stores the information from IMDB searches.
type Movie struct {
	ID             string            `xmlrpc:"id"`
	Title          string            `xmlrpc:"title"`
	Cover          string            `xmlrpc:"cover"`
	Year           string            `xmlrpc:"year"`
	Duration       string            `xmlrpc:"duration"`
	TagLine        string            `xmlrpc:"tagline"`
	Plot           string            `xmlrpc:"plot"`
	Goofs          string            `xmlrpc:"goofs"`
	Trivia         string            `xmlrpc:"trivia"`
	Cast           map[string]string `xmlrpc:"cast"`
	Directors      map[string]string `xmlrpc:"directors"`
	Writers        map[string]string `xmlrpc:"writers"`
	Awards         []string          `xmlrpc:"awards"`
	Genres         []string          `xmlrpc:"genres"`
	Countries      []string          `xmlrpc:"country"`
	Languages      []string          `xmlrpc:"language"`
	Certifications []string          `xmlrpc:"certification"`
}

// Movies is just a slice of movies.
type Movies []Movie

// Empty checks whether Movies is empty.
func (m Movies) Empty() bool {
	return len(m) == 0
}

// FileSearch searches subtitles for a file and list of languages.
func (c *Client) FileSearch(path string, langs []string) (Subtitles, error) {
	// Hash file, and other params values.
	params, err := c.fileToSearchParams(path, langs)
	if err != nil {
		return nil, err
	}
	return c.SearchSubtitles(params)
}

// IMDBSearchByID searches subtitles that match some IMDB IDs.
func (c *Client) IMDBSearchByID(ids []string, langs []string) (Subtitles, error) {
	// OSDB search params struct
	params := []interface{}{
		c.Token,
		[]map[string]string{},
	}

	// Convert ids []string into a slice of map[string]string for search. Ouch!
	for _, imdbID := range ids {
		params[1] = append(
			params[1].([]map[string]string),
			map[string]string{
				"imdbid":        imdbID,
				"sublanguageid": strings.Join(langs, ","),
			},
		)
	}

	return c.SearchSubtitles(&params)
}

// IMDBSearchByIDFiltered Searches for a movie or tv episode by IMDB code. Set isMovie to true to ignore the season and episode
// otherwise the search try to find a subtitle that matches the season and episode supplied. Note: When looking for Episodes by the
// Series IMDB code, if you don't filter by episode/season - the result might include too many subtitles to return and thus
// some results will be ommited.
func (c *Client) IMDBSearchByIDFiltered(imdbCode string, isMovie bool, season uint, episode uint, lang []string) (Subtitles, error) {
	if !isMovie {
		params := []interface{}{
			c.Token,
			[]struct {
				Imdbid        string `xmlrpc:"imdbid"`
				Sublanguageid string `xmlrpc:"sublanguageid"`
				Season        int64  `xmlrpc:"season"`
				Episode       int64  `xmlrpc:"episode"`
			}{{
				imdbCode,
				strings.Join(lang, ","),
				int64(season),
				int64(episode),
			}},
		}
		return c.SearchSubtitles(&params)
	} else {
		return c.IMDBSearchByID([]string{imdbCode}, lang)
	}
}

// HashSearch Searches for subtitles that match a specific hash/size/language combination.
// This function does not require the path of the movie file, just the hash/size values.
func (c *Client) HashSearch(hash uint64, size int64, langs []string) (Subtitles, error) {
	if hash == 0 || size == 0 {
		return nil, errors.New("called OS search by Hash with no hash value")
	}
	if size < 15000 {
		return nil, errors.New("called OS search by Hash with a small file")
	}
	params := []interface{}{
		c.Token,
		[]struct {
			Hash  string `xmlrpc:"moviehash"`
			Size  int64  `xmlrpc:"moviebytesize"`
			Langs string `xmlrpc:"sublanguageid"`
		}{{
			fmt.Sprintf("%016x", hash),
			size,
			strings.Join(langs, ","),
		}},
	}
	return c.SearchSubtitles(&params)
}

// SearchSubtitles searches OSDB with your own parameters.
func (c *Client) SearchSubtitles(params *[]interface{}) (Subtitles, error) {
	res := struct {
		Data Subtitles `xmlrpc:"data"`
	}{}

	if err := c.Call("SearchSubtitles", *params, &res); err != nil {
		if strings.Contains(err.Error(), "type mismatch") {
			return nil, err
		}
	}
	return res.Data, nil
}

// IMDBSearch searches movies on IMDB.
func (c *Client) IMDBSearch(q string) (Movies, error) {
	params := []interface{}{c.Token, q}
	res := struct {
		Status string `xmlrpc:"status"`
		Data   Movies `xmlrpc:"data"`
	}{}
	if err := c.Call("SearchMoviesOnIMDB", params, &res); err != nil {
		return nil, err
	}
	if res.Status != StatusSuccess {
		return nil, fmt.Errorf("SearchMoviesOnIMDB error: %s", res.Status)
	}
	return res.Data, nil
}

// BestMoviesByHashes searches for the best matching movies for each
// of the hashes (only for <200). This returns incomplete Movies, with
// the following fields only: ID, Title and Year.
func (c *Client) BestMoviesByHashes(hashes []uint64) ([]*Movie, error) {
	hashStrings := make([]string, len(hashes))
	for i, hash := range hashes {
		hashStrings[i] = hashString(hash)
	}

	params := []interface{}{c.Token, hashStrings}
	res := struct {
		Status string                 `xmlrpc:"status"`
		Data   map[string]interface{} `xmlrpc:"data"`
	}{}

	if err := c.Call("CheckMovieHash", params, &res); err != nil {
		return nil, err
	}

	if res.Status != StatusSuccess {
		return nil, fmt.Errorf("CheckMovieHash error: %s", res.Status)
	}

	movies := make([]*Movie, len(hashes))
	for i, hashString := range hashStrings {
		switch v := res.Data[hashString].(type) {
		case []interface{}:
			// this works around a bug (feature?) in the opensubtitles API:
			// when a hash is missing in the database, the API returns
			// an empty array instead of a null value or an empty map
			movies[i] = nil
		case map[string]interface{}:
			// this is probably a movie
			movie, err := movieFromMap(v)
			if err != nil {
				return nil, fmt.Errorf(
					"CheckMovieHash returned malformed data: %s", err,
				)
			}
			movies[i] = movie
		default:
			return nil, fmt.Errorf("CheckMovieHash returned unknown data")
		}
	}

	return movies, nil
}

func movieFromMap(values map[string]interface{}) (*Movie, error) {
	movie := &Movie{}
	var ok bool
	if movie.ID, ok = values["MovieImdbID"].(string); !ok {
		return nil, fmt.Errorf("movie has malformed IMDB ID")
	}
	if movie.Title, ok = values["MovieName"].(string); !ok {
		return nil, fmt.Errorf("movie has malformed name")
	}
	if movie.Year, ok = values["MovieYear"].(string); !ok {
		return nil, fmt.Errorf("movie has malformed year")
	}
	return movie, nil
}

// GetIMDBMovieDetails fetches movie details from IMDB by ID.
func (c *Client) GetIMDBMovieDetails(id string) (*Movie, error) {
	params := []interface{}{c.Token, id}
	res := struct {
		Status string `xmlrpc:"status"`
		Data   Movie  `xmlrpc:"data"`
	}{}
	if err := c.Call("GetIMDBMovieDetails", params, &res); err != nil {
		return nil, err
	}
	if res.Status != StatusSuccess {
		return nil, fmt.Errorf("GetIMDBMovieDetails error: %s", res.Status)
	}
	return &res.Data, nil
}

// DownloadSubtitlesByIds downloads subtitles by ID.
func (c *Client) DownloadSubtitlesByIds(ids []int) ([]SubtitleFile, error) {
	params := []interface{}{c.Token, ids}
	res := struct {
		Status string         `xmlrpc:"status"`
		Data   []SubtitleFile `xmlrpc:"data"`
	}{}
	if err := c.Call("DownloadSubtitles", params, &res); err != nil {
		return nil, err
	}
	if res.Status != StatusSuccess {
		return nil, fmt.Errorf("DownloadSubtitles error: %s", res.Status)
	}
	return res.Data, nil
}

// DownloadSubtitles downloads subtitles in bulk.
func (c *Client) DownloadSubtitles(subtitles Subtitles) ([]SubtitleFile, error) {
	ids := make([]int, len(subtitles))
	for i := range subtitles {
		id, err := strconv.Atoi(subtitles[i].IDSubtitleFile)
		if err != nil {
			return nil, fmt.Errorf("malformed subtitle ID: %s", err)
		}
		ids[i] = id
	}

	subtitleFiles, err := c.DownloadSubtitlesByIds(ids)
	if err != nil {
		return nil, err
	}

	for i := range subtitleFiles {
		encodingName := subtitles[i].SubEncoding
		if encodingName != "" {
			subtitleFiles[i].Encoding, err = encodingFromName(subtitles[i].SubEncoding)
			if err != nil {
				return nil, err
			}
		}
	}

	return subtitleFiles, nil
}

// Download saves a subtitle file to disk, using the OSDB specified name.
func (c *Client) Download(s *Subtitle) error {
	return c.DownloadTo(s, s.SubFileName)
}

// DownloadTo saves a subtitle file to the specified path.
func (c *Client) DownloadTo(s *Subtitle, path string) (err error) {
	// Download
	files, err := c.DownloadSubtitles(Subtitles{*s})
	if err != nil {
		return
	}
	if len(files) == 0 {
		return fmt.Errorf("No file match this subtitle ID")
	}

	// Save to disk.
	r, err := files[0].Reader()
	if err != nil {
		return
	}
	defer r.Close()

	w, err := os.Create(path)
	if err != nil {
		return
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return
}

// HasSubtitles checks whether subtitles already exists in OSDB. The
// mandatory fields in the received Subtitle slice are: SubHash,
// SubFileName, MovieHash, MovieByteSize, and MovieFileName.
func (c *Client) HasSubtitles(subs Subtitles) (bool, error) {
	subArgs, err := subs.toTryUploadParams()
	if err != nil {
		return true, err
	}
	args := []interface{}{c.Token, subArgs}
	res := struct {
		Status string `xmlrpc:"status"`
		Exists int    `xmlrpc:"alreadyindb"`
	}{}
	if err := c.Call("TryUploadSubtitles", args, &res); err != nil {
		return true, err
	}
	if res.Status != StatusSuccess {
		return true, fmt.Errorf("HasSubtitles: %s", res.Status)
	}

	return res.Exists == 1, nil
}

// UploadSubtitles uploads subtitles.
//
// XXX Mandatory fields in the received Subtitle slice are: SubHash,
// SubFileName, MovieHash, SubLanguageID, MovieByteSize, and
// MovieFileName.
func (c *Client) UploadSubtitles(subs Subtitles) (string, error) {
	subArgs, err := subs.toUploadParams()
	if err != nil {
		return "", err
	}
	// XXX WIP
	// fmt.Println("Upload Params:", subArgs)
	return "", fmt.Errorf("WIP")

	args := []interface{}{c.Token, subArgs}
	res := struct {
		Status string `xmlrpc:"status"`
		URL    string `xmlrpc:"data"`
	}{}
	if err := c.Call("UploadSubtitles", args, &res); err != nil {
		return "", err
	}
	if res.Status != StatusSuccess {
		return "", fmt.Errorf("UploadSubtitles: %s", res.Status)
	}

	return res.URL, nil
}

// Noop keeps a session alive.
func (c *Client) Noop() (err error) {
	res := struct {
		Status string `xmlrpc:"status"`
	}{}
	err = c.Call("NoOperation", []interface{}{c.Token}, &res)
	if err == nil && res.Status != StatusSuccess {
		err = fmt.Errorf("NoOp: %s", res.Status)
	}
	return
}

// LogIn to the API, and return a session token.
func (c *Client) LogIn(user string, pass string, lang string) (err error) {
	c.Login = user
	c.Password = pass
	c.Language = lang
	args := []interface{}{user, pass, lang, c.UserAgent}
	res := struct {
		Status string `xmlrpc:"status"`
		Token  string `xmlrpc:"token"`
	}{}
	if err = c.Call("LogIn", args, &res); err != nil {
		return
	}

	if res.Status != StatusSuccess {
		return fmt.Errorf("Login: %s", res.Status)
	}
	c.Token = res.Token
	return
}

// LogOut ...
func (c *Client) LogOut() (err error) {
	args := []interface{}{c.Token}
	res := struct {
		Status string `xmlrpc:"status"`
	}{}
	return c.Call("LogOut", args, &res)
}

// Build query parameters for hash-based movie search.
func (c *Client) fileToSearchParams(path string, langs []string) (*[]interface{}, error) {
	// File size
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := fi.Size()

	// File hash
	h, err := HashFile(file)
	if err != nil {
		return nil, err
	}

	params := []interface{}{
		c.Token,
		[]struct {
			Hash  string `xmlrpc:"moviehash"`
			Size  int64  `xmlrpc:"moviebytesize"`
			Langs string `xmlrpc:"sublanguageid"`
		}{{
			hashString(h),
			size,
			strings.Join(langs, ","),
		}},
	}
	return &params, nil
}

// Create a string representation of hash
func hashString(hash uint64) string {
	return fmt.Sprintf("%016x", hash)
}

// Tries to guess the character encoding by its name
// (or whatever Opensubtitles thinks its name is)
func encodingFromName(name string) (encoding.Encoding, error) {
	return htmlindex.Get(name)
}
