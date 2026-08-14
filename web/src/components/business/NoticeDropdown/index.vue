<template>
  <el-dropdown :hide-on-click="false" trigger="click" @visible-change="handleDropdownVisible">
    <el-button
      type="text"
      class="transition-all duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] px-[var(--space-sm)] py-[var(--space-md)] rounded-[var(--radius-md)] hover:bg-[var(--bg-secondary)] hover:text-[var(--primary-color)]"
    >
      <el-icon>
        <Bell />
      </el-icon>
      <el-badge v-if="unreadCount > 0" :value="unreadCount" :max="99" type="danger" />
    </el-button>
    <template #dropdown>
      <el-dropdown-menu class="notice-dropdown-menu">
        <el-dropdown-item disabled>
          <div class="flex justify-between items-center">
            <span class="text-[14px] font-semibold text-[var(--text-primary)]">公告通知</span>
            <el-button
              v-if="noticeList.length > 0"
              type="primary"
              link
              size="small"
              @click="markAllAsRead"
              class="text-[12px]"
            >
              全部已读
            </el-button>
          </div>
        </el-dropdown-item>
        <el-dropdown-item v-if="noticeList.length === 0" disabled>
          <div class="flex flex-col items-center justify-center py-6 text-[var(--text-secondary)]">
            <el-icon class="text-[40px] opacity-30 mb-2">
              <Bell />
            </el-icon>
            <span class="text-[13px]">暂无公告通知</span>
          </div>
        </el-dropdown-item>
        <el-dropdown-item
          v-for="notice in noticeList"
          :key="notice.id"
          class="notice-item"
          @click="handleViewNotice(notice)"
        >
          <div class="flex flex-col gap-1 w-full">
            <div class="flex justify-between items-start">
              <span class="text-[14px] font-medium text-[var(--text-primary)] truncate max-w-[180px]">
                {{ notice.title }}
              </span>
              <el-tag v-if="!notice.isRead" size="small" type="danger" class="ml-2 shrink-0">新</el-tag>
            </div>
            <div class="text-[12px] text-[var(--text-secondary)] line-clamp-2" v-html="stripHtml(notice.content)" />
            <div class="text-[11px] text-[var(--text-placeholder)]">
              {{ formatTime(notice.publishTime) }}
            </div>
          </div>
        </el-dropdown-item>
        <el-dropdown-item v-if="noticeList.length > 0" disabled>
          <div class="text-center text-[12px] text-[var(--text-placeholder)] py-1">
            共 {{ noticeTotal }} 条公告
          </div>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>

  <!-- 公告详情弹窗 -->
  <el-dialog
    v-model="detailVisible"
    :title="currentNotice?.title || '公告详情'"
    width="650px"
    :close-on-click-modal="true"
  >
    <div class="notice-detail">
      <div class="flex items-center gap-4 mb-4 text-[13px] text-[var(--text-secondary)]">
        <span>
          <el-tag
            :type="noticeTypeTag(currentNotice?.type || '')"
            size="small"
          >
            {{ noticeTypeLabel(currentNotice?.type || '') }}
          </el-tag>
        </span>
        <span>发布时间：{{ formatTime(currentNotice?.publishTime || '') }}</span>
      </div>
      <div class="notice-detail-content" v-html="currentNotice?.content || ''" />
    </div>
    <template #footer>
      <el-button @click="detailVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Bell } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { getNoticeList, markNoticeAsRead, type Notice } from '@/api/modules/notice'

defineOptions({
  name: 'NoticeDropdown',
})

// 公告列表
const noticeList = ref<(Notice & { isRead?: boolean })[]>([])
const noticeTotal = ref(0)
const detailVisible = ref(false)
const currentNotice = ref<Notice | null>(null)

// 未读数量
const unreadCount = computed(() => {
  return noticeList.value.filter((n) => !n.isRead).length
})

