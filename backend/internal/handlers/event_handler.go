package handlers

import (
	"net/http"
	"strconv"

	"BetKZ/internal/repository"
	"BetKZ/internal/service"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService *service.EventService
}

func NewEventHandler(eventService *service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// GET /api/sports
func (h *EventHandler) ListSports(c *gin.Context) {
	sports, err := h.eventService.ListSports(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sports"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sports": sports})
}

// GET /api/events
func (h *EventHandler) ListEvents(c *gin.Context) {
	filters := repository.EventFilters{
		Page:  1,
		Limit: 20,
	}

	if p := c.Query("page"); p != "" {
		if page, err := strconv.Atoi(p); err == nil {
			filters.Page = page
		}
	}
	if l := c.Query("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filters.Limit = limit
		}
	}
	if s := c.Query("sport_id"); s != "" {
		if sportID, err := strconv.Atoi(s); err == nil {
			filters.SportID = &sportID
		}
	}
	if s := c.Query("status"); s != "" {
		filters.Status = &s
	}
	if s := c.Query("search"); s != "" {
		filters.Search = &s
	}
	if s := c.Query("date_from"); s != "" {
		filters.DateFrom = &s
	}
	if s := c.Query("date_to"); s != "" {
		filters.DateTo = &s
	}

	result, err := h.eventService.ListEvents(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	event, err := h.eventService.GetEvent(c.Request.Context(), idStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// POST /api/admin/events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req service.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventService.CreateEvent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, event)
}

// PUT /api/admin/events/:id
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	var req service.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.eventService.UpdateEvent(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "event updated"})
}

// PATCH /api/admin/events/:id/status
func (h *EventHandler) UpdateEventStatus(c *gin.Context) {
	var req service.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.eventService.UpdateEventStatus(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// DELETE /api/admin/events/:id
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	err := h.eventService.DeleteEvent(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "event cancelled"})
}
