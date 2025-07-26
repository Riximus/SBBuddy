package api

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"math/rand"
	"strings"
	"time"
)

//go:embed major_stations.txt
var majorStationsFS embed.FS

var majorStations []string

func init() {
	data, err := majorStationsFS.ReadFile("major_stations.txt")
	if err != nil {
		// fall back to a small set if embedding fails
		majorStations = []string{"Basel SBB", "Bern", "Zürich HB", "Lausanne"}
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			majorStations = append(majorStations, line)
		}
	}
	if len(majorStations) == 0 {
		majorStations = []string{"Basel SBB", "Bern", "Zürich HB", "Lausanne"}
	}
}

func randomStationsFromList(r *rand.Rand, via int) ([]string, error) {
	if via < 0 || via+2 > len(majorStations) {
		return nil, errors.New("invalid via count")
	}
	idxs := r.Perm(len(majorStations))[:via+2]
	res := make([]string, via+2)
	for i, id := range idxs {
		res[i] = majorStations[id]
	}
	return res, nil
}

// RandomStations returns a slice of station names representing a random
// connection. The first and last elements are the origin and destination. The
// number of via stations is determined by the via parameter.
func RandomStations(via int) ([]string, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return randomStationsFromList(r, via)
}

// randomStationsExclude selects a given number of stations that are not present
// in the exclude map. The returned stations are unique.
func randomStationsExclude(r *rand.Rand, exclude map[string]struct{}, count int) ([]string, error) {
	available := make([]string, 0, len(majorStations))
	for _, s := range majorStations {
		if _, skip := exclude[strings.ToLower(s)]; !skip {
			available = append(available, s)
		}
	}
	if count < 0 || count > len(available) {
		return nil, errors.New("invalid count")
	}
	idxs := r.Perm(len(available))[:count]
	res := make([]string, count)
	for i, id := range idxs {
		res[i] = available[id]
	}
	return res, nil
}

// RandomStationsExclude returns randomly selected stations excluding the given
// list. The count parameter specifies how many stations should be returned.
func RandomStationsExclude(count int, exclude []string) ([]string, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	m := make(map[string]struct{}, len(exclude))
	for _, e := range exclude {
		m[strings.ToLower(e)] = struct{}{}
	}
	return randomStationsExclude(r, m, count)
}
