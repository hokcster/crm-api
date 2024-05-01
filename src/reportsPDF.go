package main

import (
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/PCManiac/logrus_init"
	"github.com/labstack/echo/v4"

	"github.com/jung-kurt/gofpdf"
)

/*
Отчет по тренировке PDF
*/

func (h *handler) reportPDFWorkout(c echo.Context) error {
	club_id := c.Get("club_id").(int32)

	newpath := filepath.Join(".", "reports")
	os.MkdirAll(newpath, os.ModePerm)

	newpath += filepath.Join(".", "reports/"+string(club_id))
	os.MkdirAll(newpath, os.ModePerm)

	filename := "hello.pdf"
	filepath := filepath.Join(newpath, filename)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Hello, World!")

	err := pdf.OutputFileAndClose(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	mime_type := "application/pdf"

	f, err := os.Open(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.Stream(http.StatusOK, mime_type, f)
}
