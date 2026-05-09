package publish

import (
	publishsvc "house-manager/internal/service/publish"

	"github.com/gin-gonic/gin"
)

func (h *PublishHandler) CreateCentralizedProject(c *gin.Context) {
	var req createCentralizedProjectRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	result, err := h.service.CreateCentralizedProject(c.Request.Context(), publishsvc.CreateCentralizedProjectInput{
		ProjectName: req.ProjectName,
		ProjectCode: req.ProjectCode,
		City:        req.City,
		District:    req.District,
		AddressText: req.AddressText,
		Geo:         req.Geo.toServiceInput(),
		BrandName:   req.BrandName,
	})
	writePublishResult(c, "create centralized project failed", result, err)
}

func (h *PublishHandler) CentralizedProjectDetail(c *gin.Context) {
	id, ok := bindIDRequest(c)
	if !ok {
		return
	}
	result, err := h.service.GetCentralizedProject(c.Request.Context(), id)
	writePublishResult(c, "get centralized project failed", result, err)
}

func (h *PublishHandler) ListCentralizedProjects(c *gin.Context) {
	var req listByCityRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	result, err := h.service.ListCentralizedProjects(c.Request.Context(), publishsvc.ListCentralizedProjectsInput{
		City:     req.City,
		District: req.District,
	})
	writePublishResult(c, "list centralized projects failed", result, err)
}

func (h *PublishHandler) UpdateCentralizedProject(c *gin.Context) {
	var req updateCentralizedProjectRequest
	if !bindPublishJSON(c, &req) {
		return
	}
	id, ok := objectIDFromHex(c, req.ID)
	if !ok {
		return
	}
	result, err := h.service.UpdateCentralizedProject(c.Request.Context(), publishsvc.UpdateCentralizedProjectInput{
		ID:          id,
		ProjectName: req.ProjectName,
		City:        req.City,
		District:    req.District,
		AddressText: req.AddressText,
		Geo:         req.Geo.toServiceInput(),
		BrandName:   req.BrandName,
	})
	writePublishResult(c, "update centralized project failed", result, err)
}
