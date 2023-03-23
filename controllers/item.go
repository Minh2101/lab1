package controllers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lab1/collections"
	"lab1/database"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"go.mongodb.org/mongo-driver/bson"
)

type ListID struct {
	ID []primitive.ObjectID `json:"id"`
}

func ListItems(c *gin.Context) {
	data := bson.M{}
	DB := database.GetMongoDB()
	entry := collections.Item{}
	entries := collections.Items{}
	var err error
	var pagination = BindRequestTable(c, "created_at")

	filter := pagination.CustomFilters(bson.M{})
	opts := pagination.CustomOptions(options.Find())

	//Search theo title
	if pagination.Search != "" {
		filter["$or"] = []bson.M{
			{
				"title": bson.M{
					"$regex":   strings.TrimSpace(pagination.Search),
					"$options": "i",
				},
			},
		}
	}
	//Search theo status item
	if c.Request.FormValue("status") != "" {
		statusItem, _ := strconv.ParseBool(c.Request.FormValue("status"))
		filter["status"] = statusItem
	}
	//Search theo khoảng thời gian tạo item
	fromDate := ConvertTimeYYYYMMDD(c.Request.FormValue("from-date"))
	toDate := ConvertTimeYYYYMMDD(c.Request.FormValue("to-date"))
	if !fromDate.IsZero() || !toDate.IsZero() {
		if toDate.IsZero() {
			toDate = Now()
		} else {
			toDate = toDate.AddDate(0, 0, 1)
		}
		filter["created_at"] = bson.M{
			"$gte": fromDate,
			"$lte": toDate,
		}
	}

	if entries, err = entry.Find(DB, filter, opts); err != nil && err != mongo.ErrNoDocuments {
		ResponseError(c, http.StatusInternalServerError, "Tìm kiếm dữ liệu lỗi", nil)
		return
	} else if err == mongo.ErrNoDocuments {
		ResponseError(c, http.StatusNotFound, "Không tìm thấy dữ liệu", nil)
		return
	}
	pagination.Total, _ = entry.Count(DB, filter)
	data["entries"] = entries
	data["pagination"] = pagination
	ResponseSuccess(c, http.StatusOK, "Lấy dữ liêu thành công", data)
	return
}

func CreateItem(c *gin.Context) {
	data := bson.M{}
	entry := collections.Item{}
	DB := database.GetMongoDB()
	var err error
	// Bind dữ liệu
	if err = c.ShouldBindBodyWith(&entry, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Dữ liệu gửi lên không chính xác", nil)
		return
	}

	// Validate thông tin
	val := validate.Validate(
		&validators.StringIsPresent{Name: "Title", Field: entry.Title, Message: "Tiêu đề không được bỏ trống"},
	)
	if val.HasAny() {
		ResponseError(c, http.StatusUnprocessableEntity, val.Errors[val.Keys()[0]][0], nil)
		return
	}

	// Lưu dữ liệu
	if err = entry.Create(DB); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Tạo item lỗi", nil)
		return
	}

	data["entry"] = entry
	ResponseSuccess(c, http.StatusOK, "Tạo dữ liêu thành công", data)

	return
}

func UpdateItem(c *gin.Context) {
	data := bson.M{}
	entry := collections.Item{}
	DB := database.GetMongoDB()
	var err error

	//bind dữ liệu
	if err = c.ShouldBindBodyWith(&entry, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Binding lỗi", nil)
		return
	}
	// Validate
	val := validate.Validate(
		&validators.StringIsPresent{Name: "Title", Field: entry.Title, Message: "Tiêu đề không được bỏ trống"},
	)
	if val.HasAny() {
		ResponseError(c, http.StatusUnprocessableEntity, val.Errors[val.Keys()[0]][0], nil)
		return
	}
	if err = entry.Update(DB); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Cập nhật dữ liệu lỗi", nil)
		return
	}
	data["entry"] = entry
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công", data)
	return
}

func ChangeStatusItems(c *gin.Context) {
	data := bson.M{}
	DB := database.GetMongoDB()
	entry := collections.Item{}
	entries := collections.Items{}
	var err error
	request := ListID{}
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Binding dữ liệu lỗi", err)
		return
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": request.ID,
		},
		"deleted_at": nil,
	}
	opts := options.Find()
	if entries, err = entry.Find(DB, filter, opts); err != nil && err != mongo.ErrNoDocuments {
		ResponseError(c, http.StatusInternalServerError, "Tìm kiếm dữ liệu lỗi", nil)
		return
	}
	for i, _ := range entries {
		entries[i].Status = !entries[i].Status
		_ = entries[i].Update(DB)
	}
	data["entries"] = entries
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công!", data)
}

func DeleteItems(c *gin.Context) {
	DB := database.GetMongoDB()
	entry := collections.Item{}
	entries := collections.Items{}
	var err error
	request := ListID{}
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Binding dữ liệu lỗi", err)
		return
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": request.ID,
		},
		"deleted_at": nil,
	}
	opts := options.Find()
	if entries, err = entry.Find(DB, filter, opts); err != nil && err != mongo.ErrNoDocuments {
		ResponseError(c, http.StatusInternalServerError, "Tìm kiếm dữ liệu lỗi", nil)
		return
	}
	for i, _ := range entries {
		err = entries[i].Delete(DB)
	}
	ResponseSuccess(c, http.StatusOK, "Xóa dữ liệu thành công!", nil)
}
