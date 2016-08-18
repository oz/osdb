package osdb

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
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
	subFilePath        string
}

func (s *Subtitle) toUploadParams() map[string]string {
	return map[string]string{
		"subhash":       s.SubHash,
		"subfilename":   s.SubFileName,
		"moviehash":     s.MovieHash,
		"moviebytesize": s.MovieByteSize,
		"moviefilename": s.MovieFileName,
	}
}

func (s *Subtitle) encodeFile() (string, error) {
	fh, err := os.Open(s.subFilePath)
	if err != nil {
		return "", err
	}
	defer fh.Close()
	dest := bytes.NewBuffer([]byte{})
	gzWriter := gzip.NewWriter(dest)
	enc := base64.NewEncoder(base64.StdEncoding, gzWriter)
	_, err = io.Copy(enc, fh)
	if err != nil {
		return "", err
	}
	// XXX DEBUG
	fmt.Println("upload content size:", dest.Len())
	return dest.String(), nil
}

// Subtitles is a collection of subtitles.
type Subtitles []Subtitle

// ByDownloads implements sort interface for Subtitles, by download count.
type ByDownloads Subtitles

func (s ByDownloads) Len() int      { return len(s) }
func (s ByDownloads) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByDownloads) Less(i, j int) bool {
	iCnt, err := strconv.Atoi(s[i].SubDownloadsCnt)
	if err != nil {
		return false
	}
	jCnt, err := strconv.Atoi(s[j].SubDownloadsCnt)
	if err != nil {
		return true
	}
	return iCnt > jCnt
}

// Best finds the best subsitle in a Subtitles collection. Of course
// "best" is hardly an absolute concept: here, we just take the most
// downloaded file.
func (subs Subtitles) Best() *Subtitle {
	if len(subs) > 0 {
		sort.Sort(ByDownloads(subs))
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

// NewSubtitles builds a Subtitles from a movie path and a slice of
// subtitles paths. Intended to be used with for osdb.HasSubtitles() and
// osdb.UploadSubtitles().
func NewSubtitles(moviePath string, subPaths []string) (Subtitles, error) {
	subs := Subtitles{}
	for _, subPath := range subPaths {
		sub, err := NewSubtitle(moviePath, subPath)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// NewSubtitle builds a Subtitle from a movie and subtitle file path.
// Intended to be used with for osdb.HasSubtitles() and
// osdb.UploadSubtitles().
func NewSubtitle(moviePath string, subPath string) (s Subtitle, err error) {
	s.subFilePath = subPath
	s.SubFileName = path.Base(subPath)
	// Compute md5 sum
	subIO, err := os.Open(subPath)
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
	s.MovieFileName = path.Base(moviePath)
	movieIO, err := os.Open(moviePath)
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

// Serialize Subtitle to OSDB's XMLRPC params when trying to upload.
func (subs *Subtitles) toTryUploadParams() (map[string]interface{}, error) {
	subMap := map[string]interface{}{}
	for i, s := range *subs {
		key := "cd" + strconv.Itoa(i+1) // keys are cd1, cd2, ...
		subMap[key] = s.toUploadParams()
	}

	return subMap, nil
}

// Serialize Subtitle to OSDB's XMLRPC params when uploading.
func (subs *Subtitles) toUploadParams() (map[string]interface{}, error) {
	subMap := map[string]interface{}{}
	for i, s := range *subs {
		key := "cd" + strconv.Itoa(i+1) // keys are cd1, cd2, ...
		param := s.toUploadParams()
		encoded, err := s.encodeFile()
		if err != nil {
			return nil, err
		}
		param["subcontent"] = encoded
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
