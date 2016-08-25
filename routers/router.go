package routers

import (
	"github.com/astaxie/beego"
	"github.com/wanghaoran1988/itrms/controllers"
)

func init() {
	beego.Router("/", &controllers.ImageController{})
	beego.Router("/image", &controllers.ImageController{}, "GET:ListAll")
	beego.Router("/image", &controllers.ImageController{}, "POST:Post")
	beego.Router("/addImage", &controllers.AddImageController{})
	beego.Router("/passImage", &controllers.ImageController{}, "*:PassImage")
	beego.Router("/delImage", &controllers.ImageController{}, "*:DelImage")
}
