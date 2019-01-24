package qiniu // import "go.zoe.im/s3-qiniu/qiniu"

import (
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"

	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type S3 struct {
	accessKey, secretKey string
	zonesLock            sync.RWMutex
	zones                map[string]*storage.Zone
	bm                   *storage.BucketManager
}

func (s *S3) WithZoneInfo(bucket string) (*storage.Zone, error) {
	s.zonesLock.RLock()
	if z, ok := s.zones[bucket]; ok {
		s.zonesLock.RUnlock()
		return z, nil
	}
	s.zonesLock.RUnlock()

	// try to get zone info from remote
	// get zone of bucket
	zone, err := s.bm.Zone(bucket)
	if err != nil {
		return nil, err
	}

	s.zonesLock.Lock()
	s.zones[bucket] = zone
	s.zonesLock.Unlock()

	return zone, nil
}

func (s *S3) Stat(bucket, key string, opts ...RequestOption) (data GetRet, err error) {

	c := NewConfig(opts...)

	zone, err := s.WithZoneInfo(bucket)
	if err != nil {
		return
	}

	host := zone.GetRsHost(c.useHTTPS)
	if !strings.HasPrefix(host, "http") {
		host = "http://" + host
	}

	ctx := context.WithValue(context.TODO(), "mac", s.bm.Mac)

	url := fmt.Sprintf("%s/%s/%s", host, "get", storage.EncodedEntry(bucket, key))

	err = s.bm.Client.Call(ctx, &data, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	return
}

func NewS3(opts ...Option) (*S3, error) {
	s := &S3{
		zones: make(map[string]*storage.Zone),
	}

	for _, o := range opts {
		o(s)
	}

	// do prepare checker
	if s.accessKey == "" {
		return nil, errors.New("empty access key.")
	}

	if s.secretKey == "" {
		return nil, errors.New("empty secret key.")
	}

	// init with server and check with server
	mac := qbox.NewMac(s.accessKey, s.secretKey)
	s.bm = storage.NewBucketManager(mac, &storage.Config{})

	_, err := s.bm.Buckets(false)
	if err != nil {
		return nil, fmt.Errorf("init qiniu qbox ends with error: %s", err)
	}

	// it's really for serve

	return s, nil
}
