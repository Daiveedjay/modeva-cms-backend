package order_controller

import (
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/Modeva-Ecommerce/modeva-cms-backend/config"
	"github.com/Modeva-Ecommerce/modeva-cms-backend/models"
	"github.com/gin-gonic/gin"
)

// GetOrderStats godoc
// @Summary Get order stats (CMS)
// @Description Returns all-time total orders + per-status breakdown, plus current month total and % change vs last month.
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ApiResponse{data=models.OrderStatsResponse}
// @Failure 401 {object} models.ApiResponse "Unauthorized"
// @Failure 403 {object} models.ApiResponse "Forbidden"
// @Failure 500 {object} models.ApiResponse "Internal server error"
// @Router /admin/orders/stats [get]
func GetOrderStats(c *gin.Context) {
	log.Printf("[admin.order.stats] start path=%s method=%s rawQuery=%s", c.FullPath(), c.Request.Method, c.Request.URL.RawQuery)

	// Optional admin guard (keep if your middleware sets it)
	if v, ok := c.Get("isAdmin"); ok {
		log.Printf("[admin.order.stats] ctx isAdmin=%v (type=%T)", v, v)
		if isAdmin, ok2 := v.(bool); ok2 && !isAdmin {
			log.Printf("[admin.order.stats] forbidden: isAdmin=false")
			c.JSON(http.StatusForbidden, models.ErrorResponse(c, "Forbidden"))
			return
		}
	} else {
		log.Printf("[admin.order.stats] ctx isAdmin missing (confirm admin middleware sets this)")
	}

	ctx, cancel := config.WithTimeout()
	defer cancel()

	// All-time totals + all-time breakdown, but month-over-month change from monthly totals
	q := `
		WITH
		all_time AS (
			SELECT
				COUNT(*)::int AS total,
				COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0)::int    AS pending,
				COALESCE(SUM(CASE WHEN status = 'processing' THEN 1 ELSE 0 END), 0)::int AS processing,
				COALESCE(SUM(CASE WHEN status = 'shipped' THEN 1 ELSE 0 END), 0)::int    AS shipped,
				COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0)::int  AS completed,
				COALESCE(SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END), 0)::int  AS cancelled
			FROM orders
		),
		cur AS (
			SELECT COUNT(*)::int AS total
			FROM orders
			WHERE created_at >= date_trunc('month', NOW())
			  AND created_at <  date_trunc('month', NOW()) + INTERVAL '1 month'
		),
		prev AS (
			SELECT COUNT(*)::int AS total
			FROM orders
			WHERE created_at >= date_trunc('month', NOW()) - INTERVAL '1 month'
			  AND created_at <  date_trunc('month', NOW())
		)
		SELECT
			all_time.total,
			cur.total,
			prev.total,
			all_time.pending,
			all_time.processing,
			all_time.shipped,
			all_time.completed,
			all_time.cancelled
		FROM all_time, cur, prev;
	`

	log.Printf("[admin.order.stats] sql=%s", strings.ReplaceAll(q, "\n", " "))

	var totalAllTime, curTotal, prevTotal int
	var pending, processing, shipped, completed, cancelled int

	err := config.EcommerceGorm.WithContext(ctx).Raw(q).Row().Scan(
		&totalAllTime,
		&curTotal,
		&prevTotal,
		&pending,
		&processing,
		&shipped,
		&completed,
		&cancelled,
	)
	if err != nil {
		log.Printf("[admin.order.stats] ERROR query failed err=%v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(c, "Failed to fetch order stats"))
		return
	}

	var changePct *float64
	if prevTotal > 0 {
		v := (float64(curTotal-prevTotal) / float64(prevTotal)) * 100
		v = math.Round(v*10) / 10
		changePct = &v
	} else {
		// If last month was 0, percent change is undefined
		changePct = nil
	}

	res := models.OrderStatsResponse{
		TotalOrders:                totalAllTime,
		ChangePercentFromLastMonth: changePct,
		CurrentMonthTotal:          curTotal,
		LastMonthTotal:             prevTotal,
		Pending: models.OrderStatsBreakdown{
			Count:       pending,
			Description: "Awaiting processing",
		},
		Processing: models.OrderStatsBreakdown{
			Count:       processing,
			Description: "Being prepared",
		},
		Shipped: models.OrderStatsBreakdown{
			Count:       shipped,
			Description: "On the way",
		},
		Completed: models.OrderStatsBreakdown{
			Count:       completed,
			Description: "Successfully delivered",
		},
		Cancelled: models.OrderStatsBreakdown{
			Count:       cancelled,
			Description: "Cancelled orders",
		},
	}

	log.Printf("[admin.order.stats] done totalAllTime=%d cur=%d prev=%d changePct=%v pending=%d processing=%d shipped=%d completed=%d cancelled=%d",
		totalAllTime, curTotal, prevTotal, changePct, pending, processing, shipped, completed, cancelled)

	c.JSON(http.StatusOK, models.SuccessResponse(
		c,
		"Order stats retrieved successfully",
		res,
	))
}
