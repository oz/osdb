![travis](https://api.travis-ci.org/oz/osdb.png?branch=master)

This is a Go client library for [OpenSubtitles](http://opensubtitles.org/).


This lib has not reached version `0.1` yet, and its API will change in many
breaking ways.  But of course, you are welcome to check it out, and
participate. :)

To get started...

 * Install with `go get -d github.com/oz/osdb`,
 * import `"github.com/oz/osdb"`,
 * and try some of the examples.

# Examples

## Hashing a file

```go
hash, err := osdb.Hash("somefile.avi")
if err != nil {
	// ...
}
fmt.Println("hash: %x\n", hash)
```

## Searching subtitles

Subtitle search is (for now) entirely file based: you can not yet search by
movie name, or using an IMDB ID. So, in order to search subtitles, you must
have a movie file for the which we compute a hash, then used to search OSDB.

```go
path := "/path/to/movie.avi"
languages := []string{"eng"}

// Compute file hash, and use it to search...
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

Let's say you have just made a search, for example using `osdb.FileSearch()`,
and as the API provided a few results, you decide to pick one for download:

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

By default, the library operates with the anonymous user. If you ever need to
login with a specific account, use the following:

```go
token, err := osdb.Login("user", "password", "language")
if err != nil {
	// ...
}
osdb.Token = token // the library will now operate with this Token.

```

# Documentation

The generated documentation for this package is available at:
http://godoc.org/github.com/oz/osdb

If you have read OSDB's [developer documentation][osdb], you should notice that
you need to register an "official" user agent in order to use their API (meh).
By default this library uses their "test" agent: it is fine for tests, but it
is probably not what you would use in production. For registered applications,
change the `UserAgent` variable with:

```go
osdb.UserAgent = "My totally official agent"
```

# License

BSD, see the LICENSE file.

[osdb]: http://trac.opensubtitles.org/projects/opensubtitles
