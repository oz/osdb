package osdb

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/kolo/xmlrpc"
)

const (
	OSDBServer       = "http://api.opensubtitles.org/xml-rpc"
	DefaultUserAgent = "OS Test User Agent" // OSDB's test agent
	SearchLimit      = 100
	StatusSuccess    = "200 OK"
)

type Client struct {
	UserAgent string
	Token     string
	Login     string
	Password  string
	Language  string
	*xmlrpc.Client
}

type Movie struct {
	Id    string `xmlrpc:"id"`
	Title string `xmlrpc:"title"`
}

// Search subtitles matching a file hash.
func (c *Client) FileSearch(path string, langs []string) ([]Subtitle, error) {
	// Hash file, and other params values.
	params, err := c.hashSearchParams(path, langs)
	if err != nil {
		return nil, err
	}
	return c.SearchSubtitles(params)
}

// Search subtitles matching IMDB IDs.
func (c *Client) ImdbIdSearch(ids []string, langs []string) ([]Subtitle, error) {
	// OSDB search params struct
	params := []interface{}{
		c.Token,
		[]map[string]string{},
	}

	// Convert ids []string into a slice of map[string]string for search. Ouch!
	for _, imdbId := range ids {
		params[1] = append(
			params[1].([]map[string]string),
			map[string]string{
				"imdbid":        imdbId,
				"sublanguageid": strings.Join(langs, ","),
			},
		)
	}

	return c.SearchSubtitles(&params)
}

// Search Subtitles, DIY method.
func (c *Client) SearchSubtitles(params *[]interface{}) ([]Subtitle, error) {
	res := struct {
		Data []Subtitle `xmlrpc:"data"`
	}{}

	if err := c.Call("SearchSubtitles", *params, &res); err != nil {
		return nil, err
	}

	return res.Data, nil
}

func (c *Client) SearchOnImdb(q string) ([]Movie, error) {
	params := []interface{}{c.Token, q}
	res := struct {
		Status string  `xmlrpc:"status"`
		Data   []Movie `xmlrpc:"data"`
	}{}
	if err := c.Call("SearchMoviesOnIMDB", params, &res); err != nil {
		return nil, err
	}
	if res.Status != StatusSuccess {
		return nil, fmt.Errorf("SearchMoviesOnIMDB error: %s", res.Status)
	}
	return res.Data, nil
}

// Download subtitles by file ID.
func (c *Client) DownloadSubtitles(ids []int) ([]SubtitleFile, error) {
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

// Save subtitle file to disk, using the OSDB specified name.
func (c *Client) Download(s *Subtitle) error {
	return c.DownloadTo(s, s.SubFileName)
}

// Save subtitle file to disk, using the specified path.
func (c *Client) DownloadTo(s *Subtitle, path string) (err error) {
	id, err := strconv.Atoi(s.IDSubtitleFile)
	if err != nil {
		return
	}

	// Download
	files, err := c.DownloadSubtitles([]int{id})
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

// Checks wether subtitles already exists in OSDB. The mandatory fields in the
// received Subtitle slice are: SubHash, SubFileName, MovieHash, MovieByteSize,
// and MovieFileName.
func (c *Client) HasSubtitles(subs []Subtitle) (bool, error) {
	args := c.hasSubtitlesParams(subs)
	res := struct {
		Status string   `xmlrpc:"status"`
		Exists int      `xmlrpc:"alreadyindb"`
		Data   Subtitle `xmlrpc:"data"`
	}{}
	if err := c.Call("TryUploadSubtitles", args, &res); err != nil {
		return false, err
	}
	if res.Status != StatusSuccess {
		return false, fmt.Errorf("HasSubtitles error: %s", res.Status)
	}

	return res.Exists == 1, nil
}

// Keep session alive
func (c *Client) Noop() (err error) {
	res := struct {
		Status string `xmlrpc:"status"`
	}{}
	err = c.Call("NoOperation", []interface{}{c.Token}, &res)
	if err == nil && res.Status != StatusSuccess {
		err = fmt.Errorf("NoOp error: %s", res.Status)
	}
	return
}

// Login to the API, and return a session token.
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
		return fmt.Errorf("login error: %s", res.Status)
	}
	c.Token = res.Token
	return
}

// Logout...
func (c *Client) LogOut() (err error) {
	args := []interface{}{c.Token}
	res := struct {
		Status string `xmlrpc:"status"`
	}{}
	return c.Call("LogOut", args, &res)
}

// Build query parameters for hash-based movie search.
func (c *Client) hashSearchParams(path string, langs []string) (*[]interface{}, error) {
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
			fmt.Sprintf("%x", h),
			size,
			strings.Join(langs, ","),
		}},
	}
	return &params, nil
}

// Query parameters for TryUploadSubtitles
func (c *Client) hasSubtitlesParams(subs []Subtitle) *[]interface{} {
	// Convert subs param to map[string]struct{...}, because OSDb.
	subMap := map[string]interface{}{}
	for i, s := range subs {
		key := "cd" + strconv.Itoa(i+1) // keys are cd1, cd2, ...
		param := map[string]string{
			"subhash":       s.SubHash,
			"subfilename":   s.SubFileName,
			"moviehash":     s.MovieHash,
			"moviebytesize": s.MovieByteSize,
			"moviefilename": s.MovieFileName,
		}
		subMap[key] = param
	}

	return &[]interface{}{c.Token, subMap}
}
