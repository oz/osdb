![travis](https://api.travis-ci.org/oz/osdb.png?branch=master)

This is a Go client library for [OpenSubtitles](http://opensubtitles.org/).


The API has not reached version `0.1` yet, and is therefore subject to change.
Nonetheless, you are welcome to check it out, or participate.

 * Install with `go get -d github.com/oz/osdb`,
 * and `import "github.com/oz/osdb"` to use.

# Examples

## Hashing a file

```go
hash, err := osdb.Hash("somefile.avi")
if err != nil {
	// ...
}
fmt.Println("hash: %x\n", hash)
```

## Searching subtitles (file based)

```go
path := "/path/to/movie.avi"
languages := []string{"eng"}

// Hash file, then search.
res, err := osdb.FileSearch(path, languages)
if err != nil {
	// ...
}

for _, sub := range res {
	fmt.Printf("Found %s subtitles file \"%s\" at %s\n",
		sub.LanguageName, sub.SubFileName, sub.ZipDownloadLink)
}
```

## Downloading subtitles

Let's say you just made a search, for example using `osdb.FileSearch()`, and as
the API provided a few results, you decide to pick one for download:

```go
subs, err := osdb.FileSearch(...)

// Download subtitle file, and write to disk using subs[0].SubFileName
if err := subs[0].Download(); err != nil {
	// ...
}

// Alternatively, use the filename of your choice:
if err := subs[0].DownloadTo("safer-name.srt"); err != nil {
	// ...
}
```

## Getting a user session token 

By default, the library operates with the anonymous user. If you need to login
with a specific account, use the following:

```go
token, err := osdb.Login("user", "password", "language")
if err != nil {
	// ...
}
osdb.Token = token // the library will now operate with this Token.

```

# License

BSD, see the LICENSE file.
