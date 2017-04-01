package models

import (
	"log"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/astaxie/beego"
	"github.com/coreos/etcd/client"
)

var Etcdclient client.KeysAPI

const (
	ImageIDPrefix = "/imagesid/test"
)

func init() {
	etcdurl := beego.AppConfig.String("etcd_url")
	cfg := client.Config{
		Endpoints: []string{etcdurl},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	Etcdclient = client.NewKeysAPI(c)
}

var mu sync.Mutex

func GetImageID() int {
	mu.Lock()
	defer mu.Unlock()
	var id int = 0
	resp, err := Etcdclient.Get(context.Background(), ImageIDPrefix, nil)
	if err != nil {
		log.Printf("Set is error. %s\n", err)
		//log.Fatal(err)
	} else {
		id, _ = strconv.Atoi(resp.Node.Value)
		id = id + 1
		beego.Info("id:", id)
	}
	resp, err = Etcdclient.Set(context.Background(), ImageIDPrefix, strconv.Itoa(id), nil)
	if err != nil {
		log.Printf("Set is error. %s\n", err)
		//log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp)
	}
	return id
}
