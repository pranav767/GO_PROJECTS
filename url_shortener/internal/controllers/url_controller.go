// HTTP handlers
package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"url_shortner/internal/models"
	"url_shortner/internal/services"
)

type Server struct {
	storage services.URLRepository
}

func NewServer(storage services.URLRepository) *Server {
	return &Server{storage: storage}
}

func (s *Server) HandleCreate(c *gin.Context) {
	var req models.URL
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}
	if err := s.storage.Create(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (s *Server) HandleUpdate(c *gin.Context) {
	shortCode := c.Param("shortCode")
	var req models.URL
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}
	if req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}
	updated := &models.URL{
		URL:       req.URL,
		ShortCode: shortCode,
	}
	if err := s.storage.Update(shortCode, updated); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, _ := s.storage.GetByShortCode(shortCode)
	c.JSON(http.StatusOK, item)
}

func (s *Server) HandleDelete(c *gin.Context) {
	shortCode := c.Param("shortCode")
	if err := s.storage.Delete(shortCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (s *Server) HandleGetByShortCode(c *gin.Context) {
	shortCode := c.Param("shortCode")
	item, err := s.storage.GetByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) HandleGetDetails(c *gin.Context) {
	shortCode := c.Param("shortCode")
	item, err := s.storage.GetByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *Server) HandleStats(c *gin.Context) {
	shortCode := c.Param("shortCode")
	item, err := s.storage.GetByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"shortCode": shortCode, "accessCount": item.AccessCount})
}

func (s *Server) HandleRedirect(c *gin.Context) {
	shortCode := c.Param("shortCode")
	item, err := s.storage.GetByShortCode(shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "shortCode not found"})
		return
	}
	// Increment access count
	if err := s.storage.IncrementAccessCount(shortCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not increment access count"})
		return
	}
	c.Redirect(http.StatusFound, item.URL)
}
