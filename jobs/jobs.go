package jobs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/CenturyLinkLabs/docker-reg-client/registry"
	"github.com/astaxie/beego"
	irc "github.com/thoj/go-ircevent"
	c "github.com/wanghaoran1988/itrms/controllers"
	"github.com/wanghaoran1988/itrms/models"
)

const server = "irc.lab.bos.redhat.com:6667"
const channel = "#devexp-standup"

var msg = ""

func init() {
	go StartIRCRobot(channel, server)
}

func UpdateImageIDTask() {
	msg = ""
	beego.Info("UpdateImageIDTask")
	imageList := c.ListImage()

	client := registry.NewClient()
	basicAuth := registry.NilAuth{}
	var imageTotal, imageChanged int = 0, 0
	imageTotal = len(imageList)

	for _, image := range imageList {
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
			event.Desc = image.ImageName
			jsonString, _ = json.Marshal(event)
			models.Etcdclient.CreateInOrder(context.Background(), c.EventPrefix, string(jsonString), nil)
			// set the mssage to sent to irc
		} else {
			beego.Info("image:", image.ImageName, "not changed ")
		}
		if image.Status == c.ImageStatusNew {
			imageChanged += 1
			msg += image.Owner + ", Image updated:" + image.ImageName + ". "
		}

	}

	models.Etcdclient.Set(context.Background(), c.EtcdPrefixImageCount+"total", strconv.Itoa(imageTotal), nil)
	models.Etcdclient.Set(context.Background(), c.EtcdPrefixImageCount+"changed", strconv.Itoa(imageChanged), nil)
	beego.Info("We have ", imageTotal, " images and ", imageChanged, " changed")
}
func StartIRCRobot(channel, server string) {
	rand.Seed(time.Now().UnixNano())
	irccon1 := irc.IRC("devexp-robot", "robot")
	irccon1.VerboseCallbackHandler = true

	irccon1.AddCallback("001", func(e *irc.Event) { irccon1.Join(channel) })
	irccon1.AddCallback("002", func(e *irc.Event) {
		go func(e *irc.Event) {
			tick := time.NewTicker(30 * time.Minute)
			for {
				<-tick.C
				if msg != "" {
					irccon1.Privmsgf(channel, "%s\n", msg)
				}
			}
		}(e)
	})
	err := irccon1.Connect(server)
	if err != nil {
		fmt.Printf("error connect to server")
	}
	irccon1.Loop()
}