// 加载公告列表（只获取已发布的）
const loadNotices = async () => {
  try {
    const res = await getNoticeList({
      page: 1,
      pageSize: 5,
      status: 1, // 只获取已发布状态
    })
    const pageData = res.data as Record<string, unknown>
    const list = (pageData?.list as Notice[]) || []
    noticeTotal.value = (pageData?.total as number) || 0

    // 从 localStorage 读取已读记录
    const readIds = getReadNoticeIds()
    noticeList.value = list.map((n) => ({
      ...n,
      isRead: readIds.includes(n.id),
    }))
  } catch {
    noticeList.value = []
    noticeTotal.value = 0
  }
}

// 获取本地已读公告ID列表
const getReadNoticeIds = (): string[] => {
  try {
    const stored = localStorage.getItem('ixpay_read_notices')
    return stored ? JSON.parse(stored) : []
  } catch {
    return []
  }
}

// 保存已读公告ID
const saveReadNoticeId = (id: string) => {
  const readIds = getReadNoticeIds()
  if (!readIds.includes(id)) {
    readIds.push(id)
    localStorage.setItem('ixpay_read_notices', JSON.stringify(readIds))
  }
}

// 查看公告详情
const handleViewNotice = async (notice: Notice & { isRead?: boolean }) => {
  currentNotice.value = notice
  detailVisible.value = true

  // 标记为已读
  if (!notice.isRead) {
    notice.isRead = true
    saveReadNoticeId(notice.id)
    try {
      await markNoticeAsRead(notice.id)
    } catch {
      // 后端标记失败不影响前端状态
    }
  }
}

// 全部已读
const markAllAsRead = async () => {
  const unreadNotices = noticeList.value.filter((n) => !n.isRead)
  for (const notice of unreadNotices) {
    notice.isRead = true
    saveReadNoticeId(notice.id)
    try {
      await markNoticeAsRead(notice.id)
    } catch {
      // 忽略单个标记失败
    }
  }
  ElMessage.success('所有公告已标记为已读')
}

// 下拉框展开时刷新数据
const handleDropdownVisible = (visible: boolean) => {
  if (visible) {
    loadNotices()
  }
}

// 去除HTML标签，获取纯文本摘要
const stripHtml = (html: string): string => {
  if (!html) return ''
  const text = html.replace(/<[^>]+>/g, '').replace(/&nbsp;/g, ' ').trim()
  return text.length > 80 ? text.slice(0, 80) + '...' : text
}

// 格式化时间
const formatTime = (time: string): string => {
  if (!time) return ''
  return time.replace('T', ' ').slice(0, 16)
}

// 公告类型标签颜色
const noticeTypeTag = (type: string): string => {
  const map: Record<string, string> = {
    system: 'primary',
    activity: 'success',
    notice: 'warning',
    urgent: 'danger',
  }
  return map[type] || 'info'
}

// 公告类型标签文字
const noticeTypeLabel = (type: string): string => {
  const map: Record<string, string> = {
    system: '系统公告',
    activity: '活动公告',
    notice: '普通通知',
    urgent: '紧急通知',
  }
  return map[type] || type
}

onMounted(() => {
  loadNotices()
})
</script>

<style scoped>
.notice-dropdown-menu {
  min-width: 320px;
  max-width: 380px;
  padding: 0;
}

.notice-item {
  padding: 10px 16px !important;
  border-bottom: 1px solid var(--border-primary);
  cursor: pointer;
  transition: background-color 0.2s;
}

.notice-item:last-of-type {
  border-bottom: none;
}

.notice-item:hover {
  background-color: var(--bg-secondary) !important;
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* 公告详情 */
.notice-detail {
  max-height: 60vh;
  overflow-y: auto;
}

.notice-detail-content {
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 8px;
  line-height: 1.8;
  color: var(--text-primary);
  word-break: break-word;
}

.notice-detail-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
}

.notice-detail-content :deep(a) {
  color: var(--primary-color);
}

.notice-detail-content :deep(blockquote) {
  border-left: 3px solid var(--primary-color);
  padding-left: 12px;
  margin: 8px 0;
  color: var(--text-secondary);
}

.notice-detail-content :deep(ul),
.notice-detail-content :deep(ol) {
  padding-left: 20px;
}

.notice-detail-content :deep(pre) {
  background: var(--bg-tertiary);
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
}
</style>