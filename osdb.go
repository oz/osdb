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

	"github.com/mattn/go-xmlrpc"
)

const (
	ChunkSize     = 65536 // 64k
	Server        = "http://api.opensubtitles.org/xml-rpc"
	SearchLimit   = 100
	SuccessStatus = "200 OK"
)

var (
	UserAgent = "SubDownloader 2.0.10" // FIXME register a user-agent, one day.
	Token     = ""
)

type Result struct {
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
func SearchForFile(path string, langs []string) ([]Result, error) {
	// Login anonymously if no token has been set yet.
	if Token == "" {
		tok, err := Login("", "", "")
		Token = tok
		if err != nil {
			return nil, err
		}
	}

	// Get file size
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	size := fi.Size()

	// Get file hash
	h, err := Hash(path)
	if err != nil {
		return nil, err
	}

	// Query OSDB....
	res, err := xmlrpc.Call(Server,
		"SearchSubtitles",
		Token,
		[1]map[string]interface{}{ // go-xmlrpc panics on slices...
			map[string]interface{}{
				"moviehash":     fmt.Sprintf("%x", h),
				"moviebytesize": size,
				"sublanguageid": strings.Join(langs, ","),
			},
		},
		map[string]int{"limit": SearchLimit},
	)
	if err != nil {
		return nil, err
	}
	return parseSearchResults(&res)
}

// Untangle the XMLRPC mess, yay.
func parseSearchResults(resp *interface{}) ([]Result, error) {
	res := (*resp).(xmlrpc.Struct)
	if res["status"] != SuccessStatus {
		return nil, fmt.Errorf("Search error: %v", res["status"])
	}

	count := len(res["data"].(xmlrpc.Array))
	results := make([]Result, count)

	// FIXME debug mattn/go-xmlrpc, then
	// FIXME build a slice of Result
	//for _, raw := range res["data"].(xmlrpc.Struct) { }

	return results, nil
}

// Login to the API. Return a token
func Login(user string, pass string, lang string) (string, error) {
	res, err := xmlrpc.Call(Server, "LogIn", user, pass, lang, UserAgent)
	if err != nil {
		return "", err
	}
	data := res.(xmlrpc.Struct)
	if data["status"] != SuccessStatus {
		return "", fmt.Errorf("login error: %s", data["status"])
	}
	return data["token"].(string), nil
}

// Logout
func Logout(tok string) (err error) {
	_, err = xmlrpc.Call(Server, "LogOut", tok)
	return
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
