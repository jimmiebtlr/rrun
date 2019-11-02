package remote

import (
	"time"

	"github.com/mlafeldt/chef-runner/rsync"
)

func Sync(addr, src, dst string) {
	client := &rsync.Client{
		Archive: true,
		//Delete:     true,  Until 2 way sync works better, stop deleting my progress
		Compress:   true,
		RemoteHost: addr,
		Options:    []string{"-r"},
	}

	for {
		client.Copy("jimmiebtlr_gmail_com@"+addr+":"+dst, src)
		time.Sleep(time.Second)
	}
}
