![travis](https://api.travis-ci.org/oz/osdb.png?branch=master)

This is a Go client library for [OpenSubtitles](http://opensubtitles.org/).

The generated documentation for this package is available at:
http://godoc.org/github.com/oz/osdb

This lib has not reached version `0.1` yet, and its API will change in many
breaking ways.  But of course, you are welcome to check it out, and
participate. :)

# Getting started...

To get started...

 * Install with `go get -d github.com/oz/osdb`,
 * import `"github.com/oz/osdb"`,
 * and try some of the examples.

To use OpenSubtitles' API you need to allocate a client, and to login (even
anonymously) in order to receive a session token. Here is an example:

```go
package main

import "github.com/oz/osdb"

func main() {
	c, err := osdb.NewClient()
	if err != nil {
		// ...
	}

	// Anonymous login will set c.Token when successful
	if err = c.LogIn("", "", ""); err != nil {
		// ...
	}

	// etc.
}

```

# Basic examples

## Getting a user session token 

Although this library tries to be simple, to use OpenSubtitles' API you need to
login first so as to receive a session token: without it you will not be able
to call any API method.

```go
c, err := osdb.NewClient()
if err != nil {
	// ...
}

err := c.LogIn("user", "password", "language")
if err != nil {
	// ...
}
// c.Token is now set.
```

However, you do not need to register a user, to login anonymously, just leave
the `user` and `password` parameters blank:

```go
c.LogIn("", "", "")
```

## Searching subtitles

Subtitle search is (for now) entirely file based: you can not yet search by
movie name, or using an IMDB ID. So, in order to search subtitles, you *must*
have a movie file for the which we compute a hash, then used to search OSDB.

```go
path := "/path/to/movie.avi"
languages := []string{"eng"}

// Hash movie file, and search...
res, err := client.FileSearch(path, languages)
if err != nil {
	// ...
}

for _, sub := range res {
	fmt.Printf("Found %s subtitles file \"%s\" at %s\n",
		sub.LanguageName, sub.SubFileName, sub.ZipDownloadLink)
}
```

## Downloading subtitles

Let's say you have just made a search, for example using `FileSearch()`, and as
the API provided a few results, you decide to pick one for download:

```go
subs, err := c.FileSearch(...)

// Download subtitle file, and write to disk using subs[0].SubFileName
if err := c.Download(&subs[0]); err != nil {
	// ...
}

// Alternatively, use the filename of your choice:
if err := c.DownloadTo(&subs[0], "safer-name.srt"); err != nil {
	// ...
}
```

## Checking if a subtitle exists

Before trying to upload an allegedly "new" subtitles file to OSDB, you should
always check whether they already have it.

As some movies fit on more than one "CD" (remember those?), you will need to
create a slice of `Subtitle`, one per subtitle file:

```go
subs := []osdb.Subtitle{
		{
			SubHash:       subHash,       // md5 hash of subtitle file
			SubFileName:   subFileName,
			MovieHash:     movieHash,     // see osdb.Hash()
			MovieByteSize: movieByteSize, // careful, it's a string...
			MovieFileName: movieFileName,
		},
}
```

Then simply feed that to `HasSubtitles`, and you'll be done.

```go
found, err := c.HasSubtitles(subs)
if err != nil {
	// ...
}
```

## Hashing a file

OSDB uses a custom hash to identify movie files.

```go
hash, err := osdb.Hash("somefile.avi")
if err != nil {
	// ...
}
fmt.Println("hash: %x\n", hash)
```


# On user agents...

If you have read OSDB's [developer documentation][osdb], you should notice that
you need to register an "official" user agent in order to use their API (meh).
By default this library uses their "test" agent: fine for tests, but it is
probably not what you would use in production. For registered applications, you
can change the `UserAgent` variable with:

```go
c, err := osdb.NewClient()
if err != nil {
	// ...
}
c.UserAgent = "My custom user agent"
```

# TODO

[Full API coverage][apidocs]:

  - :ballot_box_with_check: LogIn
  - :ballot_box_with_check: LogOut
  - :ballot_box_with_check: NoOperation
  - :ballot_box_with_check: SearchSubtitles by hash
  - SearchSubtitles by IMDB ID
  - SearchToMail
  - :ballot_box_with_check: DownloadSubtitles
  - :ballot_box_with_check: TryUploadSubtitles
  - UploadSubtitles
  - SearchMoviesOnIMDB
  - GetIMDBMovieDetails
  - InsertMovie
  - ServerInfo
  - ReportWrongMovieHash
  - SubtitlesVote
  - AddComment
  - GetSubLanguages
  - DetectLanguage
  - GetAvailableTranslations
  - GetTranslation
  - AutoUpdate
  - CheckMovieHash
  - CheckSubHash

# License

BSD, see the LICENSE file.

[osdb]: http://trac.opensubtitles.org/projects/opensubtitles
[apidocs]: http://trac.opensubtitles.org/projects/opensubtitles/wiki/XmlRpcIntro#XML-RPCmethods
