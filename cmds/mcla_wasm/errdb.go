package main

import (
	"fmt"
	"io"
	"net/url"

	"github.com/GlobeMC/mcla"
	"github.com/GlobeMC/mcla/ghdb"
)

type HTTPStatusErr struct {
	URL        string
	StatusCode int
}

func (e *HTTPStatusErr) Error() string {
	return fmt.Sprintf("HTTP status code error: %d when getting %q", e.StatusCode, e.URL)
}

var ghRepoPrefix = "https://raw.githubusercontent.com/kmcsr/mcla-db-dev/main"

var defaultErrDB = &ghdb.ErrDB{
	Fetch: func(path string) (io.ReadCloser, error) {
		path, err := url.JoinPath(ghRepoPrefix, path)
		if err != nil {
			return nil, err
		}
		res, err := fetch(path)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			res.Body.Close()
			return nil, &HTTPStatusErr{res.Url, res.StatusCode}
		}
		return res.Body, nil
	},
}

var defaultAnalyzer = &mcla.Analyzer{
	DB: defaultErrDB,
}
