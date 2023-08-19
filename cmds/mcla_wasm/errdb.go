
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	. "github.com/kmcsr/mcla"
)

const syntaxVersion = 0 // 0 means dev

type UnsupportSyntaxErr struct {
	Version int
}

func (e *UnsupportSyntaxErr)Error()(string){
	return fmt.Sprintf("MCLA-DB syntax version %d is not supported, please update the application", e.Version)
}

type HTTPStatusErr struct {
	URL        string
	StatusCode int
}

func (e *HTTPStatusErr)Error()(string){
	return fmt.Sprintf("HTTP status code error: %d when getting %q", e.StatusCode, e.URL)
}

type versionDataT struct {
	Major         int `json:"major"`
	Minor         int `json:"minor"`
	Patch         int `json:"patch"`
	ErrorIncId    int `json:"errorIncId"`
	SolutionIncId int `json:"solutionIncId"`
}

type ghErrDB struct {
	Prefix string
	cachedVersion versionDataT
}

var _ ErrorDB = (*ghErrDB)(nil)

func (db *ghErrDB)fetch(subpaths ...string)(res *Response, err error){
	var path string
	if path, err = url.JoinPath(db.Prefix, subpaths...); err != nil {
		return
	}
	if res, err = fetch(path); err != nil {
		return
	}
	return
}

func (db *ghErrDB)fetchGhDBVersion()(v versionDataT, err error){
	var res *Response
	if res, err = db.fetch("version.json"); err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = &HTTPStatusErr{ res.Url, res.StatusCode }
		return
	}
	if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
		return
	}
	if v.Major != syntaxVersion {
		err = &UnsupportSyntaxErr{ v.Major }
		return
	}
	return
}

func (db *ghErrDB)getErrorDesc(id int)(desc *ErrorDesc, err error){
	if getStorageValue(fmt.Sprintf("gh.db.errs.%d", id), &desc) {
		return
	}
	res, err := db.fetch("errors", fmt.Sprintf("%d.json", id))
	if res.StatusCode != 200 {
		res.Body.Close()
		if res.StatusCode == 404 {
			return
		}
		return nil, &HTTPStatusErr{ res.Url, res.StatusCode }
	}
	if err = json.NewDecoder(res.Body).Decode(&desc); err != nil {
		return
	}
	return
}

func (db *ghErrDB)checkUpdate()(err error){
	newVersion, err := db.fetchGhDBVersion()
	if err != nil {
		return
	}
	// TODO: do we really need caches?
	if newVersion.Major != db.cachedVersion.Major || newVersion.Minor != db.cachedVersion.Minor {
		// TODO: refresh all data
		db.cachedVersion = newVersion
	}else if newVersion.Patch != db.cachedVersion.Patch {
		var wg sync.WaitGroup
		wg.Add(newVersion.ErrorIncId - db.cachedVersion.ErrorIncId)
		for i := db.cachedVersion.ErrorIncId + 1; i <= newVersion.ErrorIncId; i++ {
			go func(i int){
				defer wg.Done()
				var (
					res *Response
					err error
				)
				if res, err = db.fetch("errors", fmt.Sprintf("%d.json", i)); err != nil {
					return
				}
				stokey := fmt.Sprintf("gh.db.errs.%d", i)
				if res.StatusCode != 200 {
					res.Body.Close()
					setStorageValue(stokey, nil)
					return
				}
				var v *ErrorDesc
				e := json.NewDecoder(res.Body).Decode(&v)
				res.Body.Close()
				if e != nil {
					delStorageValue(stokey)
					return
				}
				setStorageValue(stokey, v)
			}(i)
		}
		wg.Wait()
		db.cachedVersion.ErrorIncId = newVersion.ErrorIncId
		for i := db.cachedVersion.SolutionIncId + 1; i <= newVersion.SolutionIncId; i++ {
			// var res *Response
			// if res, err = db.fetch("solutions", fmt.Sprintf("%d.json", i)); res.StatusCode != 200 {
			// 	res.Body.Close()
			// 	if res.StatusCode == 404 {
			// 		continue
			// 	}
			// 	return &HTTPStatusErr{ res.Url, res.StatusCode }
			// }
		}
		db.cachedVersion.SolutionIncId = newVersion.SolutionIncId
		db.cachedVersion.Patch = newVersion.Patch
	}
	setStorageValue("gh.db.version", db.cachedVersion)
	return
}

func (db *ghErrDB)ForEachErrors(callback func(*ErrorDesc)(error))(err error){
	if err = db.checkUpdate(); err != nil {
		return
	}
	for i := 1; i <= db.cachedVersion.ErrorIncId; i++ {
		var desc *ErrorDesc
		if desc, err = db.getErrorDesc(i); err != nil {
			return
		}
		if err = callback(desc); err != nil {
			return
		}
	}
	return
}

func (*ghErrDB)GetSolution(id int)(*SolutionDesc){
	panic("Unimplemented operation 'GetSolution'")
	return nil
}

var defaultErrDB = &ghErrDB{
	Prefix: "https://raw.githubusercontent.com/kmcsr/mcla-db-dev/main",
}

var defaultAnalyzer = &Analyzer{
	DB: defaultErrDB,
}
