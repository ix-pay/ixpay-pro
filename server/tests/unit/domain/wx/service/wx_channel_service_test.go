package wxservice

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ix-pay/ixpay-pro/internal/domain/wx/service"
)

// TestParseDateRange_Default7Days 不传参时默认近 7 天
func TestParseDateRange_Default7Days(t *testing.T) {
	before := time.Now().AddDate(0, 0, -7)
	start, end, err := service.ParseDateRange("", "")
	require.NoError(t, err)

	assert.WithinDuration(t, before, start, 2*time.Second, "默认起始应为近 7 天前")
	assert.WithinDuration(t, time.Now(), end, 2*time.Second, "默认结束应为当前时间")
}

// TestParseDateRange_CustomRange 自定义日期范围：结束日期扩展到当天末尾（23:59:59）
func TestParseDateRange_CustomRange(t *testing.T) {
	start, end, err := service.ParseDateRange("2026-01-01", "2026-01-31")
	require.NoError(t, err)

	assert.Equal(t, "2026-01-01", start.Format("2006-01-02"))
	assert.Equal(t, "2026-01-31", end.Format("2006-01-02"))
	assert.Equal(t, 23, end.Hour(), "结束日期应扩展到当天 23 点")
	assert.Equal(t, 59, end.Minute(), "结束日期应扩展到当天 59 分")
	assert.Equal(t, 59, end.Second(), "结束日期应扩展到当天 59 秒")
}

// TestParseDateRange_OnlyStart 仅传开始日期：结束默认当前时间
func TestParseDateRange_OnlyStart(t *testing.T) {
	start, end, err := service.ParseDateRange("2026-05-01", "")
	require.NoError(t, err)

	assert.Equal(t, "2026-05-01", start.Format("2006-01-02"))
	assert.WithinDuration(t, time.Now(), end, 2*time.Second)
}

// TestParseDateRange_InvalidFormat 非法日期格式应返回错误
func TestParseDateRange_InvalidFormat(t *testing.T) {
	_, _, err := service.ParseDateRange("2026/01/01", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start_date")

	_, _, err = service.ParseDateRange("", "01-31-2026")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end_date")
}
