package handler

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/xralf/fluid/cmd/utils"
	"github.com/xralf/fluid/pkg/model"
)

var PrepsDirectoryPath string

type PrepHandler interface {
	UploadFile(*gin.Context)
	Delete(*gin.Context)
}

type prepHandler struct {
	config utils.Configuration
}

func NewPrepHandler(config utils.Configuration) PrepHandler {
	return &prepHandler{config: config}
}

func init() {
	if PrepsDirectoryPath == "" { // called for the first time, initialize the path
		PrepsDirectoryPath = filepath.Join("/", "tmp", "uploads", "preps")
		if err := os.MkdirAll(PrepsDirectoryPath, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func (app *prepHandler) UploadFile(c *gin.Context) {
	var file *multipart.FileHeader
	var err error
	if file, err = c.FormFile("file"); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file received",
		})
		return
	}

	id := uuid.NewString()
	filePath := prepFilePath(id)

	if err = c.SaveUploadedFile(file, filePath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Unable to save file %s", filePath),
		})
		return
	}

	resp := model.AddResponse{
		Id:      id,
		Message: fmt.Sprintf("successfully uploaded file %s", filePath),
	}
	c.JSON(http.StatusOK, resp)
}

func (app *prepHandler) Delete(c *gin.Context) {
	logger.Info("Delete: BEGIN")
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.config.App.Timeout)*time.Second)
	defer ctxErr()

	prep := model.Prep{}
	if err := c.ShouldBindBodyWith(&prep, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusAccepted, &prep)
	logger.Debug(
		"delete",
		"prep", prep,
	)
	//id := repository.PrepID(c.Param("id"))
	filePath := prepFilePath(prep.Id)

	var err error
	if err = os.Remove(filePath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Unable to delete file %s: %s", filePath, err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func prepFilePath(id string) string {
	return PrepsDirectoryPath + "/" + id + ".sh"
}
