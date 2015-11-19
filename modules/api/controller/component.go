package controller 

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-gonic/gin"
	"github.com/fernandez14/spartangeek-blacker/modules/components"
)

type ComponentAPI struct {
	Components *components.Module `inject:""`
}

func (this ComponentAPI) UpdatePrice(c *gin.Context) {
	
	var form ComponentPriceUpdateForm

	slug := c.Param("slug")

	if c.BindJSON(&form) == nil {

		component, err := this.Components.Get(bson.M{"slug": slug})

		if err != nil {
			c.JSON(400, gin.H{"message": "Invalid request, component not found.", "status": "error"})
			return
		}

		component.UpdatePrice(form.Price)

		c.JSON(200, gin.H{"status": "okay"})
	}
}

type ComponentPriceUpdateForm struct {
	Price  map[string]float64 `json:"price" binding:"required"`
}