package handle

import (
	"github.com/fernandez14/spartangeek-blacker/mongo"
    "github.com/fernandez14/spartangeek-blacker/model"
	"github.com/gin-gonic/gin"
	"github.com/xuyu/goredis"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type SitemapAPI struct {
	DataService  *mongo.Service `inject:""`
	CacheService *goredis.Redis `inject:""`
}

func (di *SitemapAPI) GetSitemap(c *gin.Context) {

	var urls []model.SitemapUrl
	var posts []model.Post
	var location string

	// Get the database interface from the DI
	database := di.DataService.Database

	err := database.C("posts").Find(bson.M{}).Sort("-pinned", "-created_at").All(&posts)

	if err != nil {
		panic(err)
	}

	for _, post := range posts {

		// Generate the post url
		location = "http://www.spartangeek.com/p/" + post.Slug + "/" + post.Id.Hex()

		// Add to the sitemap url
		urls = append(urls, model.SitemapUrl{Location: location, Updated: post.Updated.Format("2006-01-02T15:04:05.999999-07:00"), Priority: "0.6"})
	}

	urls = append(urls, model.SitemapUrl{Location: "http://www.spartangeek.com", Updated: time.Now().Format("2006-01-02T15:04:05.999999-07:00"), Priority: "1.0"})

	sitemap := model.SitemapSet{
		Urls: urls,
		XMLNs: "http://www.sitemaps.org/schemas/sitemap/0.9",
		XSI: "http://www.w3.org/2001/XMLSchema-instance",
		XSILocation: "http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd",
	}

	c.XML(200, sitemap)
}