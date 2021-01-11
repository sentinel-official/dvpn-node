package utils

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	radius = 6378.137
	p      = math.Pi / 180.0
)

func distance(lat1, lon1, lat2, lon2 float64) float64 {
	var (
		x1 = p * lat1
		y1 = p * lon1
		x2 = p * lat2
		y2 = p * lon2
	)

	return radius * math.Acos(
		math.Cos(y2-y1)*
			math.Cos(x2)*math.Cos(x1)+
			math.Sin(x2)*math.Sin(x1),
	)
}

type server struct {
	URL       string  `xml:"url,attr"`
	Latitude  float64 `xml:"lat,attr"`
	Longitude float64 `xml:"lon,attr"`
	distance  float64
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

	var result struct {
		Servers []server `xml:"servers>server"`
	}

	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Servers, nil
}

func uploadBandwidth(s server, load, size int) (int64, error) {
	var (
		client   = http.Client{Timeout: 5 * time.Second}
		data     = url.Values{}
		group, _ = errgroup.WithContext(context.Background())
	)

	data.Add("content", strings.Repeat("0", size))

	start := time.Now()
	for i := 0; i < load; i++ {
		group.Go(func() error {
			resp, err := client.PostForm(s.URL, data)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("recevied status code %d", resp.StatusCode)
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					panic(err)
				}
			}()

			_, _ = ioutil.ReadAll(resp.Body)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return 0, err
	}

	return int64(float64(load*size) / time.Since(start).Seconds()), nil
}

func downloadBandwidth(s server, load, size int) (int64, error) {
	var (
		client   = http.Client{Timeout: 5 * time.Second}
		endpoint = fmt.Sprintf("%s/random%dx%d.jpg", strings.Split(s.URL, "/upload")[0], size, size)
		group, _ = errgroup.WithContext(context.Background())
	)

	start := time.Now()
	for i := 0; i < load; i++ {
		group.Go(func() error {
			resp, err := client.Get(endpoint)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("recevied status code %d", resp.StatusCode)
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					panic(err)
				}
			}()

			_, _ = ioutil.ReadAll(resp.Body)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return 0, err
	}

	return int64(float64(load*2*size*size) / time.Since(start).Seconds()), nil
}

func Bandwidth() (upload, download int64, err error) {
	location, err := FetchGeoIPLocation()
	if err != nil {
		return 0, 0, err
	}

	servers, err := fetchServers()
	if err != nil {
		return 0, 0, err
	}

	for i := 0; i < len(servers); i++ {
		servers[i].distance = distance(
			location.Latitude, location.Longitude,
			servers[i].Latitude, servers[i].Longitude,
		)
	}

	sort.Slice(servers, func(i, j int) bool {
		return servers[i].distance < servers[j].distance
	})

	for i := range servers[:8] {
		upload, err = uploadBandwidth(servers[i], 8, 4*1e6)
		if err == nil {
			break
		}
	}

	for i := range servers[:8] {
		download, err = downloadBandwidth(servers[i], 8, 1500)
		if err == nil {
			break
		}
	}

	return upload, download, nil
}
