package handler

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/repo"
)

type sectorsCacheData struct {
	Sectors   []map[string]interface{}
	UpdatedAt time.Time
}

var (
	sectorsCache   *sectorsCacheData
	sectorsCacheMu sync.RWMutex
)

var defaultMockSectors = []map[string]interface{}{
	{"name": "半导体", "change": 3.82, "flow": 85},
	{"name": "人工智能", "change": 3.45, "flow": 78},
	{"name": "新能源", "change": 2.68, "flow": 72},
	{"name": "消费电子", "change": 2.12, "flow": 65},
	{"name": "医药生物", "change": 1.56, "flow": 48},
	{"name": "金融", "change": 0.85, "flow": 35},
}

// ListPublicSectors 返回今日热门板块数据
func (h *UserGrowthHandler) ListPublicSectors(c *gin.Context) {
	sectorsCacheMu.RLock()
	if sectorsCache != nil && time.Since(sectorsCache.UpdatedAt) < 1*time.Hour {
		data := sectorsCache.Sectors
		sectorsCacheMu.RUnlock()
		c.JSON(http.StatusOK, dto.OK(data))
		return
	}
	sectorsCacheMu.RUnlock()

	sectorsCacheMu.Lock()
	defer sectorsCacheMu.Unlock()

	// 双重检查
	if sectorsCache != nil && time.Since(sectorsCache.UpdatedAt) < 1*time.Hour {
		c.JSON(http.StatusOK, dto.OK(sectorsCache.Sectors))
		return
	}

	token := strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
	if token == "" && h.service != nil {
		items, _, err := h.service.AdminListSystemConfigs("tushare.token", 1, 1)
		if err == nil && len(items) > 0 {
			token = strings.TrimSpace(items[0].ConfigValue)
		}
	}

	if token == "" {
		log.Println("[Warning] TUSHARE_TOKEN not configured in env or system_configs, falling back to mock sectors")
		h.updateCache(defaultMockSectors)
		c.JSON(http.StatusOK, dto.OK(defaultMockSectors))
		return
	}

	liveSectors, err := repo.FetchShenWanSectorsFromTushare(token)
	if err != nil {
		log.Printf("[Error] Failed to fetch sectors from Tushare: %v, falling back to mock sectors", err)
		h.updateCache(defaultMockSectors)
		c.JSON(http.StatusOK, dto.OK(defaultMockSectors))
		return
	}

	// 只要拿回来的数据不为空，就限制最多 6 个作为热门板块返回
	if len(liveSectors) > 0 {
		if len(liveSectors) > 6 {
			liveSectors = liveSectors[:6]
		}
		h.updateCache(liveSectors)
		c.JSON(http.StatusOK, dto.OK(liveSectors))
	} else {
		h.updateCache(defaultMockSectors)
		c.JSON(http.StatusOK, dto.OK(defaultMockSectors))
	}
}

func (h *UserGrowthHandler) updateCache(data []map[string]interface{}) {
	sectorsCache = &sectorsCacheData{
		Sectors:   data,
		UpdatedAt: time.Now(),
	}
}
