// package git provides wrappers around basic git operations
package git

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Repository basics
type Repository struct {
	Name string
	Path string
	Type string // https, local, ssh
}

var tempdir = filepath.Join(os.TempDir(), "gits")
var tarpath = "/archive/master.tar.gz" // url path, not file path
var errNotRepository = fmt.Errorf("not a git repository")

// ErrNotSupported catch all
var ErrNotSupported = fmt.Errorf("feature not supported")
// Clone a repository from source to destination.
// If destination is an empty string, it will behave like 'git clone <repo>'
func Clone(repo, destination string) (err error) {
	r, err := ParseRepository(repo)
	if err != nil {
		return err
	}
	if destination == "" {
		destination = r.Name
	} else {
		var files []os.FileInfo
		files, err = ioutil.ReadDir(destination)
		if err != nil {
			if !strings.Contains(err.Error(), "no such") {
				return err
			}
		}
		if len(files) != 0 {
			return fmt.Errorf("%s is not empty", destination)
		}
	}

	println("Cloning into:", destination)

	switch r.Type {
	case "https":
		return clone(r, destination)
	default:
		return fmt.Errorf("only supporting https for now")
	}
}

func (r Repository) URL() string {
	var base, namespace string
	u, err := url.Parse(r.Path)
	if err != nil {
		return ""
	}
	base = u.Hostname()
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) != 2 {
		fmt.Fprintf(os.Stderr, "incorrect path: %v %s\n", len(parts), u.Path)
		return ""
	}
	namespace = parts[0]
	return fmt.Sprint("https://" + base + "/" + namespace + "/" + r.Name + tarpath)
}
func clone(r Repository, destination string) error {
	switch r.Type {
	case "https":
		if strings.HasPrefix(r.Path, "https://github.com/") {
			return clonehttps(r, destination)
		}
		println("trying unsupported source:", r.Path)
		return clonehttps(r, destination)
	default:
		return fmt.Errorf("%q type not supported", r.Type)
	}
}

func clonehttps(r Repository, destination string) (err error) {
	req, err := http.NewRequest("GET", r.URL(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "gitlike")
	println(req.URL.String())
	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	body, _ := ioutil.ReadAll(res.Body)

	defer res.Body.Close()
	file := filepath.Join(tempdir, r.Name+".tar.gz")
	println("Saving tarball:", file)

	out, err := os.Create(file)
	if err != nil {
		return err
	}
	n, err := out.Write(body)
	if err != nil {
		return err
	}
	if n != len(body) {
		return fmt.Errorf("Wanted to write %v bytes, only wrote %v", len(body), n)
	}
	err = untar(file)
	if err != nil {
		return err
	}
	//mv files
	err = intermove(filepath.Join(tempdir, r.Name+"-master"), destination)
	if err != nil {
		return err
	}

	println(r.Name, "cloned to", destination)
	return nil
}

// untar into /tmp/gits/project-master
func untar(file string) (err error) {
	os.MkdirAll(tempdir, 1)
	cmd := exec.Command("tar", "-v", "-C", tempdir, "-x", "-z", "-f", file) // expanded for busybox tar
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

// cant just os.Rename, so adding 'mv' as a dependency for now
func intermove(source, destination string) error {
	cmd := exec.Command("mv", "-v", source, destination)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()

}

// ParseRepository reads
func ParseRepository(repo string) (r Repository, err error) {
	switch {
	case strings.HasPrefix(repo, "http://"):
		r.Name = path.Base(repo)
		r.Path = repo
		r.Type = "http"
		return r, ErrNotSupported
	case strings.HasPrefix(repo, "https://"):
		r.Name = path.Base(repo)
		r.Name = strings.TrimSuffix(path.Base(repo), ".git")
		r.Path = repo
		r.Type = "https"
		return r, nil
	case strings.HasPrefix(repo, "git@"):
		r.Name = path.Base(repo)
		r.Path = repo
		r.Type = "ssh"
		return r, ErrNotSupported
	default: // local clone
		stat, err := os.Stat(repo)
		if err != nil {
			if err.Error() == "stat "+repo+": no such file or directory" {

				return r, fmt.Errorf("%q does not exist", repo)
			}
			return r, err
		}
		if !stat.IsDir() {
			return r, fmt.Errorf("%q is not a directory", repo)
		}

		stat, err = os.Stat(repo + "/.git")
		if err != nil {
			if strings.Contains(err.Error(), "no such") {
				return r, errNotRepository
			}
			return r, err
		}

		r.Name = path.Base(repo)
		r.Path = repo
		r.Type = "local"
		return r, nil
	}
}
