package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gin-gonic/gin"
)

var (
	// Cloudflare API Dialer
	api *cloudflare.API

	// バックグラウンドコンテクスト
	ctx context.Context
)

// 待ち受けるサーバのルーターを定義します
func SetupRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group("v1")

	v1.PUT("/ipv4", updateRecord)
	v1.PUT("/ipv6", updateRecord)

	return router
}

// A / AAAA レコードのIPアドレスを更新します
func updateRecord(c *gin.Context) {
	auth_header := c.GetHeader("Authorization")
	if auth_header == "" {
		c.JSON(http.StatusUnauthorized, errNotFoundAuthorizationHeader)
		c.Abort()
		return
	}

	splitted := strings.Split(auth_header, " ")
	if len(splitted) != 2 {
		c.JSON(http.StatusUnauthorized, errInvalidAuthorizationHeader)
		c.Abort()
		return
	}

	if splitted[0] != "Bearer" || splitted[1] == "" {
		c.JSON(http.StatusUnauthorized, errInvalidAuthorizationHeader)
		c.Abort()
		return
	}

	if err := initAPI(splitted[1]); err != nil {
		c.JSON(http.StatusForbidden, ErrorMessage{
			Code:    "E501",
			Message: err.Error(),
		})
		c.Abort()
		return
	}

	var req UpdateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errInvalidRequestedData)
		c.Abort()
		return
	}

	str_zone_id, err := api.ZoneIDByName(req.ZoneName)

	if err != nil {
		c.JSON(http.StatusBadRequest, errNotFoundZone)
		c.Abort()
		return
	}

	zone_id := cloudflare.ZoneIdentifier(str_zone_id)
	rec_type := "A"

	if strings.Contains(c.FullPath(), "ipv6") {
		rec_type = "AAAA"
	}

	param := cloudflare.ListDNSRecordsParams{Type: rec_type}

	recs, _, err := api.ListDNSRecords(ctx, zone_id, param)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Code:    "E501",
			Message: err.Error(),
		})
		c.Abort()
		return
	}

	res := make([]ResponseSet, 0, len(req.Contents))
	update_targets := make([]cloudflare.UpdateDNSRecordParams, 0, len(recs))

	for _, data := range req.Contents {
		name := fmt.Sprintf("%s.%s", data.HostName, req.ZoneName)

		if data.HostName == "@" {
			name = req.ZoneName
		}

		if !isValidIPAddress(data.Content) {
			res = append(res, ResponseSet{
				Name:      name,
				Content:   data.Content,
				Succeeded: false,
				Error:     "無効なIPアドレスです",
			})

			continue
		}

		exists := false

		for _, record := range recs {
			if record.Name == name {
				exists = true

				if record.Content == data.Content {
					res = append(res, ResponseSet{
						Name:      record.Name,
						Content:   data.Content,
						Succeeded: false,
						Error:     "Contentが既存の設定値と同一です",
					})

					break
				}

				p := cloudflare.UpdateDNSRecordParams{
					Type:    rec_type,
					Name:    record.Name,
					Content: data.Content,
					ID:      record.ID,
					TTL:     record.TTL,
					Proxied: record.Proxied,
				}

				update_targets = append(update_targets, p)
				break
			}
		}

		if !exists {
			err_msg := fmt.Sprintf(
				"対象のホスト名の%sレコードが存在しません", rec_type)

			res = append(res, ResponseSet{
				Name:      name,
				Content:   data.Content,
				Succeeded: false,
				Error:     err_msg,
			})
		}
	}

	if len(update_targets) != 0 {
		for _, target := range update_targets {
			res_rec, err := api.UpdateDNSRecord(ctx, zone_id, target)
			if err != nil {
				res = append(res, ResponseSet{
					Name:      target.Name,
					Content:   target.Content,
					Succeeded: false,
					Error:     "更新に失敗しました",
				})

				continue
			}

			res = append(res, ResponseSet{
				Name:      res_rec.Name,
				Content:   res_rec.Content,
				Succeeded: true,
				Error:     "",
			})
		}
	}

	c.JSON(http.StatusOK, UpdateResponse{
		ZoneName: req.ZoneName,
		Results:  res,
	})
}

// Cloudflare API Dialerを初期化します
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

// 有効なIPアドレスか確認します
func isValidIPAddress(ip string) bool {
	parsed := net.ParseIP(ip)

	if parsed == nil {
		return false
	}

	return ip == parsed.String()
}
