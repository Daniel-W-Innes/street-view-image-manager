package manager

import (
	"image"
	"math"
	"sync"
)

type Cash struct {
	mux       sync.RWMutex
	pointCash map[Location]*Point
}

type Point struct {
	mux      sync.RWMutex
	distance float64
	images   map[int]image.Image
}

func (c *Cash) add(request DownloadRequest, img image.Image) {
	c.mux.RLock()
	if point, ok := c.pointCash[request.location]; ok {
		defer c.mux.RUnlock()
		point.mux.Lock()
		defer point.mux.Unlock()
		point.images[request.angle] = img
	} else {
		c.mux.RUnlock()
		p := Point{distance: math.MaxFloat64, images: make(map[int]image.Image)}
		p.images[request.angle] = img
		c.mux.Lock()
		defer c.mux.Unlock()
		c.pointCash[request.location] = &p
	}
}

func (c *Cash) has(request DownloadRequest) bool {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if point, ok := c.pointCash[request.location]; ok {
		point.mux.RLock()
		defer point.mux.RUnlock()
		_, ok = point.images[request.angle]
		return ok
	}
	return false
}

func (p *Point) update(l1, l2 Location, angle int, minDistance float64) (float64, image.Image, bool, bool) {
	p.mux.Lock()
	defer p.mux.Unlock()
	distance := l1.distance(l2)
	if distance < minDistance {
		p.distance = distance
		return distance, p.images[angle], false, true
	} else if distance < p.distance {
		return 0, nil, true, false
	}
	p.distance = distance
	return 0, nil, false, false
}

func (c *Cash) removeInLoop(location Location) {
	c.mux.RUnlock()
	c.mux.Lock()
	defer c.mux.Unlock()
	defer c.mux.RLock()
	delete(c.pointCash, location)
}

func (c *Cash) getAndClean(request DownloadRequest) image.Image {
	c.mux.RLock()
	defer c.mux.RUnlock()
	minDistance := math.MaxFloat64
	var next image.Image
	for l, point := range c.pointCash {
		newDistance, img, remove, newNext := point.update(l, request.location, request.angle, minDistance)
		if newNext {
			minDistance = newDistance
			next = img
		} else if remove {
			c.removeInLoop(l)
		}
	}
	return next
}
