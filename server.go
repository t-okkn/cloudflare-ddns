package main

import (
	"context"
	//"fmt"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gin-gonic/gin"
)

var api *cloudflare.API
var ctx context.Context

// 待ち受けるサーバのルーターを定義します
func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.PUT("/", updateRecord)

	return router
}

// A / AAAA レコードを更新します
func updateRecord(c *gin.Context) {
	token := "*****"
	if initAPI(token) != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "エラーだよ",
		})
		return
	}

	res, err := api.ListZonesContext(ctx)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, res)
}

func initAPI(token string) error {
	if api != nil {
		return nil
	}

	var err error
	api, err = cloudflare.NewWithAPIToken(token)

	if err != nil {
		return err
	}

	ctx = context.Background()
	return nil
}
