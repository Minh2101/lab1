package controllers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

type RequestTable struct {
	Page      int64  `json:"page" form:"page"`
	Length    int64  `json:"length" form:"length"`
	Search    string `json:"search" form:"search"`
	Order     string `json:"order_by" form:"order_by"`
	Dir       string `json:"order_dir" form:"order_dir"`
	DirNumber int    `json:"-"`
	Total     int64  `json:"total"`
	Deleted   bool   `json:"deleted"`
}

func (u *RequestTable) CustomOptions(opts *options.FindOptions) *options.FindOptions {
	return opts.SetSort(bson.M{u.Order: u.DirNumber}).SetLimit(u.Length).SetSkip((u.Page - 1) * u.Length)
}

func (u *RequestTable) CustomFilters(filter bson.M) bson.M {
	if u.Deleted == false {
		filter["deleted_at"] = nil
	}
	return filter
}

func BindRequestTable(c *gin.Context, order string) RequestTable {
	var request RequestTable
	_ = c.BindQuery(&request)
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Length <= 0 {
		request.Length = 12
	}
	if request.Search != "" {
		request.Search = strings.TrimSpace(request.Search)
	}
	if request.Order == "" {
		request.Order = order
	}
	if strings.ToLower(request.Dir) != "asc" {
		request.DirNumber = -1
		request.Dir = "desc"
	} else {
		request.DirNumber = 1
		request.Dir = "asc"
	}
	return request
}

func ResponseSuccess(c *gin.Context, code int, message string, data interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"data":    data,
		"message": message,
	})
}

func ResponseError(c *gin.Context, code int, message string, error interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"code":    code,
		"error":   error,
		"message": message,
	})
}
