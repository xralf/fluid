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
	"github.com/google/uuid"
	"github.com/xralf/fluid/cmd/utils"
	"github.com/xralf/fluid/pkg/model"
)

var CatalogsDirectoryPath string

type CatalogHandler interface {
	UploadFile(*gin.Context)
	Delete(*gin.Context)
}

type catalogHandler struct {
	config utils.Configuration
}

func NewCatalogHandler(config utils.Configuration) CatalogHandler {
	return &catalogHandler{config: config}
}

func init() {
	if CatalogsDirectoryPath == "" { // called for the first time, initialize the path
		CatalogsDirectoryPath =
			filepath.Join("/", "tmp", "uploads", "catalogs")
		if err := os.MkdirAll(CatalogsDirectoryPath, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func (app *catalogHandler) UploadFile(c *gin.Context) {
	var file *multipart.FileHeader
	var err error
	if file, err = c.FormFile("file"); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file received",
		})
		return
	}

	id := uuid.NewString()
	filePath := catalogFilePath(id)

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

func (app *catalogHandler) Delete(c *gin.Context) {
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.config.App.Timeout)*time.Second)
	defer ctxErr()

	id := "foobar"
	//id := repository.CatalogID(c.Param("id"))
	filePath := catalogFilePath(id)

	var err error
	if err = os.Remove(filePath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Unable to delete file %s", filePath),
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func catalogFilePath(id string) string {
	return CatalogsDirectoryPath + "/" + id + ".json"
}
