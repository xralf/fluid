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

var QueriesDirectoryPath string

type QueryHandler interface {
	UploadFile(*gin.Context)
	Delete(*gin.Context)
}

type queryHandler struct {
	config utils.Configuration
}

func NewQueryHandler(config utils.Configuration) QueryHandler {
	return &queryHandler{config: config}
}

func init() {
	if QueriesDirectoryPath == "" { // called for the first time, initialize the path
		QueriesDirectoryPath = filepath.Join("/", "tmp", "uploads", "queries")
		if err := os.MkdirAll(QueriesDirectoryPath, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

type Response struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

func (app *queryHandler) UploadFile(c *gin.Context) {
	var file *multipart.FileHeader
	var err error
	if file, err = c.FormFile("file"); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No file received",
		})
		return
	}

	id := uuid.NewString()
	filePath := queryFilePath(id)

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

func (app *queryHandler) Delete(c *gin.Context) {
	logger.Info("Delete: BEGIN")
	_, ctxErr := context.WithTimeout(c.Request.Context(), time.Duration(app.config.App.Timeout)*time.Second)
	defer ctxErr()

	query := model.Query{}
	if err := c.ShouldBindBodyWith(&query, binding.JSON); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusAccepted, &query)
	logger.Debug(
		"delete",
		"query", query,
	)
	//id := repository.QueryID(c.Param("id"))
	filePath := queryFilePath(query.Id)

	var err error
	if err = os.Remove(filePath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Unable to delete file %s: %s", filePath, err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, nil)
}

func queryFilePath(id string) string {
	return QueriesDirectoryPath + "/" + id + ".fql"
}
