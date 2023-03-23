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
	var (
		data       = bson.M{}
		db         = database.GetMongoDB()
		entry      = collections.Item{}
		entries    = collections.Items{}
		pagination = BindRequestTable(c, "created_at")
		filter     = pagination.CustomFilters(bson.M{})
		opts       = pagination.CustomOptions(options.Find())
	)

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
		statusItem, err := strconv.ParseBool(c.Request.FormValue("status"))
		if err != nil {
			ResponseError(c, http.StatusBadRequest, "Trạng thái item không hợp lệ", nil)
			return
		}
		filter["status"] = statusItem
	}
	//Search theo khoảng thời gian tạo item
	fromDate := ConvertTimeYYYYMMDD(c.Request.FormValue("from-date"))
	toDate := ConvertTimeYYYYMMDD(c.Request.FormValue("to-date"))
	if !fromDate.IsZero() {
		filter["created_at"] = bson.M{
			"$gte": fromDate,
		}
	}
	if !toDate.IsZero() {
		filter["created_at"] = bson.M{
			"$lte": toDate.AddDate(0, 0, 1),
		}
	}

	var err error
	if entries, err = entry.Find(db, filter, opts); err != nil && err != mongo.ErrNoDocuments {
		ResponseError(c, http.StatusInternalServerError, "Tìm kiếm dữ liệu lỗi", nil)
		return
	} else if err == mongo.ErrNoDocuments {
		ResponseError(c, http.StatusNotFound, "Không tìm thấy dữ liệu", nil)
		return
	}
	pagination.Total, _ = entry.Count(db, filter)
	data["entries"] = entries
	data["pagination"] = pagination
	ResponseSuccess(c, http.StatusOK, "Lấy dữ liêu thành công", data)
	return
}

func CreateItem(c *gin.Context) {
	var (
		data  = bson.M{}
		entry = collections.Item{}
		db    = database.GetMongoDB()
		err   error
	)
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
	if err = entry.Create(db); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Tạo item lỗi", nil)
		return
	}

	data["entry"] = entry
	ResponseSuccess(c, http.StatusOK, "Tạo dữ liêu thành công", data)

	return
}

func UpdateItem(c *gin.Context) {
	var (
		data  = bson.M{}
		entry = collections.Item{}
		db    = database.GetMongoDB()
		err   error
	)

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
	//Check data
	exist := collections.Item{}
	filter := bson.M{
		"_id":        entry.ID,
		"deleted_at": nil,
	}
	if err = exist.First(db, filter); err == mongo.ErrNoDocuments {
		ResponseError(c, http.StatusNotFound, "Dữ liệu không tồn tại", nil)
		return
	}

	//Update
	if err = entry.Update(db); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Cập nhật dữ liệu lỗi", nil)
		return
	}

	data["entry"] = entry
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công", data)
	return
}

func ChangeStatusItems(c *gin.Context) {
	var (
		data    = bson.M{}
		db      = database.GetMongoDB()
		entry   = collections.Item{}
		entries = collections.Items{}
		err     error
		request = ListID{}
	)

	if err = c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Binding dữ liệu lỗi", err)
		return
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": request.ID,
		},
		"deleted_at": nil,
	}
	//Check data
	opts := options.Find()
	if entries, err = entry.Find(db, filter, opts); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Tìm kiếm dữ liệu lỗi", nil)
		return
	} else if len(entries) == 0 {
		ResponseError(c, http.StatusNotFound, "Dữ liệu không tồn tại", nil)
		return
	}
	//Update data
	for i, _ := range entries {
		entries[i].Status = !entries[i].Status
		_ = entries[i].Update(db)
	}
	data["entries"] = entries
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công!", data)
}

func DeleteItems(c *gin.Context) {
	var (
		db      = database.GetMongoDB()
		entry   = collections.Item{}
		entries = collections.Items{}
		err     error
		request = ListID{}
	)
	// Bind data
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
	//Check data
	opts := options.Find()
	if entries, err = entry.Find(db, filter, opts); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Tìm kiếm dữ liệu lỗi", nil)
		return
	} else if len(entries) == 0 {
		ResponseError(c, http.StatusNotFound, "Dữ liệu không tồn tại", nil)
		return
	}

	//Delete data
	for i, _ := range entries {
		err = entries[i].Delete(db)
	}
	ResponseSuccess(c, http.StatusOK, "Xóa dữ liệu thành công!", nil)
}
