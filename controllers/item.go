package controllers

import (
	"bytes"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"lab1/collections"
	"lab1/database"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ListIDRequest struct {
	ID []primitive.ObjectID `json:"id"`
}

// ListItems godoc
// @Summary get list items
// @Description get list items form the database
// @Tags items
// @Accept json
// @Produce json
// @Param from-date query string false "Ngày bắt đầu lấy dữ liệu theo format YYYY-MM-DD"
// @Param to-date query string false "Ngày kết thúc lấy dữ liệu theo format YYYY-MM-DD"
// @Param status query boolean false "Trạng thái item, true hoặc false"
// @Param search query string false "Từ khóa tìm kiếm theo tiêu đề item"
// @Success 200 {array} string "Lấy dữ liệu thành công"
// @Failure 400 {object} string "Trạng thái item tìm kiếm không hợp lệ"
// @Failure 404 {object} string "Không tìm thấy dữ liệu"
// @Failure 500 {object} string "Tìm kiếm dữ liệu lỗi"
// @Router /items [get]
func ListItems(c *gin.Context) {
	var (
		data       = bson.M{}
		db         = database.GetMongoDB()
		entry      = collections.Item{}
		entries    = collections.Items{}
		pagination = BindRequestTable(c, "created_at")
		filter     = pagination.CustomFilters(bson.M{})
		opts       = pagination.CustomOptions(options.Find())
		userID, _  = primitive.ObjectIDFromHex((c.MustGet("user_id").(string)))
	)
	//Filter
	filter["user_id"] = userID
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
			ResponseError(c, http.StatusBadRequest, "Trạng thái item tìm kiếm không hợp lệ", nil)
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
	ResponseSuccess(c, http.StatusOK, "Lấy dữ liệu thành công", data)
	return
}

// CreateItem godoc
// @Summary Create a Item
// @Description Create a new Item
// @Tags items
// @Accept json
// @Produce json
// @Param item body collections.Item true "New Item"
// @Success 201 {object} collections.Item "Tạo dữ liệu thành công"
// @Failure 400 {object} string "Dữ liệu gửi lên không chính xác"
// @Failure 422 {object} string "Tiêu đề không được bỏ trống"
// @Failure 500 {object} string "Tạo item lỗi"
// @Router /item [post]
func CreateItem(c *gin.Context) {
	var (
		//data  = bson.M{}
		db        = database.GetMongoDB()
		entry     = collections.Item{}
		err       error
		userID, _ = primitive.ObjectIDFromHex((c.MustGet("user_id").(string)))
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
	entry.UserId = userID
	if err = entry.Create(db); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Tạo item lỗi", nil)
		return
	}

	//data["entry"] = entry
	ResponseSuccess(c, http.StatusCreated, "Tạo dữ liệu thành công", entry)

	return
}

// UpdateItem godoc
// @Summary Update an item
// @Description Update an existing item
// @Tags items
// @Accept json
// @Produce json
// @Param id path string true "Item ID"
// @Param item body collections.Item true "Item object that needs to update"
// @Success 200 {object} collections.Item "Cập nhật dữ liệu thành công"
// @Failure 400 {object} string "Binding lỗi"
// @Failure 404 {object} string "Dữ liệu không tồn tại"
// @Failure 422 {object} string "Tiêu đề không được bỏ trống"
// @Failure 500 {object} string "Cập nhật dữ liệu lỗi"
// @Router /item/{id} [put]
func UpdateItem(c *gin.Context) {
	var (
		//data  = bson.M{}
		db         = database.GetMongoDB()
		entryId, _ = primitive.ObjectIDFromHex(c.Param("id"))
		entry      = collections.Item{}
		exist      = collections.Item{}
		err        error
		userID, _  = primitive.ObjectIDFromHex((c.MustGet("user_id").(string)))
	)
	//Check exist data
	filter := bson.M{
		"_id":        entryId,
		"user_id":    userID,
		"deleted_at": nil,
	}
	if err = exist.First(db, filter); err != nil && err != mongo.ErrNoDocuments {
		ResponseError(c, http.StatusInternalServerError, "Lấy dữ liệu lỗi", nil)
		return
	} else if err == mongo.ErrNoDocuments {
		ResponseError(c, http.StatusNotFound, "Không tìm thấy dữ liệu", nil)
		return
	}

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

	entry.ID = entryId
	entry.UserId = userID
	entry.CreatedAt = exist.CreatedAt
	//Update
	if err = entry.Update(db); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Cập nhật dữ liệu lỗi", nil)
		return
	}

	//data["entry"] = entry
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công", entry)
	return
}

