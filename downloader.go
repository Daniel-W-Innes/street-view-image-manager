package manager

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const SIZE = "1280x960"

func getImageFromGoogle(request DownloadRequest, key string) (image.Image, error) {
	response, err := http.Get(fmt.Sprintf("https://maps.googleapis.com/maps/api/streetview?size=%s&Location=%f,%f&heading=%d&key=%s", SIZE, request.Location.Latitude, request.Location.Longitude, request.Angle, key))
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
	decode, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		log.Fatalln(err)
	}
	return decode, nil
}

func getImage(request DownloadRequest, key string) (image.Image, error) {
	path := fmt.Sprintf("Cash/%s,%d.jpg", request.Location.toString(), request.Angle)
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		img, err := getImageFromGoogle(request, key)
		if err != nil {
			return img, err
		}
		out, err := os.Create(path)
		if err != nil {
			return img, err
		}
		opt := jpeg.Options{Quality: 100}
		err = jpeg.Encode(out, img, &opt)
		return img, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func download(input <-chan DownloadRequest, cash *Cash, key string) {
	for {
		downloadRequest, ok := <-input
		if !ok {
			return
		}
		if !cash.has(downloadRequest) {
			img, err := getImage(downloadRequest, key)
			if err != nil {
				return
			}
			cash.add(downloadRequest, img)
		}
	}
}
