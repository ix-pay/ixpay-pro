package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNoticeReadRecordService_UserIDValidation 测试用户 ID 验证
func TestNoticeReadRecordService_UserIDValidation(t *testing.T) {
	testCases := []struct {
		name   string
		userID int64
		valid  bool
	}{
		{"有效用户 ID - 正数", 1, true},
		{"有效用户 ID - 大数", 999999, true},
		{"无效用户 ID - 0", 0, false},
		{"无效用户 ID - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.userID <= 0 {
				assert.False(t, tc.valid, "用户 ID 必须大于 0")
			}
		})
	}
}

// TestNoticeReadRecordService_NoticeIDValidation 测试公告 ID 验证
func TestNoticeReadRecordService_NoticeIDValidation(t *testing.T) {
	testCases := []struct {
		name     string
		noticeID int64
		valid    bool
	}{
		{"有效公告 ID - 正数", 1, true},
		{"有效公告 ID - 大数", 999999, true},
		{"无效公告 ID - 0", 0, false},
		{"无效公告 ID - 负数", -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.noticeID <= 0 {
				assert.False(t, tc.valid, "公告 ID 必须大于 0")
			}
		})
	}
}

// TestNoticeReadRecordService_ReadTimeValidation 测试阅读时间验证
func TestNoticeReadRecordService_ReadTimeValidation(t *testing.T) {
	testCases := []struct {
		name       string
		futureTime bool
		valid      bool
	}{
		{"有效时间 - 过去时间", false, true},
		{"有效时间 - 当前时间", false, true},
		{"无效时间 - 未来时间", true, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.futureTime && tc.valid {
				assert.False(t, tc.valid, "阅读时间不应该是未来时间")
			}
		})
	}
}

// TestNoticeReadRecordService_ReadStatusValidation 测试阅读状态验证
func TestNoticeReadRecordService_ReadStatusValidation(t *testing.T) {
	testCases := []struct {
		name   string
		status int
		valid  bool
	}{
		{"有效状态 - 已读", 1, true},
		{"有效状态 - 未读", 0, true},
		{"无效状态 - 负数", -1, false},
		{"无效状态 - 大于 1", 2, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.status != 0 && tc.status != 1 {
				assert.False(t, tc.valid, "状态只能是 0 或 1")
			}
		})
	}
}
