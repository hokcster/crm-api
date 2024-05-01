package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FileInfo struct {
	OriginalfileName string `json:"name" db:"name"`
	FileSize         int64  `json:"size" db:"size"`
	MimeType         string `json:"type" db:"type"`
	FileSystemName   string `json:"fs_name" db:"fs_name"`
}

func (a *FileInfo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(b, &a)
}

type PlayerPhotoFileInfo struct {
	Id       int32    `json:"file_id" db:"file_id"`
	FileData FileInfo `json:"file_data" db:"file_data"`
}

func (h *handler) getPlayerPhoto(c echo.Context) error {
	player_id := c.Param("id")

	claims := c.Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	file_info := new(PlayerPhotoFileInfo)
	err := h.DB.QueryRow(`SELECT * FROM api_sight."playersGetPhoto"($1, $2);`, player_id, club_id).Scan(&file_info.Id, &file_info.FileData)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/player.png")
	}

	filepath := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + file_info.FileData.FileSystemName
	mime_type := file_info.FileData.MimeType

	f, err := os.Open(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.Stream(http.StatusOK, mime_type, f)
}

func (h *handler) uploadPlayerPhoto(c echo.Context) error {
	player_id := c.Param("id")

	claims := c.Get("user").(*jwt.Token).Claims.(*UserClaims)
	user_id := claims.ID
	club_id := int(claims.Data["club_id"].(float64))

	old_file_info := new(PlayerPhotoFileInfo)
	err := h.DB.QueryRow(`SELECT * FROM api_sight."playersGetPhoto"($1, $2);`, player_id, club_id).Scan(&old_file_info.Id, &old_file_info.FileData)
	if err == nil {
		old_filepath := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + old_file_info.FileData.FileSystemName

		var found bool
		err := h.DB.QueryRow(`SELECT * FROM api_sight."playersRmPhoto"($1, $2, $3);`, player_id, club_id, user_id).Scan(&found)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		_ = os.Remove(old_filepath)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	new_file_info := FileInfo{
		OriginalfileName: file.Filename,
		FileSize:         file.Size,
		MimeType:         file.Header.Get("Content-Type"),
		FileSystemName:   uuid.New().String(),
	}

	new_filepath := strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + new_file_info.FileSystemName

	dst, err := os.Create(new_filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	js_file_info, err := json.Marshal(new_file_info)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var file_id *int
	err = h.DB.QueryRow(`SELECT * FROM api_sight."playersAddPhoto"($1, $2, $3, $4);`, player_id, string(js_file_info), club_id, user_id).Scan(&file_id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"result": true,
		"data":   file_id,
	})
}

func (h *handler) getClubLogo(c echo.Context) error {
	claims := c.Get("user").(*jwt.Token).Claims.(*UserClaims)
	club_id := int(claims.Data["club_id"].(float64))

	club_info := new(ClubInfo)
	err := h.DB.QueryRow(`SELECT * FROM api_sight."clubGetById"($1);`, club_id).Scan(&club_info.Id, &club_info.Name, &club_info.CreateTime, &club_info.UpdateTime, &club_info.Params)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	filepath := strings.TrimSuffix(h.cfg.AssetsDir, "/") + "/noimage.png"
	logo_mime := "image/png"

	switch logo_path := club_info.Params["logo_path"].(type) {
	case string:
		filepath = strings.TrimSuffix(h.cfg.FilesDir, "/") + "/" + logo_path
	}

	switch mime := club_info.Params["logo_mime"].(type) {
	case string:
		logo_mime = mime
	}

	f, err := os.Open(filepath)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.Stream(http.StatusOK, logo_mime, f)
}
