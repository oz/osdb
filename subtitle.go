package osdb

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/text/encoding"
)

// A Subtitle with its many OSDB attributes...
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
	MovieFileName      string `xmlrpc:"MovieName"`
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
	SubEncoding        string `xmlrpc:"SubEncoding"`
	SubtitlesLink      string `xmlrpc:"SubtitlesLink"`
	UserID             string `xmlrpc:"UserID"`
	UserNickName       string `xmlrpc:"UserNickName"`
	UserRank           string `xmlrpc:"UserRank"`
	ZipDownloadLink    string `xmlrpc:"ZipDownloadLink"`
}

// Subtitles is a collection of subtitles.
type Subtitles []Subtitle

// Best finds the best subsitle in a Subtitles collection. Of course
// "best" is hardly an absolute concept: here, we just take the first
// that OSDB returned.
func (subs Subtitles) Best() *Subtitle {
	if len(subs) > 0 {
		return &subs[0]
	}
	return nil
}

// SubtitleFile contains file data as returned by OSDB's API, that is to
// say: gzip-ped and base64-encoded text.
type SubtitleFile struct {
	ID       string `xmlrpc:"idsubtitlefile"`
	Data     string `xmlrpc:"data"`
	Encoding encoding.Encoding
	reader   io.ReadCloser
}

// Reader interface for SubtitleFile. Subtitle's contents are
// decompressed, and usually encoded to UTF-8: if encoding info is
// missing, no re-encoding is done.
func (sf *SubtitleFile) Reader() (r io.ReadCloser, err error) {
	if sf.reader != nil {
		return sf.reader, err
	}

	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(sf.Data))
	gzReader, err := gzip.NewReader(dec)
	if err != nil {
		return nil, err
	}

	if sf.Encoding == nil {
		sf.reader = gzReader
	} else {
		sf.reader = newCloseableReader(
			sf.Encoding.NewDecoder().Reader(gzReader),
			gzReader.Close,
		)
	}

	return sf.reader, nil
}

// NewSubtitleWithFile builds a Subtitle for a file, intended to be used with for osdb.HasSubtitles()
func NewSubtitleWithFile(movieFile string, subFile string) (s Subtitle, err error) {
	s.SubFileName = path.Base(subFile)
	// Compute md5 sum
	subIO, err := os.Open(subFile)
	if err != nil {
		return
	}
	defer subIO.Close()
	h := md5.New()
	_, err = io.Copy(h, subIO)
	if err != nil {
		return
	}
	s.SubHash = fmt.Sprintf("%x", h.Sum(nil))

	// Movie filename, byte-size, & hash.
	s.MovieFileName = path.Base(movieFile)
	movieIO, err := os.Open(movieFile)
	if err != nil {
		return
	}
	defer movieIO.Close()
	stat, err := movieIO.Stat()
	if err != nil {
		return
	}
	s.MovieByteSize = strconv.FormatInt(stat.Size(), 10)
	movieHash, err := HashFile(movieIO)
	if err != nil {
		return
	}
	s.MovieHash = fmt.Sprintf("%x", movieHash)
	return
}

// Convert Subtitle to a map[string]string{}, because OSDB requires a
// specific structure to match subtitles when uploading (or trying to).
func (subs *Subtitles) toUploadParams() (map[string]interface{}, error) {
	subMap := map[string]interface{}{}
	for i, s := range *subs {
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

	return subMap, nil
}

// Implement io.ReadCloser by wrapping io.Reader
type closeableReader struct {
	io.Reader
	close func() error
}

// Close the reader by calling a preset close function
func (c *closeableReader) Close() error {
	return c.close()
}

// Create a ReadCloser which will read from r and call close() upon closing
func newCloseableReader(r io.Reader, close func() error) io.ReadCloser {
	return &closeableReader{r, close}
}
