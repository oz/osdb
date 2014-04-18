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
	SuccessStatus    = "200 OK"
)

type Client struct {
	UserAgent string
	Token     string
	Login     string
	Password  string
	Language  string
	*xmlrpc.Client
}

// Search subtitles in `languages` for a file at `path`.
func (c *Client) FileSearch(path string, langs []string) ([]Subtitle, error) {
	// OSDB search params struct
	params, err := c.hashSearchParams(path, langs)
	if err != nil {
		return nil, err
	}
	res := struct {
		Data []Subtitle `xmlrpc:"data"`
	}{}

	if err = c.Call("SearchSubtitles", *params, &res); err != nil {
		return nil, err
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
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("No file match this subtitle ID")
	}

	// Save to disk.
	file := files[0]
	r, err := file.Reader()
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

// Checks wether a subtitle already exists in OSDB. The mandatory fields
// in the received Subtitle map are: subhash, subfilename, moviehash,
// moviebytesize, and moviefilename.
func (c *Client) HasSubtitles(subs []Subtitle) (bool, error) {
	// Convert subs param to map[string]Subtitle, because OSDb.
	subMap := map[string]Subtitle{}
	for i, s := range subs {
		key := "cd" + strconv.Itoa(i+1) // keys are cd1, cd2, ...
		subMap[key] = s
	}

	args := []interface{}{c.Token, &subMap}
	res := struct {
		Status string     `xmlrpc:"status"`
		Exists bool       `xmlrpc:"alreadyindb"`
		Data   []Subtitle `xmlrpc:"data"`
	}{}
	if err := c.Call("TryUploadSubtitles", args, &res); err != nil {
		return false, err
	}

	return res.Exists, nil
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

	if res.Status != SuccessStatus {
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
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	size := fi.Size()

	// File hash
	h, err := Hash(path)
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
