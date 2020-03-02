package  router

import ("../controllers"
"github.com/gin-gonic/gin"
		"fmt"
)

type Handler struct {}

type RequestJson struct{
	Event string `json:"event"`
	Token string `json:"token"`
}

func (h *Handler) NewEventHTTP(c *gin.Context){
	d :=controllers.DynamodbDriver{}
	event := RequestJson{}
	err := c.ShouldBindJSON(&event)
	if err !=nil{
		c.JSON(400,gin.H{
			"error":"Error on event in JSON",
		})
		return
	}
	_ , err = d.NewEvent(event.Event)
	if err!=nil{
		c.JSON(400,gin.H{
			"error":"Error on new Event Create",
		})
		return
	}
	c.JSON(200,gin.H{
		"message":"New Event Created",
	})

}


func (h *Handler)IssueTokenHTTP(c *gin.Context){
	d := controllers.DynamodbDriver{}
	event :=  RequestJson{}
	err := c.ShouldBindJSON(&event)
	if err !=nil{
		c.JSON(400,gin.H{
			"error":"You need `event` in JSON POST request",
		})
		return
	}
	token,err := d.GetToken(event.Event)
	if err !=nil{
		c.JSON(400,gin.H{
			"error":"Error on the string Name",
		})
		return
	}
	tokenString := fmt.Sprintf("%v",token)

	c.SetCookie(
		"token",
		tokenString,
		60*60,
		"/",
		"localhost",
		false,
		true,
		)
	c.JSON(200,gin.H{
		"token":tokenString,
	})
}

func (h *Handler) ValidateTokenAndCheckQueueHTTP(c *gin.Context){
	d := controllers.DynamodbDriver{}
	token := RequestJson{}
	err := c.ShouldBindJSON(&token)
	if err!=nil{
		c.JSON(400,gin.H{
			"error":"Error on token String in JSON",

		})
		return
	}
	claims , err := d.ValidateToken(token.Token)
	if err !=nil{
		c.JSON(400,gin.H{
			"error":"Error on Token Validation",

		})
		return
	}


	newTokenString ,err := d.CheckQueue(claims)
	if err !=nil{
		c.JSON(400,gin.H{
			"error":err,
		})
		return
	}

	c.JSON(200,gin.H{
		"token":newTokenString,
	})
}