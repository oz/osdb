/*
An API client for opensubtitles.org

This is a client for the OSDb protocol. Currently the package only allows movie
identification, and subtitles search.
*/
package osdb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/kolo/xmlrpc"
)

const (
	ChunkSize     = 65536 // 64k
	Server        = "http://api.opensubtitles.org/xml-rpc"
	SearchLimit   = 100
	SuccessStatus = "200 OK"
)

var (
	UserAgent = "OS Test User Agent" // FIXME register a user-agent, one day.
	Token     = ""
	Client, _ = xmlrpc.NewClient(Server, nil)
)

type Subtitle struct {
	IDMovie            string `xmlrpc:"IDMovie"`
	IDMovieImdb        string `xmlrpc:"IDMovieImdb"`
	IDSubMovieFile     string `xmlrpc:"IDSubMovieFile"`
	IDSubtitle         string `xmlrpc:"IDSubtitle"`
	IDSubtitleFile     string `xmlrpc:"IDSubtitleFile"`
	ISO639             string `xmlrpc:"ISO639"`
	LanguageName       string `xmlrpc:"LanguageName"`
	MatchedBy          string `xmlrpc:"MatchedBy"`
	MovieByteSize      string `xmlrpc:"MovieByteSize"`
	MovieFPS           string `xmlrpc:"MovieFPS"`
	MovieHash          string `xmlrpc:"MovieHash"`
	MovieImdbRating    string `xmlrpc:"MovieImdbRating"`
	MovieKind          string `xmlrpc:"MovieKind"`
	MovieName          string `xmlrpc:"MovieName"`
	MovieNameEng       string `xmlrpc:"MovieNameEng"`
	MovieReleaseName   string `xmlrpc:"MovieReleaseName"`
	MovieTimeMS        string `xmlrpc:"MovieTimeMS"`
	MovieYear          string `xmlrpc:"MovieYear"`
	QueryNumber        string `xmlrpc:"QueryNumber"`
	SeriesEpisode      string `xmlrpc:"SeriesEpisode"`
	SeriesIMDBParent   string `xmlrpc:"SeriesIMDBParent"`
	SeriesSeason       string `xmlrpc:"SeriesSeason"`
	SubActualCD        string `xmlrpc:"SubActualCD"`
	SubAddDate         string `xmlrpc:"SubAddDate"`
	SubAuthorComment   string `xmlrpc:"SubAuthorComment"`
	SubBad             string `xmlrpc:"SubBad"`
	SubComments        string `xmlrpc:"SubComments"`
	SubDownloadLink    string `xmlrpc:"SubDownloadLink"`
	SubDownloadsCnt    string `xmlrpc:"SubDownloadsCnt"`
	SubFeatured        string `xmlrpc:"SubFeatured"`
	SubFileName        string `xmlrpc:"SubFileName"`
	SubFormat          string `xmlrpc:"SubFormat"`
	SubHash            string `xmlrpc:"SubHash"`
	SubHD              string `xmlrpc:"SubHD"`
	SubHearingImpaired string `xmlrpc:"SubHearingImpaired"`
	SubLanguageID      string `xmlrpc:"SubLanguageID"`
	SubRating          string `xmlrpc:"SubRating"`
	SubSize            string `xmlrpc:"SubSize"`
	SubSumCD           string `xmlrpc:"SubSumCD"`
	SubtitlesLink      string `xmlrpc:"SubtitlesLink"`
	UserID             string `xmlrpc:"UserID"`
	UserNickName       string `xmlrpc:"UserNickName"`
	UserRank           string `xmlrpc:"UserRank"`
	ZipDownloadLink    string `xmlrpc:"ZipDownloadLink"`
}

// Generate a OSDB hash for a file
func Hash(path string) (hash uint64, err error) {
	// Check file size.
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	if fi.Size() < ChunkSize {
		return 0, fmt.Errorf("File is too small")
	}

	file, err := os.Open(path)
	if err != nil {
		return
	}

	// Read head and tail blocks.
	buf := make([]byte, ChunkSize*2)
	err = readChunk(file, 0, buf[:ChunkSize])
	if err != nil {
		return
	}
	err = readChunk(file, fi.Size()-ChunkSize, buf[ChunkSize:])
	if err != nil {
		return
	}

	// Convert to uint64, and sum.
	var nums [(ChunkSize * 2) / 8]uint64
	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, &nums)
	if err != nil {
		return 0, err
	}
	for _, num := range nums {
		hash += num
	}

	return hash + uint64(fi.Size()), nil
}

// Search subtitles in `languages` for a file at `path`.
func FileSearch(path string, langs []string) ([]Subtitle, error) {
	// Login anonymously if no token has been set yet.
	if Token == "" {
		tok, err := Login("", "", "")
		Token = tok
		if err != nil {
			return nil, err
		}
	}

	// Query OSDB....
	params, err := hashSearchParams(path, langs)
	if err != nil {
		return nil, err
	}
	res := struct {
		Data []Subtitle `xmlrpc:"data"`
	}{}
	if err = Client.Call("SearchSubtitles", params, &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// Build query parameters for hash-based movie search.
func hashSearchParams(path string, langs []string) (interface{}, error) {
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
	return []interface{}{
		Token,
		[]struct {
			Hash  string `xmlrpc:"moviehash"`
			Size  int64  `xmlrpc:"moviebytesize"`
			Langs string `xmlrpc:"sublanguageid"`
		}{{
			fmt.Sprintf("%x", h),
			size,
			strings.Join(langs, ","),
		}},
	}, nil
}

// Login to the API, and return a session token.
func Login(user string, pass string, lang string) (tok string, err error) {
	args := []interface{}{user, pass, lang, UserAgent}
	res := struct {
		Status string `xmlrpc:"status"`
		Token  string `xmlrpc:"token"`
	}{}
	if err = Client.Call("LogIn", args, &res); err != nil {
		return
	}

	if res.Status != SuccessStatus {
		return tok, fmt.Errorf("login error: %s", res.Status)
	}
	tok = res.Token
	return
}

// Logout
func Logout(tok string) (err error) {
	args := []interface{}{tok}
	res := struct {
		Status string `xmlrpc:"status"`
	}{}
	return Client.Call("LogOut", args, &res)
}

// Read a chunk of a file at `offset` so as to fill `buf`.
func readChunk(file *os.File, offset int64, buf []byte) (err error) {
	n, err := file.ReadAt(buf, offset)
	if err != nil {
		return
	}
	if n != ChunkSize {
		return fmt.Errorf("Invalid read", n)
	}
	return
}
