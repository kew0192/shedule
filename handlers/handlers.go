package handlers

import (
	"net/http"
	"task_logic/objects"

	"github.com/gin-gonic/gin"
)

type LogicHandler struct {
	Service *objects.Service
}

func NewLogicHandler(service *objects.Service) *LogicHandler {
	return &LogicHandler{Service: service}
}

// ===== КЛАССЫ =====

func (h *LogicHandler) CreateClassHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	class, err := h.Service.Repo.CreateClass(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "class created successfully",
		"class":   class,
	})
}

func (h *LogicHandler) GetClassHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class name required"})
		return
	}

	class, err := h.Service.Repo.GetClassByName(c.Request.Context(), name)
	if err != nil || class == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}

	c.JSON(http.StatusOK, class)
}

func (h *LogicHandler) GetAllClassesHandler(c *gin.Context) {
	classes, err := h.Service.Repo.GetAllClasses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, classes)
}

func (h *LogicHandler) DeleteClassHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	class, err := h.Service.Repo.GetClassByName(c.Request.Context(), req.Name)
	if err != nil || class == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}

	err = h.Service.Repo.DeleteClass(c.Request.Context(), class.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "class deleted successfully",
	})
}

// ===== УЧИТЕЛЯ =====

func (h *LogicHandler) CreateTeacherHandler(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Lesson struct {
			Name string `json:"name" binding:"required"`
		} `json:"lesson" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacher := &objects.Teacher{
		Name: req.Name,
		Lesson: objects.Lesson{
			Name: req.Lesson.Name,
		},
	}

	err := h.Service.Repo.CreateTeacher(c.Request.Context(), teacher)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "teacher created successfully",
		"teacher": teacher,
	})
}

func (h *LogicHandler) GetTeacherHandler(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "teacher name required"})
		return
	}

	teacher, err := h.Service.Repo.GetTeacherByName(c.Request.Context(), name)
	if err != nil || teacher == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found"})
		return
	}

	c.JSON(http.StatusOK, teacher)
}

func (h *LogicHandler) DeleteTeacherHandler(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacher, err := h.Service.Repo.GetTeacherByName(c.Request.Context(), req.Name)
	if err != nil || teacher == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found"})
		return
	}

	err = h.Service.Repo.DeleteTeacher(c.Request.Context(), teacher.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "teacher deleted successfully",
	})
}

// ===== РАСПИСАНИЕ =====

func (h *LogicHandler) AddLessonsHandler(c *gin.Context) {
	var req struct {
		TeacherName string `json:"teacher_name" binding:"required"`
		ClassName   string `json:"class_name" binding:"required"`
		Subject     string `json:"subject" binding:"required"`
		Hours       int    `json:"hours" binding:"required,min=1,max=40"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teacher, err := h.Service.Repo.GetTeacherByNameAndSubject(c.Request.Context(), req.TeacherName, req.Subject)
	if err != nil || teacher == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teacher not found"})
		return
	}

	class, err := h.Service.Repo.GetClassByName(c.Request.Context(), req.ClassName)
	if err != nil || class == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}

	updatedClass, err := h.Service.AddLessons(c.Request.Context(), *teacher, *class, req.Hours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Service.Repo.UpdateClass(c.Request.Context(), updatedClass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "lessons added successfully",
		"class":   updatedClass,
	})
}

func (h *LogicHandler) GetScheduleHandler(c *gin.Context) {
	name := c.Param("class")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "class name required"})
		return
	}

	class, err := h.Service.Repo.GetClassByName(c.Request.Context(), name)
	if err != nil || class == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "class not found"})
		return
	}

	c.JSON(http.StatusOK, class)
}
