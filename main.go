package manager

import (
	"image"
	"time"
)

type Downloader struct {
	Input            chan DownloadRequest
	LocationUpdater  chan DownloadRequest
	Output           chan image.Image
	downloadRequests chan DownloadRequest
	cash             *Cash
}

type DownloadRequest struct {
	location Location
	angle    int
}

func exporter(cash *Cash, input <-chan DownloadRequest, output chan<- image.Image) {
	ticker := time.NewTicker(1 / 60 * time.Second)
	for range ticker.C {
		downloadRequest, ok := <-input
		if !ok {
			close(output)
			return
		}
		output <- cash.getAndClean(downloadRequest)
	}
}

func (d *Downloader) Run(key string) {
	go preload(d.Input, d.downloadRequests, key)
	go download(d.downloadRequests, d.cash, key)
	go exporter(d.cash, d.LocationUpdater, d.Output)
}
