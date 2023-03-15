package controllers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lab1/collections"
	"lab1/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"go.mongodb.org/mongo-driver/bson"
)

type ListID struct {
	ID []primitive.ObjectID `json:"id"`
}

func CreateItem(c *gin.Context) {
	data := bson.M{}
	entry := collections.Item{}
	DB := database.GetMongoDB()
	var err error
	// Bind dữ liệu
	if err = c.ShouldBindBodyWith(&entry, binding.JSON); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Dữ liệu gửi lên không chính xác", nil)
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
		ResponseError(c, http.StatusInternalServerError, "Binding lỗi", nil)
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

func DeleteItems(c *gin.Context) {
	DB := database.GetMongoDB()
	entry := collections.Item{}
	entries := collections.Items{}
	var err error
	request := ListID{}
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Binding dữ liệu lỗi", err)
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
