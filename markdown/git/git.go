package git

import (
	"log"
	"time"

	gitv "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func GitStats(path string) (stats map[string]int64) {
	stats = map[string]int64{}
	r, err := gitv.PlainOpen(path)
	if err != nil {
		log.Println("open git error:" + path)
		return
	}

	// pull
	// wr, err := r.Worktree()
	// err = wr.Pull(&gitv.PullOptions{})

	until := time.Now()
	cIter, err := r.Log(&gitv.LogOptions{Until: &until})
	if err != nil {
		log.Println("git log error:" + path)
		return
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		//c.Committer.When
		s := c.Committer.When.Format("2006-01-02")
		stats[s] = stats[s] + 1
		return nil
	})
	if err != nil {
		log.Println("git log ForEach error:" + path)
		return
	}
	return
}
