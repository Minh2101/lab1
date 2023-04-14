package exportPDF

import (
	"bytes"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"lab1/controllers"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ExportPDF godoc
// @Summary Export an HTML file to a PDF file.
// @Description Converts html to a PDF file using wkhtmltopdf library and returns the PDF file.
// @Accept application/html
// @Produce application/pdf
// @Param html body string true "HTML file to be converted to PDF"
// @Success 200 {file} PDF "PDF file as an attachment"
// @Failure 500 {object} string "Internal Server Error"
// @Router /export-pdf [post]
func ExportPDF(c *gin.Context) {
	//Request html
	html, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		controllers.ResponseError(c, http.StatusInternalServerError, "Failed to read request body", nil)
		return
	}
	//Set Path
	wkhtmltopdf.SetPath("C:\\Program Files\\wkhtmltopdf\\bin\\wkhtmltopdf.exe")
	pdf, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		controllers.ResponseError(c, http.StatusInternalServerError, "Failed to create PDF generator", nil)
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
		controllers.ResponseError(c, http.StatusInternalServerError, "Fail to create pfd", nil)
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
