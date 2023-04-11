package controllers

import (
	"bytes"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
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

func CreateItem(c *gin.Context) {
	var (
		//data  = bson.M{}
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

	//data["entry"] = entry
	ResponseSuccess(c, http.StatusCreated, "Tạo dữ liệu thành công", entry)

	return
}

func UpdateItem(c *gin.Context) {
	var (
		//data  = bson.M{}
		entryId, _ = primitive.ObjectIDFromHex(c.Param("id"))
		entry      = collections.Item{}
		exist      = collections.Item{}
		db         = database.GetMongoDB()
		err        error
	)
	//Check exist data
	filter := bson.M{
		"_id":        entryId,
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

func ChangeStatusItems(c *gin.Context) {
	var (
		//data    = bson.M{}
		db      = database.GetMongoDB()
		entry   = collections.Item{}
		entries = collections.Items{}
		err     error
		request = ListIDRequest{}
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
	//data["entries"] = entries
	ResponseSuccess(c, http.StatusOK, "Cập nhật dữ liệu thành công!", entries)
}

func DeleteItems(c *gin.Context) {
	var (
		db      = database.GetMongoDB()
		entry   = collections.Item{}
		entries = collections.Items{}
		err     error
		request = ListIDRequest{}
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

func ExportListItems(c *gin.Context) {
	var (
		b          bytes.Buffer
		err        error
		fileName   string
		entries    = collections.Items{}
		entry      = collections.Item{}
		db         = database.GetMongoDB()
		pagination = BindRequestTable(c, "created_at")
		file       = excelize.NewFile()
	)
	filter := bson.M{
		"deleted_at": nil,
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

func ExportPDF(c *gin.Context) {
	//Request html
	html, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to read request body", nil)
		return
	}
	//Set Path
	wkhtmltopdf.SetPath("C:\\Program Files\\wkhtmltopdf\\bin\\wkhtmltopdf.exe")
	pdf, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to create PDF generator", nil)
		return
	}

	page := wkhtmltopdf.NewPageReader(bytes.NewBuffer(html))
	pdf.AddPage(page)

	// Set options for PDF generator
	pdf.Dpi.Set(300)
	pdf.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	page.FooterRight.Set("[page]")

	err = pdf.Create()
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "Fail to create pfd", nil)
		return
	}
	// Write PDF to file
	filename := "output-" + time.Now().Format("15-04-05-02-01-2006") + ".pdf"
	path, _ := os.Getwd()
	os.Mkdir(filepath.Join(path, "pdf"), 0755)
	filePath := filepath.Join(path, "pdf", filename)
	if err = pdf.WriteFile(filePath); err != nil {
		return
	}
	// Set headers for response
	pdfBytes := pdf.Bytes()
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(200, "application/pdf", pdfBytes)
}
