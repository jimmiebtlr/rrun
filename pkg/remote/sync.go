package remote

import (
	"time"

	"github.com/mlafeldt/chef-runner/rsync"
)

func Sync(user, addr, src, dst string) {
	client := &rsync.Client{
		Archive: true,
		//Delete:     true,  Until 2 way sync works better, stop deleting my progress
		Compress:   true,
		RemoteHost: addr,
		//Verbose:    true,
		Options: []string{"-r", "--exclude=.git"},
	}

	for {
		//fmt.Println("Running sync")
		client.Copy(user+"@"+addr+":"+dst+"/", src+"/")
		client.Copy(src+"/", user+"@"+addr+":"+dst+"/")
		time.Sleep(time.Second)
	}
}
