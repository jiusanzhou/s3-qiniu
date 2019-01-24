package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.zoe.im/knife-go/convert"
	"go.zoe.im/s3-qiniu/qiniu"
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
)

type service struct {
	addr, accessKey, secretKey       string
	defaultBucket                    string
	enablePage                       bool
	wildMode                         bool
	whiteBucketList, blackBucketList []string

	*qiniu.S3
}

func newService() (*service, error) {
	s := &service{
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
		return nil, errors.New("empty access key")
	}

	if s.secretKey == "" {
		return nil, errors.New("empty secret key")
	}

	if s.defaultBucket == "" && !s.wildMode {
		return nil, errors.New("you must offer a bucket name or use wild mode")
	}

	var err error

	s.S3, err = qiniu.NewS3(qiniu.AccessKey(s.accessKey), qiniu.SecretKey(s.secretKey))
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	bucket := vars["bucket"]
	objectID := vars["key"]

	info, err := s.Stat(bucket, objectID)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Write(convert.String2Bytes(err.Error()))
		return
	}

	if r.URL.Query().Get("stat") == "true" {
		data, _ := json.Marshal(info)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	w.Header().Set("Location", info.URL)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusTemporaryRedirect)

	return
}

func main() {

	flag.Parse()

	handler, err := newService()

	r := mux.NewRouter()
	r.PathPrefix("/{bucket}/{key}").Handler(handler)

	if err != nil {
		log.Printf("Init s3 with error: %s.\n", err)
		return
	}

	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Printf("Exit with error: %s.\n", err)
	}

	log.Println("exit s3-qiniu")
}
