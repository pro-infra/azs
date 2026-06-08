package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	GITHUB_OWNER = "pro-infra"
	GITHUB_REPO  = "azs"
)

type semver struct {
	maj, min, pat int
}

type GitHubTagResponse struct {
	Ref string `json:"ref"`
	Url string `json:"url"`
}

var versionExp = regexp.MustCompile(`^refs/tags/v[0-9]+.[0-9]+(.[0-9]+)?$`)
var versionNumExp = regexp.MustCompile(`[0-9]+`)

func versionFromString(str string) semver {
	if !versionExp.MatchString(str) {
		return semver{}
	}

	s := versionNumExp.FindAllString(str, -1)
	if len(s) < 2 {
		return semver{}
	}

	var ver semver
	maj, err := strconv.Atoi(s[0])
	if err != nil {
		return semver{}
	}
	ver.maj = maj

	min, err := strconv.Atoi(s[1])
	if err != nil {
		return semver{}
	}
	ver.min = min

	if len(s) == 3 {
		pat, err := strconv.Atoi(s[2])
		if err != nil {
			return semver{}
		}
		ver.pat = pat
	}
	return ver
}

func (v semver) eq(v2 semver) bool {
	return v.maj == v2.maj && v.min == v2.min && v.pat == v2.pat
}

func (v semver) gt(v2 semver) bool {
	if v.maj != v2.maj {
		return v.maj > v2.maj
	}
	if v.min != v2.min {
		return v.min > v2.min
	}
	return v.pat > v2.pat
}

func (v semver) ge(v2 semver) bool {
	return v.eq(v2) || v.gt(v2)
}

func (v semver) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.maj, v.min, v.pat)
}

func updateAzs(dryRun bool) {
	versions, err := getAvailableVersions(GITHUB_OWNER, GITHUB_REPO)
	if err != nil {
		log.Fatalln(err)
	}

	if len(versions) == 0 {
		log.Println("No versions found")
		return
	}

	max := versions[0]
	for _, v := range versions {
		if v.gt(max) {
			max = v
		}
	}
	current := versionFromString(version)
	if current.ge(max) {
		log.Println("Newest version is already installed")
		return
	}
	log.Println("Update needed")

	filename, err := os.Executable()
	if err != nil {
		panic(err)
	}
	log.Printf("Update executable: %s\n", filename)

	if err = checkWriteProtection(filename); err != nil {
		log.Fatalln(err)
	}

	ext := ""
	if goOS == "windows" {
		ext = ".exe"
	}
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/azs.%s_%s%s", GITHUB_OWNER, GITHUB_REPO, max.String(), goOS, goArch, ext)
	if dryRun {
		log.Println("Would download", url, "to", filename)
	} else {
		if err = downloadFile(url, filename); err != nil {
			log.Fatalln(err)
		}
	}
	log.Println("success")
}

func getAvailableVersions(owner, repo string) ([]semver, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/tags", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{"Accept": {"application/json"}}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tags []GitHubTagResponse
	if err = json.Unmarshal(body, &tags); err != nil {
		return nil, err
	}

	return parseVersions(tags), nil
}

func parseVersions(tags []GitHubTagResponse) []semver {
	versions := make([]semver, 0, len(tags))
	for _, tag := range tags {
		if versionExp.MatchString(tag.Ref) {
			log.Println("Found version", tag.Ref)
			versions = append(versions, versionFromString(tag.Ref))
		}
	}
	return versions
}

func checkWriteProtection(filename string) error {
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}
	if info.Mode().Perm()&0200 == 0 {
		return errors.New("can not update - file is write protected")
	}
	return nil
}

func downloadFile(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(filepath.Dir(filename), "azs-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err = io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err = tmp.Chmod(0755); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	tmp.Close()

	if err = os.Rename(tmpName, filename); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
