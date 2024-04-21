package storage

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/stregato/mio/lib/core"
)

const LockDir = ".lock"
const LockExpire = 10 * time.Second

func Lock(s Store, dir, lockType string, timeout time.Duration) (release chan bool, err error) {
	if timeout == 0 {
		return tryLock(s, dir, lockType)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return nil, nil
		case <-ticker.C:
			release, err = tryLock(s, dir, lockType)
			if err != nil {
				return nil, err
			}
			if release != nil {
				return release, nil
			}
		}
	}

}

func Unlock(release chan bool) {
	close(release)
}

func tryLock(s Store, dir, lockType string) (release chan bool, err error) {
	dir = path.Join(dir, LockDir)

	ls, err := s.ReadDir(dir, Filter{})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	lockName := fmt.Sprintf("%s-%d.lock", lockType, core.SnowID())
	lockPath := path.Join(dir, lockName)
	err = s.Write(lockPath, core.NewBytesReader(core.GenerateRandomBytes(8)), nil)
	if err != nil {
		return nil, err
	}
	time.Sleep(100 * time.Millisecond)
	stat, err := s.Stat(lockPath)
	if err != nil {
		return nil, err
	}

	oldest := stat
	for _, l := range ls {
		if strings.HasPrefix(l.Name(), lockType) {
			if l.ModTime().Add(LockExpire).Before(stat.ModTime()) {
				s.Delete(l.Name())
				continue
			}
			if l.ModTime().Before(oldest.ModTime()) || (l.ModTime().Equal(oldest.ModTime()) && l.Name() < oldest.Name()) {
				oldest = l
			}
		}
	}

	if oldest != stat {
		s.Delete(lockPath)
		return nil, nil
	}

	release = make(chan bool)
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.Write(lockPath, core.NewBytesReader(core.GenerateRandomBytes(8)), nil)
			case <-release:
				s.Delete(lockPath)
				return
			}
		}
	}()
	return release, nil
}
