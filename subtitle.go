package osdb

import (
	"compress/gzip"
	"encoding/base64"
	"strings"
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
	SubtitlesLink      string `xmlrpc:"SubtitlesLink"`
	UserID             string `xmlrpc:"UserID"`
	UserNickName       string `xmlrpc:"UserNickName"`
	UserRank           string `xmlrpc:"UserRank"`
	ZipDownloadLink    string `xmlrpc:"ZipDownloadLink"`
}

// SubtitleFile contains file data as returned by OSDB's API, that is to
// say: gzip-ped and base64-encoded text.
type SubtitleFile struct {
	Id     string `xmlrpc:"idsubtitlefile"`
	Data   string `xmlrpc:"data"`
	reader *gzip.Reader
}

// A Reader for the subtitle file contents (decoded, and decompressed).
func (sf *SubtitleFile) Reader() (r *gzip.Reader, err error) {
	if sf.reader != nil {
		return sf.reader, err
	}

	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(sf.Data))
	sf.reader, err = gzip.NewReader(dec)

	return sf.reader, err
}
