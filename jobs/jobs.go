package jobs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	"github.com/astaxie/beego"
	c "github.com/wanghaoran1988/itrms/controllers"
	"github.com/wanghaoran1988/itrms/models"
)

func UpdateImageIDTask() {
	beego.Info(" UpdateImageIDTask")
	imageList := c.ListImage()

	client := registry.NewClient()
	basicAuth := registry.NilAuth{}
	var imageTotal, imageChanged int = 0, 0
	imageTotal = len(imageList)

	for _, image := range imageList {
		if image.Status == c.ImageStatusNew {
			imageChanged += 1
			continue
		}
		slices := strings.Split(image.ImageName, "/")
		imageRegistry := slices[0]
		imageName := slices[1] + "/" + strings.Split(slices[2], ":")[0]
		tag := strings.Split(slices[2], ":")[1]
		registryBaseUrl := "http://" + imageRegistry + "/v1/"
		client.BaseURL, _ = url.Parse(registryBaseUrl)
		imageID, err := client.Repository.GetImageID(imageName, tag, basicAuth)
		if err != nil {
			fmt.Printf("the %s image does not exits ", imageName)
		}
		if imageID != image.ImageID {
			imageChanged += 1
			beego.Info("image:", image.ImageName, "changed ")
			image.ImageID = imageID
			image.Status = c.ImageStatusNew
			jsonString, err := json.Marshal(image)
			if err != nil {
				beego.Error(err)
			}
			models.Etcdclient.Set(context.Background(), c.EtcdPrefixImage+image.ID, string(jsonString), nil)
			event := models.Event{}
			event.EventType = models.EventTypeImageUpdate
			event.Time = time.Now().Format(time.RFC3339)
			jsonString, _ = json.Marshal(event)
			models.Etcdclient.CreateInOrder(context.Background(), c.EventPrefix, string(jsonString), nil)
		} else {
			beego.Info("image:", image.ImageName, "not changed ")
		}
	}
	models.Etcdclient.Set(context.Background(), c.EtcdPrefixImageCount+"total", strconv.Itoa(imageTotal), nil)
	models.Etcdclient.Set(context.Background(), c.EtcdPrefixImageCount+"changed", strconv.Itoa(imageChanged), nil)
	beego.Info("We have ", imageTotal, " images and ", imageChanged, " changed")
}
