// Github database
package ghdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	// "sync"
	"sync/atomic"
	"time"

	"github.com/GlobeMC/mcla"
)

const syntaxVersion = 0 // 0 means dev

type UnsupportSyntaxErr struct {
	Version int
}

func (e *UnsupportSyntaxErr) Error() string {
	return fmt.Sprintf("MCLA-DB syntax version %d is not supported, please update the application", e.Version)
}

type versionData struct {
	Major         int `json:"major"`
	Minor         int `json:"minor"`
	Patch         int `json:"patch"`
	ErrorIncId    int `json:"errorIncId"`
	SolutionIncId int `json:"solutionIncId"`
}

type ErrDB struct {
	Fetch func(path string) (io.ReadCloser, error)

	checking      atomic.Bool
	cachedVersion versionData
	lastCheck     time.Time
}

var _ mcla.ErrorDB = (*ErrDB)(nil)

func (db *ErrDB) fetch(subpaths ...string) (io.ReadCloser, error) {
	return db.Fetch(path.Join(subpaths...))
}

func (db *ErrDB) fetchGhDBVersion() (v versionData, err error) {
	var res io.ReadCloser
	if res, err = db.fetch("version.json"); err != nil {
		return
	}
	defer res.Close()
	if err = json.NewDecoder(res).Decode(&v); err != nil {
		return
	}
	if v.Major != syntaxVersion {
		err = &UnsupportSyntaxErr{v.Major}
		return
	}
	return
}

func (db *ErrDB) checkUpdate() error {
	if !db.checking.CompareAndSwap(false, true) {
		return nil
	}
	defer db.checking.Store(false)

	if !db.lastCheck.IsZero() && time.Since(db.lastCheck) <= time.Minute {
		return nil
	}

	return db.RefreshCache()
}

func (db *ErrDB) RefreshCache() (err error) {
	newVersion, err := db.fetchGhDBVersion()
	if err != nil {
		return
	}
	// TODO: do we really need caches?
	if newVersion.Major != db.cachedVersion.Major || newVersion.Minor != db.cachedVersion.Minor {
		// TODO: refresh all data
		db.cachedVersion = newVersion
	} else if newVersion.Patch != db.cachedVersion.Patch {
		// var wg sync.WaitGroup
		// wg.Add(newVersion.ErrorIncId - db.cachedVersion.ErrorIncId)
		// for i := db.cachedVersion.ErrorIncId + 1; i <= newVersion.ErrorIncId; i++ {
		// 	go func(i int) {
		// 		defer wg.Done()
		// 		db.GetErrorDesc(i) // refresh cache
		// 	}(i)
		// }
		// wg.Wait()
		db.cachedVersion.ErrorIncId = newVersion.ErrorIncId
		// wg.Add(newVersion.SolutionIncId - db.cachedVersion.SolutionIncId)
		// for i := db.cachedVersion.SolutionIncId + 1; i <= newVersion.SolutionIncId; i++ {
		// 	go func(i int) {
		// 		defer wg.Done()
		// 		db.GetSolution(i) // refresh cache
		// 	}(i)
		// }
		// wg.Wait()
		db.cachedVersion.SolutionIncId = newVersion.SolutionIncId
		db.cachedVersion.Patch = newVersion.Patch
	}

	db.lastCheck = time.Now()
	return
}

func (db *ErrDB) GetErrorDesc(id int) (desc *mcla.ErrorDesc, err error) {
	res, err := db.fetch("errors", fmt.Sprintf("%d.json", id))
	if err != nil {
		return
	}
	defer res.Close()
	desc = new(mcla.ErrorDesc)
	if err = json.NewDecoder(res).Decode(&desc); err != nil {
		return
	}
	return
}

func (db *ErrDB) ForEachErrors(callback func(*mcla.ErrorDesc) error) (err error) {
	db.checkUpdate()

	ctx, cancel := context.WithCancelCause(context.Background())
	resCh := make(chan *mcla.ErrorDesc, 2)

	for i := 1; i <= db.cachedVersion.ErrorIncId; i++ {
		go func(i int) {
			desc, err := db.GetErrorDesc(i)
			if err != nil {
				cancel(err)
				return
			}
			resCh <- desc
		}(i)
	}
	for i := 1; i <= db.cachedVersion.ErrorIncId; i++ {
		select {
		case desc := <-resCh:
			if err = callback(desc); err != nil {
				return
			}
		case <-ctx.Done():
			return context.Cause(ctx)
		}
	}
	return
}

func (db *ErrDB) GetSolution(id int) (sol *mcla.SolutionDesc, err error) {
	var res io.ReadCloser
	if res, err = db.fetch("solutions", fmt.Sprintf("%d.json", id)); err != nil {
		return
	}
	defer res.Close()
	sol = new(mcla.SolutionDesc)
	if err = json.NewDecoder(res).Decode(sol); err != nil {
		sol = nil
		return
	}
	return
}
