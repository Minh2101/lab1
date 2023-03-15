package controllers

import (
	"lab1/collections"
	"lab1/database"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"go.mongodb.org/mongo-driver/bson"
)

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
