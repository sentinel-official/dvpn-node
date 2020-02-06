package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	hub "github.com/sentinel-official/hub/types"
)

const (
	radius = 6378.137
	p      = math.Pi / 180.0
)

func distance(lat1, lon1, lat2, lon2 float64) float64 {
	x1, y1, x2, y2 := p*lat1, p*lon1, p*lat2, p*lon2
	z := math.Cos(y2-y1)*
		math.Cos(x2)*math.Cos(x1) +
		math.Sin(x2)*math.Sin(x1)

	return radius * math.Acos(z)
}

func fetchLocation() (lat, lon float64, err error) {
	resp, err := http.Get("http://ip-api.com/json")
	if err != nil {
		return 0, 0, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var res struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, err
	}

	return res.Lat, res.Lon, nil
}

type server struct {
	URL string  `xml:"url,attr"`
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`

	distance float64
}

func fetchServers() ([]server, error) {
	resp, err := http.Get("https://www.speedtest.net/speedtest-servers-static.php")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	var res struct {
		Servers []server `xml:"servers>server"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Servers, nil
}

func uploadSpeed(s *server, load, size int) (v int64, err error) {
	c, wg := http.Client{Timeout: 5 * time.Second}, &sync.WaitGroup{}
	_url := s.URL

	data := url.Values{}
	data.Add("content", strings.Repeat("0", size))

	start := time.Now()
	for i := 0; i < load; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, _err := c.PostForm(_url, data)
			if _err != nil || resp.StatusCode != http.StatusOK {
				err = errors.Errorf("something went wrong")
				return
			}

			defer func() {
				if _err := resp.Body.Close(); _err != nil {
					panic(_err)
				}
			}()

			_, _ = ioutil.ReadAll(resp.Body)
		}()
	}

	wg.Wait()
	if err != nil {
		return 0, err
	}

	return int64(float64(load*size) / time.Since(start).Seconds()), nil
}

func downloadSpeed(s *server, load, size int) (v int64, err error) {
	c, wg := http.Client{Timeout: 5 * time.Second}, &sync.WaitGroup{}
	_url := fmt.Sprintf("%s/random%dx%d.jpg", strings.Split(s.URL, "/upload")[0], size, size)

	start := time.Now()
	for i := 0; i < load; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, _err := c.Get(_url)
			if _err != nil || resp.StatusCode != http.StatusOK {
				err = errors.Errorf("something went wrong")
				return
			}

			defer func() {
				if _err := resp.Body.Close(); _err != nil {
					panic(_err)
				}
			}()

			_, _ = ioutil.ReadAll(resp.Body)
		}()
	}

	wg.Wait()
	if err != nil {
		return 0, err
	}

	return int64(float64(load*2*size*size) / time.Since(start).Seconds()), nil
}

func InternetSpeed() (hub.Bandwidth, error) {
	lat, lon, err := fetchLocation()
	if err != nil {
		return hub.NewBandwidthFromInt64(0, 0), err
	}

	servers, err := fetchServers()
	if err != nil {
		return hub.NewBandwidthFromInt64(0, 0), err
	}

	for i := range servers {
		s := &servers[i]
		s.distance = distance(lat, lon, s.Lat, s.Lon)
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].distance < servers[j].distance
	})

	var upload, download int64

	for i := range servers[:8] {
		upload, err = uploadSpeed(&servers[i], 8, 1500)
		if err == nil {
			break
		}
	}

	for i := range servers[:8] {
		download, err = downloadSpeed(&servers[i], 8, 1500)
		if err == nil {
			break
		}
	}

	return hub.NewBandwidthFromInt64(upload, download), err
}