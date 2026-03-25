package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

func (h *AdminGrowthHandler) ListStrategySeedSets(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStrategySeedSets(c.Query("target_type"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStrategySeedSet(c *gin.Context) {
	var req dto.StrategySeedSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminCreateStrategySeedSet(model.StrategySeedSet{
		Name:        req.Name,
		TargetType:  req.TargetType,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
		Items:       req.Items,
		Description: req.Description,
		UpdatedBy:   operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "CREATE_STRATEGY_SEED_SET", "STRATEGY_SEED_SET", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateStrategySeedSet(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.StrategySeedSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateStrategySeedSet(id, model.StrategySeedSet{
		Name:        req.Name,
		TargetType:  req.TargetType,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
		Items:       req.Items,
		Description: req.Description,
		UpdatedBy:   operator,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "strategy seed set not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_STRATEGY_SEED_SET", "STRATEGY_SEED_SET", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListStrategyAgentProfiles(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStrategyAgentProfiles(c.Query("target_type"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStrategyAgentProfile(c *gin.Context) {
	var req dto.StrategyAgentProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminCreateStrategyAgentProfile(model.StrategyAgentProfile{
		Name:                            req.Name,
		TargetType:                      req.TargetType,
		Status:                          req.Status,
		IsDefault:                       req.IsDefault,
		EnabledAgents:                   req.EnabledAgents,
		PositiveThreshold:               req.PositiveThreshold,
		NegativeThreshold:               req.NegativeThreshold,
		AllowVeto:                       req.AllowVeto,
		AllowMockFallbackOnShortHistory: req.AllowMockFallbackOnShortHistory,
		Description:                     req.Description,
		UpdatedBy:                       operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "CREATE_STRATEGY_AGENT_PROFILE", "STRATEGY_AGENT_PROFILE", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateStrategyAgentProfile(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.StrategyAgentProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateStrategyAgentProfile(id, model.StrategyAgentProfile{
		Name:                            req.Name,
		TargetType:                      req.TargetType,
		Status:                          req.Status,
		IsDefault:                       req.IsDefault,
		EnabledAgents:                   req.EnabledAgents,
		PositiveThreshold:               req.PositiveThreshold,
		NegativeThreshold:               req.NegativeThreshold,
		AllowVeto:                       req.AllowVeto,
		AllowMockFallbackOnShortHistory: req.AllowMockFallbackOnShortHistory,
		Description:                     req.Description,
		UpdatedBy:                       operator,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "strategy agent profile not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_STRATEGY_AGENT_PROFILE", "STRATEGY_AGENT_PROFILE", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListStrategyScenarioTemplates(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStrategyScenarioTemplates(c.Query("target_type"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStrategyScenarioTemplate(c *gin.Context) {
	var req dto.StrategyScenarioTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	items := make([]model.StrategyScenarioTemplateItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.StrategyScenarioTemplateItem{
			Scenario:       item.Scenario,
			Label:          item.Label,
			ThesisTemplate: item.ThesisTemplate,
			Action:         item.Action,
			RiskSignal:     item.RiskSignal,
			ScoreBias:      item.ScoreBias,
		})
	}
	id, err := h.service.AdminCreateStrategyScenarioTemplate(model.StrategyScenarioTemplate{
		Name:        req.Name,
		TargetType:  req.TargetType,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
		Items:       items,
		Description: req.Description,
		UpdatedBy:   operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "CREATE_STRATEGY_SCENARIO_TEMPLATE", "STRATEGY_SCENARIO_TEMPLATE", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateStrategyScenarioTemplate(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.StrategyScenarioTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	items := make([]model.StrategyScenarioTemplateItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, model.StrategyScenarioTemplateItem{
			Scenario:       item.Scenario,
			Label:          item.Label,
			ThesisTemplate: item.ThesisTemplate,
			Action:         item.Action,
			RiskSignal:     item.RiskSignal,
			ScoreBias:      item.ScoreBias,
		})
	}
	if err := h.service.AdminUpdateStrategyScenarioTemplate(id, model.StrategyScenarioTemplate{
		Name:        req.Name,
		TargetType:  req.TargetType,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
		Items:       items,
		Description: req.Description,
		UpdatedBy:   operator,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "strategy scenario template not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_STRATEGY_SCENARIO_TEMPLATE", "STRATEGY_SCENARIO_TEMPLATE", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListStrategyPublishPolicies(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStrategyPublishPolicies(c.Query("target_type"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStrategyPublishPolicy(c *gin.Context) {
	var req dto.StrategyPublishPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminCreateStrategyPublishPolicy(model.StrategyPublishPolicy{
		Name:                 req.Name,
		TargetType:           req.TargetType,
		Status:               req.Status,
		IsDefault:            req.IsDefault,
		MaxRiskLevel:         req.MaxRiskLevel,
		MaxWarningCount:      req.MaxWarningCount,
		AllowVetoedPublish:   req.AllowVetoedPublish,
		DefaultPublisher:     req.DefaultPublisher,
		OverrideNoteTemplate: req.OverrideNoteTemplate,
		Description:          req.Description,
		UpdatedBy:            operator,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "CREATE_STRATEGY_PUBLISH_POLICY", "STRATEGY_PUBLISH_POLICY", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateStrategyPublishPolicy(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var req dto.StrategyPublishPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateStrategyPublishPolicy(id, model.StrategyPublishPolicy{
		Name:                 req.Name,
		TargetType:           req.TargetType,
		Status:               req.Status,
		IsDefault:            req.IsDefault,
		MaxRiskLevel:         req.MaxRiskLevel,
		MaxWarningCount:      req.MaxWarningCount,
		AllowVetoedPublish:   req.AllowVetoedPublish,
		DefaultPublisher:     req.DefaultPublisher,
		OverrideNoteTemplate: req.OverrideNoteTemplate,
		Description:          req.Description,
		UpdatedBy:            operator,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "strategy publish policy not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_STRATEGY_PUBLISH_POLICY", "STRATEGY_PUBLISH_POLICY", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListStrategyEngineJobs(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStrategyEngineJobs(c.Query("job_type"), c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) GetStrategyEngineJob(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("job_id"))
	if jobID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "job_id is required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminGetStrategyEngineJob(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "strategy engine job not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) PublishStrategyEngineJob(c *gin.Context) {
	jobID := strings.TrimSpace(c.Param("job_id"))
	if jobID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "job_id is required", Data: struct{}{}})
		return
	}
	var req dto.StrategyJobPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	record, err := h.service.AdminPublishStrategyEngineJob(jobID, operator, req.Force, req.OverrideReason)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "strategy engine job not found", Data: struct{}{}})
			return
		}
		if detail, ok := extractStrategyPublishConflictDetail(err); ok {
			c.JSON(http.StatusConflict, dto.APIResponse{
				Code:    40901,
				Message: detail,
				Data: gin.H{
					"conflict_type": "PUBLISH_POLICY_BLOCKED",
					"detail":        detail,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	action := "PUBLISH_STRATEGY_ENGINE_JOB"
	if req.Force {
		action = "OVERRIDE_PUBLISH_STRATEGY_ENGINE_JOB"
	}
	h.writeOperationLog(c, "MARKET", action, "STRATEGY_JOB", jobID, "", record.PublishID, strings.TrimSpace(req.OverrideReason))
	eventType := "STRATEGY_JOB_PUBLISHED"
	level := "INFO"
	title := "策略发布完成"
	if req.Force {
		eventType = "STRATEGY_JOB_FORCE_PUBLISHED"
		level = "WARNING"
		title = "策略人工覆盖发布"
	}
	h.writeAuditEvent(c, model.AdminAuditEvent{
		EventDomain: "PUBLISH",
		EventType:   eventType,
		Level:       level,
		Module:      "STRATEGY_ENGINE",
		ObjectType:  "STRATEGY_JOB",
		ObjectID:    jobID,
		Title:       title,
		Summary:     "策略作业 " + jobID + " 已生成发布记录 " + record.PublishID,
		Detail:      strings.TrimSpace(req.OverrideReason),
		Status:      "OPEN",
		Metadata: map[string]any{
			"publish_id":       record.PublishID,
			"job_type":         record.JobType,
			"publish_version":  record.Version,
			"force_publish":    req.Force,
			"override_reason":  strings.TrimSpace(req.OverrideReason),
			"selected_count":   record.SelectedCount,
			"warning_count":    record.Replay.WarningCount,
			"storage_source":   record.Replay.StorageSource,
		},
	})
	c.JSON(http.StatusOK, dto.OK(record))
}