// ChangeStatusItems godoc
// @Summary Change status items
// @Description change status items by ID
// @Tags items
// @Accept json
// @Produce json
// @Param ID body controllers.ListIDRequest true "Change status by listID"
// @Success 200 {array} string "Cập nhật dữ liệu thành công"
// @Failure 400 {object} string "Binding dữ liệu lỗi"
// @Failure 404 {object} string "Dữ liệu không tồn tại"
// @Failure 500 {object} string "Tìm kiếm dữ liệu lỗi"
// @Router /change-status-items [post]
func ChangeStatusItems(c *gin.Context) {
	var (
		//data    = bson.M{}
		db        = database.GetMongoDB()
		entry     = collections.Item{}
		entries   = collections.Items{}
		err       error
		userID, _ = primitive.ObjectIDFromHex((c.MustGet("user_id").(string)))
		request   = ListIDRequest{}
	)

	if err = c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Binding dữ liệu lỗi", err)
		return
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": request.ID,
		},
		"user_id":    userID,
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
	//data["entries"] = entries
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công!", entries)
}

// DeleteItems godoc
// @Summary Delete Items
// @Description Delete items by ID
// @Tags items
// @Accept json
// @Produce json
// @Param ID body controllers.ListIDRequest true "Delete items by listID"
// @Success 200 {array} string "Xóa dữ liệu thành công"
// @Failure 400 {object} string "Binding dữ liệu lỗi"
// @Failure 404 {object} string "Dữ liệu không tồn tại"
// @Failure 500 {object} string "Tìm kiếm dữ liệu lỗi"
// @Router /delete-items  [post]
func DeleteItems(c *gin.Context) {
	var (
		db        = database.GetMongoDB()
		entry     = collections.Item{}
		entries   = collections.Items{}
		err       error
		userID, _ = primitive.ObjectIDFromHex((c.MustGet("user_id").(string)))
		request   = ListIDRequest{}
	)
	// Bind data
	if err = c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		ResponseError(c, http.StatusBadRequest, "Binding dữ liệu lỗi", err)
		return
	}

	filter := bson.M{
		"_id": bson.M{
			"$in": request.ID,
		},
		"user_id":    userID,
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

// ExportListItems godoc
// @Summary export list items
// @Description export excel list items form the database
// @Tags items
// @Accept json
// @Produce json
// @Param from-date query string false "Ngày bắt đầu lấy dữ liệu theo format YYYY-MM-DD"
// @Param to-date query string false "Ngày kết thúc lấy dữ liệu theo format YYYY-MM-DD"
// @Param status query boolean false "Trạng thái item, true hoặc false"
// @Param search query string false "Từ khóa tìm kiếm theo tiêu đề item"
// @Success 200 {array} string "Trả về file excel"
// @Failure 400 {object} string "Trạng thái item tìm kiếm không hợp lệ"
// @Failure 500 {object} string "Lấy dữ liệu hoặc tạo file excel lỗi"
// @Router /export-items  [get]
func ExportListItems(c *gin.Context) {
	var (
		b          bytes.Buffer
		db         = database.GetMongoDB()
		err        error
		file       = excelize.NewFile()
		fileName   string
		entries    = collections.Items{}
		entry      = collections.Item{}
		pagination = BindRequestTable(c, "created_at")
		userID, _  = primitive.ObjectIDFromHex(c.MustGet("user_id").(string))
	)
	filter := bson.M{
		"deleted_at": nil,
		"user_id":    userID,
	}
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

	opts := options.Find().SetAllowDiskUse(true).SetSort(bson.M{"created_at": -1})
	if entries, err = entry.Find(db, filter, opts); err != nil {
		ResponseError(c, http.StatusInternalServerError, "Lấy dữ liệu lỗi!", err)
	}

	if file, fileName, err = ExportExcelListItems(entries); err != nil {
		return
	}
	if err = file.Write(&b); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/octet-stream", b.Bytes())
}

func ExportExcelListItems(entries collections.Items) (file *excelize.File, fileName string, err error) {
	sheetName := "Dữ Liệu"
	file = excelize.NewFile()
	index := file.NewSheet(sheetName)
	file.DeleteSheet("Sheet1")

	// set header
	file.SetCellValue(sheetName, "A1", "STT")
	file.SetCellValue(sheetName, "B1", "Title")
	file.SetCellValue(sheetName, "C1", "Status")
	file.SetCellValue(sheetName, "D1", "Created At")
	file.SetCellValue(sheetName, "E1", "Modified At")

	// set header
	header, _ := file.NewStyle(`{"alignment":{"horizontal":"center"}, "font":{"bold":true}, "border":[{"type":"left","color":"000000","style":1},
	{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]}`)
	file.SetCellStyle(sheetName, "A1", "E1", header)

	// set border and color
	border, _ := file.NewStyle(`{"alignment":{"horizontal":"center"}, "border":[{"type":"left","color":"000000","style":1},
	{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]}`)

	borderColumTitle, _ := file.NewStyle(`{"border":[{"type":"left","color":"000000","style":1},
	{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]}`)

	color, _ := file.NewStyle(`{"alignment":{"horizontal":"center"}, "fill":{"type":"pattern","color":["#99CC00"],"pattern":1}, 
	"border":[{"type":"left","color":"000000","style":1},{"type":"top","color":"000000","style":1},
	{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]}`)

	noColor, _ := file.NewStyle(`{"alignment":{"horizontal":"center"}, "border":[{"type":"left","color":"000000","style":1},
	{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]}`)

	// Fill Data
	countLine := 2
	for _, entry := range entries {
		file.SetCellValue(sheetName, "A"+strconv.Itoa(countLine), countLine-1)
		file.SetCellValue(sheetName, "B"+strconv.Itoa(countLine), entry.Title)
		if entry.Status {
			file.SetCellValue(sheetName, "C"+strconv.Itoa(countLine), "Đã làm")
			file.SetCellStyle(sheetName, "C"+strconv.Itoa(countLine), "C"+strconv.Itoa(countLine), color)
		} else {
			file.SetCellValue(sheetName, "C"+strconv.Itoa(countLine), "Chưa làm")
			file.SetCellStyle(sheetName, "C"+strconv.Itoa(countLine), "C"+strconv.Itoa(countLine), noColor)
		}
		file.SetCellValue(sheetName, "D"+strconv.Itoa(countLine), entry.CreatedAt.Format("15:04:05 02/01/2006"))
		file.SetCellValue(sheetName, "E"+strconv.Itoa(countLine), entry.ModifiedAt.Format("15:04:05 02/01/2006"))

		// Set Borders
		file.SetCellStyle(sheetName, "A"+strconv.Itoa(countLine), "A"+strconv.Itoa(countLine), border)
		file.SetCellStyle(sheetName, "B"+strconv.Itoa(countLine), "B"+strconv.Itoa(countLine), borderColumTitle)
		file.SetCellStyle(sheetName, "D"+strconv.Itoa(countLine), "E"+strconv.Itoa(countLine), border)

		countLine++
	}

	// set column width
	file.SetColWidth(sheetName, "B", "B", 20)
	file.SetColWidth(sheetName, "C", "C", 15)
	file.SetColWidth(sheetName, "D", "D", 20)
	file.SetColWidth(sheetName, "E", "E", 20)
	file.SetActiveSheet(index)

	fileName = "Export_ListItems" + time.Now().Format("15-04-05-02-01-2006") + ".xlsx"
	path, _ := os.Getwd()
	os.Mkdir(filepath.Join(path, "excel"), 0755)
	pathFile := filepath.Join(path, "excel", fileName)

	if err = file.SaveAs(pathFile); err != nil {
		return nil, "", err
	}
	return file, fileName, err
}
