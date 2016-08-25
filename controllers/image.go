package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	"github.com/astaxie/beego"
	"github.com/coreos/etcd/client"
	"github.com/wanghaoran1988/itrms/models"
	"golang.org/x/net/context"
)

const (
	EtcdPrefixImage      = "openshift.io/image_"
	EtcdPrefixImageCount = "openshift_image_count_"
	EtcdPrefix           = "openshift.io/"
	ImageStatusNew       = "New"
	ImageStatusPass      = "Pass"
	EventPrefix          = "openshift_event"
)

type ImageInfo struct {
	ImageName string `form:"imagename"`
	Owner     string `form:"imageowner"`
	Mail      string `form:"mail"`
	Notes     string `form:"notes"`
	ImageID   string `form: "imageID"`
	Status    string `form: "status"`
	ID        string
}
type ImageController struct {
	beego.Controller
}
type AddImageController struct {
	beego.Controller
}

func (c *AddImageController) Get() {
	c.TplName = "addImage.html"
}
func (c *ImageController) ListAll() {
	resp, err := models.Etcdclient.Get(context.Background(), EtcdPrefix, &client.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Nodes)
		imageList := ListImage()
		c.Data["imageList"] = imageList
		beego.Info("imageList:", imageList)
	}
	c.TplName = "image.html"
}
func (c *ImageController) PassImage() {
	imageID := c.GetString("id")
	beego.Info("Image :", imageID, "passed")
	resp, err := models.Etcdclient.Get(context.Background(), EtcdPrefixImage+imageID, nil)
	if err != nil {
		beego.Info(err)
	} else {
		node := resp.Node
		beego.Info("node:", node.Value)
		imageInfo := ImageInfo{}
		json.Unmarshal([]byte(node.Value), &imageInfo)
		imageInfo.Status = ImageStatusPass
		jsonString, err := json.Marshal(imageInfo)
		if err != nil {
			fmt.Println(err)
		}
		resp, err := models.Etcdclient.Set(context.Background(), EtcdPrefixImage+imageID, string(jsonString), nil)
		if err != nil {
			beego.Error(err)
		} else {
			// print common key info
			log.Printf("Set is done. Metadata is %q\n", resp)
		}
	}
	c.TplName = "image.html"
	c.Redirect("image", 303)
}
func (c *ImageController) DelImage() {
	imageID := c.GetString("id")
	resp, err := models.Etcdclient.Delete(context.Background(), EtcdPrefixImage+imageID, nil)
	if err != nil {
		beego.Error(err)
	} else {
		beego.Info("Delete image :", resp.Node.Value)
	}
	beego.Info("Image :", imageID, "deleted")
	c.Ctx.WriteString("success !")

}
func (c *ImageController) Post() {
	imageInfo := ImageInfo{}
	if err := c.ParseForm(&imageInfo); err != nil {
		beego.Error(err)
	}
	slices := strings.Split(imageInfo.ImageName, "/")
	imageRegistry := slices[0]
	imageName := slices[1] + "/" + strings.Split(slices[2], ":")[0]
	var tag string
	if len(strings.Split(slices[2], ":")) == 1 {
		tag = "latest"
		imageInfo.ImageName += ":latest"
	} else {
		tag = strings.Split(slices[2], ":")[1]
	}
	imageList := ListImage()
	var exits bool = false
	for _, image := range imageList {
		if image.ImageName == imageInfo.ImageName {
			beego.Warn("image:", image.ImageName, "already added")
			exits = true
			break
		}
	}
	if !exits {
		c := registry.NewClient()
		basicAuth := registry.NilAuth{}
		fmt.Printf(" registry:%s\n", imageRegistry)
		fmt.Printf(" image:%s\n", imageName)
		fmt.Printf(" tag:%s\n", tag)
		registryBaseUrl := "http://" + imageRegistry + "/v1/"
		c.BaseURL, _ = url.Parse(registryBaseUrl)
		id, err := c.Repository.GetImageID(imageName, tag, basicAuth)
		if err != nil {
			fmt.Printf("the %s image does not exits ", imageName)
		} else {
			fmt.Printf("the %s image have id : %s ", imageName, id)
			imageInfo.ID = strconv.Itoa(models.GetImageID())
			imageInfo.ImageID = id
			imageInfo.Status = ImageStatusNew
			jsonString, err := json.Marshal(imageInfo)
			if err != nil {
				fmt.Println(err)
			}

			beego.Info("ImageInfo :", imageInfo)
			resp, err := models.Etcdclient.Set(context.Background(), EtcdPrefixImage+imageInfo.ID, string(jsonString), nil)
			if err != nil {
				log.Fatal(err)
			} else {
				// print common key info
				log.Printf("Set is done. Metadata is %q\n", resp)
			}
		}
	}
	c.TplName = "image.html"
	c.Redirect("image", 303)
}
func (c *ImageController) Get() {
	total, err := models.Etcdclient.Get(context.Background(), EtcdPrefixImageCount+"total", nil)
	if err != nil {
		c.Data["imageTotalCount"] = 0
	} else {

		c.Data["imageTotalCount"] = total.Node.Value
	}
	changed, err := models.Etcdclient.Get(context.Background(), EtcdPrefixImageCount+"changed", nil)
	if err != nil {
		c.Data["imageChangedCount"] = 0
	} else {
		c.Data["imageChangedCount"] = changed.Node.Value
	}
	eventList := ListEvent()
	c.Data["eventList"] = eventList
	beego.Info("eventList:", eventList)
	c.TplName = "index.html"
}
func ListImage() []ImageInfo {
	resp, err := models.Etcdclient.Get(context.Background(), EtcdPrefix, &client.GetOptions{Recursive: true, Sort: true})
	imageList := []ImageInfo{}
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("nodes %q\n", resp.Node.Nodes)
		for _, node := range resp.Node.Nodes {

			beego.Info("node:", node.Value)
			imageInfo := ImageInfo{}
			json.Unmarshal([]byte(node.Value), &imageInfo)
			imageList = append(imageList, imageInfo)
		}
	}
	return imageList
}
func ListEvent() []models.Event {
	resp, err := models.Etcdclient.Get(context.Background(), EventPrefix, &client.GetOptions{Recursive: true, Sort: true})
	eventList := []models.Event{}
	if err != nil {
		log.Fatal(err)
	} else {
		log.Printf("nodes %q\n", resp.Node.Nodes)
		for _, node := range resp.Node.Nodes {

			beego.Info("node:", node.Value)
			event := models.Event{}
			json.Unmarshal([]byte(node.Value), &event)
			eventList = append(eventList, event)
		}
	}
	return eventList
}
