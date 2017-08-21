package acoustid

import (
	"os"
	"path/filepath"
	"strings"

	id3 "github.com/mikkyang/id3-go"
)

type match struct {
	filename string
	result   Result
	album    string
}

type tagInfo struct {
	Artist string
	Title  string
	Album  string
}

func (ti *tagInfo) setOn(path string) {
	file, err := id3.Open(path)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	file.SetArtist(ti.Artist)
	file.SetTitle(ti.Title)

	if ti.Album != "" {
		file.SetAlbum(ti.Album)
	}
}

func TagDir(dir string) {
	filepath.Walk(dir, tagFile)
}

func tagFile(file string, info os.FileInfo, err error) error {
	if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".mp3") {

		//Get fingerprint
		fp := NewFingerprint(file)

		//Get acoustic Id
		resp := MakeAcoustIDRequest(fp)

		//Set ID3 tag
		SetID3(resp, file, info)

	}
	return nil
}

func MakeAcoustIDRequest(fp Fingerprint) AcoustIDResponse {
	apiKey := os.Getenv("ACOUSTID_API_KEY")
	if apiKey == "" {
		panic("Acoustid api key not in env ( ACOUSTID_API_KEY )")
	}

	request := AcoustIDRequest{
		Fingerprint: fp.fingerprint,
		Duration:    fp.duration,
		ApiKey:      apiKey,
		Metadata:    "recordings+releasegroups+compress",
	}

	return request.Do()
}

func SetID3(resp AcoustIDResponse, file string, info os.FileInfo) {
	// invalid response?
	if resp.Status != "ok" || len(resp.Results) == 0 {
		return
	}

	matches := []match{}
	for _, result := range resp.Results {
		if len(result.Recordings) == 0 || result.Score < 0.7 {
			continue
		}

		firstRec := result.Recordings[0]

		if firstRec.Title == "" || len(firstRec.Artists) == 0 {
			continue
		}

		if firstRec.Artists[0].Name == "" {
			continue
		}

		var bestRg string
	Outer:
		for _, rg := range firstRec.ReleaseGroups {
			// ignore singles
			if rg.Type != "Album" {
				continue
			}

			secondaries := rg.SecondaryTypes
			// ignore compilations and live releases
			for _, st := range secondaries {
				if st == "Compilation" || st == "Live" {
					continue Outer
				}
			}

			bestRg = rg.Title
			break
		}

		m := match{
			result:   result,
			filename: info.Name(),
			album:    bestRg,
		}
		matches = append(matches, m)
	}

	// no matches
	if len(matches) <= 0 {
		return
	}

	var maxScore float64 = -999
	var bestMatch match
	for _, m := range matches {
		if m.result.Score > maxScore {
			maxScore = m.result.Score
			bestMatch = m
		}
	}

	setTagsFromMatch(file, bestMatch)
}

func setTagsFromMatch(file string, m match) {
	rec := m.result.Recordings[0]

	tags := tagInfo{
		Artist: rec.Artists[0].Name,
		Title:  rec.Title,
		Album:  m.album,
	}

	tags.setOn(file)
}
