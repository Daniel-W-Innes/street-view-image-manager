package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
)

const radius = 6371

type Location struct {
	Latitude  float64 `json:"lat,omitempty"`
	Longitude float64 `json:"lng,omitempty"`
}

func toRadians(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}
func (l Location) toString() string {
	return fmt.Sprintf("%f,%f", l.Latitude, l.Longitude)
}

func (l Location) distance(location Location) float64 {
	latDistance := toRadians(location.Latitude - l.Latitude)
	lonDistance := toRadians(location.Longitude - l.Longitude)
	a := math.Sin(latDistance/2)*math.Sin(latDistance/2) + math.Cos(toRadians(l.Latitude))*math.Cos(toRadians(location.Latitude))*math.Sin(lonDistance/2)*math.Sin(lonDistance/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return radius * c * 1000
}

type Metadata struct {
	Copyright string   `json:"copyright,omitempty"`
	Date      string   `json:"date,omitempty"`
	Location  Location `json:"location"`
	PanoId    string   `json:"pano_id,omitempty"`
	Status    string   `json:"status,omitempty"`
}

const angleTolerance = 10

func repeatForAngleTolerance(downloadRequest DownloadRequest, output chan<- DownloadRequest) {
	minAngle := downloadRequest.angle - angleTolerance/2
	maxAngle := downloadRequest.angle + angleTolerance/2
	if minAngle < 0 {
		for i := 0; i < maxAngle; i++ {
			output <- DownloadRequest{location: downloadRequest.location, angle: i}
		}
		for i := 360 + minAngle; i < 360; i++ {
			output <- DownloadRequest{location: downloadRequest.location, angle: i}
		}
	} else if maxAngle > 360 {
		for i := 0; i < (360 - maxAngle); i++ {
			output <- DownloadRequest{location: downloadRequest.location, angle: i}
		}
		for i := 360; i > minAngle; i++ {
			output <- DownloadRequest{location: downloadRequest.location, angle: i}
		}
	} else {
		for i := minAngle; i < maxAngle; i++ {
			output <- DownloadRequest{location: downloadRequest.location, angle: i}
		}
	}
}

func getMetadata(location Location, key string) (*Metadata, error) {
	response, err := http.Get(fmt.Sprintf("https://maps.googleapis.com/maps/api/streetview/metadata?location=%s&key=%s", location.toString(), key))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error bad status from googleapis %d", response.StatusCode)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	metadata := Metadata{}
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func preload(input <-chan DownloadRequest, output chan<- DownloadRequest, key string) {
	for {
		downloadRequest, ok := <-input
		if !ok {
			close(output)
			return
		}
		metadata, err := getMetadata(downloadRequest.location, key)
		if err != nil {
			return
		}
		repeatForAngleTolerance(DownloadRequest{metadata.Location, downloadRequest.angle}, output)
	}
}
