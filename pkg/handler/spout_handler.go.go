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

var SpoutsDirectoryPath string

type SpoutHandler interface {
	UploadFile(*gin.Context)
	Delete(*gin.Context)
}

type spoutHandler struct {
	config utils.Configuration
}

func NewSpoutHandler(config utils.Configuration) SpoutHandler {
	return &spoutHandler{config: config}
}

func init() {
	if SpoutsDirectoryPath == "" { // called for the first time, initialize the path
		SpoutsDirectoryPath = filepath.Join("/", "tmp", "uploads", "spouts")
		if err := os.MkdirAll(SpoutsDirectoryPath, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func (app *spoutHandler) UploadFile(c *gin.Context) {
	var file *multipart.FileHeader
	var err error
	if file, err = c.FormFile("file"); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file received",
		})
		return
	}

	id := uuid.NewString()
	filePath := spoutFilePath(id)

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

func (app *spoutHandler) Delete(c *gin.Context) {
	logger.Info("Delete: BEGIN")
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.config.App.Timeout)*time.Second)
	defer ctxErr()

	spout := model.Spout{}
	if err := c.ShouldBindBodyWith(&spout, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusAccepted, &spout)
	logger.Debug(
		"delete",
		"spout", spout,
	)
	//id := repository.SpoutID(c.Param("id"))
	filePath := spoutFilePath(spout.Id)

	var err error
	if err = os.Remove(filePath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Unable to delete file %s: %s", filePath, err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func spoutFilePath(id string) string {
	return SpoutsDirectoryPath + "/" + id + ".cmd"
}
