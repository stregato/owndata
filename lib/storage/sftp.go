package storage

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"github.com/stregato/stash/lib/core"
	"golang.org/x/crypto/ssh"
)

// type SFTPConfig struct {
// 	Addr     string `json:"addr" yaml:"addr"`
// 	Username string `json:"username" yaml:"username"`
// 	Password string `json:"password" yaml:"password"`
// 	KeyPath  string `json:"keyPath" yaml:"keyPath"`
// 	Base     string `json:"base" yaml:"base"`
// }

type SFTP struct {
	c     *sftp.Client
	base  string
	id    string
	touch map[string]time.Time
}

// func ParseSFTPUrl(s string) (SFTPConfig, error) {
// 	u, err := url.Parse(s)
// 	if err != nil {
// 		return SFTPConfig{}, err
// 	}

// 	password, _ := u.User.Password()
// 	return SFTPConfig{
// 		Addr:     u.Host,
// 		Username: u.User.Username(),
// 		Password: password,
// 		Base:     u.Path,
// 	}, nil
// }

// func ToUrl(config SFTPConfig) string {
// 	return fmt.Sprintf("sftp://%s@%s/%s", config.Username, config.Addr, config.Base)
// }

// OpenSFTP create a new Exchanger. The url is in the format sftp://
func OpenSFTP(connectionUrl string) (Store, error) {
	u, err := url.Parse(connectionUrl)
	if err != nil {
		return nil, err
	}

	addr := u.Host
	if u.Port() == "" {
		addr = fmt.Sprintf("%s:22", addr)
	}

	params := u.Query()

	var id string
	var auth []ssh.AuthMethod

	password, hasPassword := u.User.Password()
	if hasPassword {
		auth = append(auth, ssh.Password(password))
	}

	if key := params.Get("k"); key != "" {
		pkey, err := base64.StdEncoding.DecodeString(key)
		if core.IsErr(err, "private key is invalid: %v") {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(pkey)
		if err != nil {
			return nil, fmt.Errorf("invalid key: %v", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	id = fmt.Sprintf("sftp://%s/%s", addr, u.Path)

	if len(auth) == 0 {
		return nil, fmt.Errorf("no auth method provided for sftp connection to %s", addr)
	}

	cc := &ssh.ClientConfig{
		User:            u.User.Username(),
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr, cc)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s in NewSFTP: %v", addr, err)
	}
	c, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("cannot create a sftp client for %s: %v", addr, err)
	}

	base := u.Path
	if base == "" {
		base = "/"
	}
	return &SFTP{c, base, id, map[string]time.Time{}}, nil
}

func (s *SFTP) ID() string {
	return s.id
}

func (s *SFTP) Read(name string, rang *Range, dest io.Writer, progress chan int64) error {
	f, err := s.c.Open(path.Join(s.base, name))
	if os.IsNotExist(err) || core.IsErr(err, "cannot open file on sftp server %v:%v", s) {
		return err
	}

	if rang == nil {
		_, err = io.Copy(dest, f)
	} else {
		left := rang.To - rang.From
		f.Seek(rang.From, 0)
		var b [4096]byte

		for left > 0 && err == nil {
			var sz int64
			if rang.To-rang.From > 4096 {
				sz = 4096
			} else {
				sz = rang.To - rang.From
			}
			n, err := f.Read(b[0:sz])
			dest.Write(b[0:n])
			left -= int64(n)
			if err != nil {
				break
			}
		}
	}
	if err != io.EOF && core.IsErr(err, "cannot read from %s/%s:%v", s, name) {
		return err
	}

	return nil
}

func (s *SFTP) Write(name string, source io.ReadSeeker, progress chan int64) error {
	name = path.Join(s.base, name)

	f, err := s.c.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
	if os.IsNotExist(err) {
		dir := path.Dir(name)
		s.c.MkdirAll(dir)
		f, err = s.c.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC)
	}
	if core.IsErr(err, "cannot create SFTP file '%s': %v", name) {
		return err
	}

	_, err = io.Copy(f, source)
	core.IsErr(err, "cannot write SFTP file '%s': %v", name)
	return err
}

func (s *SFTP) ReadDir(dir string, f Filter) ([]fs.FileInfo, error) {
	dir = path.Join(s.base, dir)
	ls, err := s.c.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var cnt int64
	var infos []fs.FileInfo
	for _, l := range ls {
		if matchFilter(l, f) {
			infos = append(infos, l)
			cnt++
		}
		if f.MaxResults > 0 && cnt >= f.MaxResults {
			break
		}
	}

	return infos, nil
}

func (s *SFTP) Stat(name string) (os.FileInfo, error) {
	return s.c.Stat(path.Join(s.base, name))
}

func (s *SFTP) Rename(old, new string) error {
	return s.c.Rename(path.Join(s.base, old), path.Join(s.base, new))
}

func (s *SFTP) Delete(name string) error {
	n := path.Join(s.base, name)
	stat, err := s.c.Stat(n)
	if core.IsErr(err, "cannot stat %s in Delete: %v", n) {
		return err
	}

	if stat.IsDir() {
		is, _ := s.c.ReadDir(n)
		for _, i := range is {
			err = s.Delete(path.Join(name, i.Name()))
			if err != nil {
				return err
			}
		}
	}
	err = s.c.Remove(n)
	core.IsErr(err, "cannot delete %s in Delete: %v", n)
	return nil
}

func (s *SFTP) Close() error {
	return s.c.Close()
}

func (s *SFTP) String() string {
	return s.id
}

// Describe implements Store.
func (*SFTP) Describe() Description {
	return Description{
		ReadCost:  0.000000001,
		WriteCost: 0.000000001,
	}
}
