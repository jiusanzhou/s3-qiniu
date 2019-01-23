package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"errors"


	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"fmt"
)

var (
	defaultUrlLayout = "{Schema}://{Host}/{bucket}/{ObjectID}"
)

var (
	addr       = flag.String("addr", ":8082", "Listen address for s3-qiniu.")
	access     = flag.String("access_key", "", "Access key for s3-qiniu.")
	secret     = flag.String("secret_key", "", "Secret key for s3-qiniu.")
	bucket     = flag.String("bucket", "", "Default bucket for s3-qiniu.")
	enablePage = flag.Bool("enable_page", false, "Enable list page of bucket.")
	wildMode   = flag.Bool("wild_mode", true, "Access bucket dynamic for s3-qiniu.")

	urlLayout = flag.String("params_layout", defaultUrlLayout, "Define template to parse params.")
)

type s3 struct {
	addr, accessKey, secretKey       string
	defaultBucket                    string
	enablePage                       bool
	wildMode                         bool
	whiteBucketList, blackBucketList []string

	bm *storage.BucketManager
}

func NewS3() (*s3, error) {
	s := &s3{
		addr:          *addr,
		accessKey:     *access,
		secretKey:     *secret,
		defaultBucket: *bucket,
		wildMode:      *wildMode,
		enablePage:    *enablePage,
	}

	// load access key and secret key from envs
	if s.accessKey == "" {
		s.accessKey = os.Getenv("S3QINIU_ACCESS_KEY")
	}

	if s.secretKey == "" {
		s.secretKey = os.Getenv("S3QINIU_SECRET_KEY")
	}

	// do prepare checker
	if s.accessKey == "" {
		return nil, errors.New("empty access key.")
	}

	if s.secretKey == "" {
		return nil, errors.New("empty secret key.")
	}

	if s.defaultBucket == "" && !s.wildMode {
		return nil, errors.New("you must offer a bucket name or use wild mode.")
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

func (s *s3) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	return
}

func main() {

	handler, err := NewS3()
	if err != nil {
		log.Printf("Init s3 with error: %s.\n", err)
	}

	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Printf("Exit with error: %s.\n", err)
	}
}
