<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow transition-colors duration-300"
  >
    <!-- 顶部操作栏 - 紧凑布局 -->
    <div
      class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <span class="text-sm text-gray-600 dark:text-gray-400">公告列表</span>
      <el-button
        type="primary"
        size="small"
        v-auth-btn="'system:notice:add'"
        @click="handleAddNotice"
      >
        <el-icon><Plus /></el-icon>
        添加公告
      </el-button>
    </div>

    <!-- 表格区域 - 占满剩余空间 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="noticeList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="title" label="公告标题" width="250" />
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="publishTime" label="发布时间" width="160" />
        <el-table-column prop="createdAt" label="创建时间" width="160" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <div class="flex gap-1">
              <el-button
                v-auth-btn="'system:notice:edit'"
                size="small"
                type="primary"
                link
                @click="handleEditNotice(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:notice:delete'"
                size="small"
                type="danger"
                link
                @click="handleDeleteNotice(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页区域 - 紧凑布局 -->
    <div
      class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-700"
    >
      <span class="text-sm text-gray-600 dark:text-gray-400">共 {{ pagination.total }} 条</span>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        layout="sizes, prev, pager, next"
        :total="pagination.total"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        small
      />
    </div>

    <!-- 公告表单对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="noticeFormRef" :model="noticeForm" :rules="formRules" label-width="100px">
        <el-form-item label="公告标题" prop="title">
          <el-input v-model="noticeForm.title" placeholder="请输入公告标题" />
        </el-form-item>
        <el-form-item label="公告类型" prop="type">
          <el-select v-model="noticeForm.type" placeholder="请选择公告类型">
            <el-option label="系统公告" value="system" />
            <el-option label="活动公告" value="activity" />
            <el-option label="维护公告" value="maintenance" />
          </el-select>
        </el-form-item>
        <el-form-item label="公告内容" prop="content">
          <el-input
            v-model="noticeForm.content"
            type="textarea"
            :rows="8"
            placeholder="请输入公告内容"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="noticeForm.status"
            :active-value="1"
            :inactive-value="0"
            active-color="#13ce66"
            inactive-color="#ff4949"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getNoticeList, createNotice, updateNotice, deleteNotice } from '@/api/modules/notice'

defineOptions({
  name: 'NoticeManagement',
})

interface Notice {
  id: number
  title: string
  content: string
  type: string
  status: number
  publishTime: string
  createdAt: string
}

const noticeList = ref<Notice[]>([])
const loading = ref(false)
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})
const dialogVisible = ref(false)
const dialogTitle = ref('')
const noticeFormRef = ref()
const noticeForm = reactive({
  id: 0,
  title: '',
  content: '',
  type: 'system',
  status: 1,
})
const formRules = reactive({
  title: [{ required: true, message: '请输入公告标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入公告内容', trigger: 'blur' }],
  type: [{ required: true, message: '请选择公告类型', trigger: 'change' }],
})

// 加载公告列表
const loadNoticeList = async () => {
  loading.value = true
  try {
    const response = await getNoticeList({
      page: pagination.page,
      pageSize: pagination.pageSize,
    })
    const pageData = response.data as Record<string, unknown>
    noticeList.value = (pageData?.list as Notice[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    ElMessage.error('获取公告列表失败')
    console.error('获取公告列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 分页大小变化
const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadNoticeList()
}

// 当前页码变化
const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadNoticeList()
}

// 添加公告
const handleAddNotice = () => {
  dialogTitle.value = '添加公告'
  Object.assign(noticeForm, { id: 0, title: '', content: '', type: 'system', status: 1 })
  dialogVisible.value = true
}

// 编辑公告
const handleEditNotice = (notice: Notice) => {
  dialogTitle.value = '编辑公告'
  Object.assign(noticeForm, notice)
  dialogVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  try {
    await noticeFormRef.value.validate()
    if (noticeForm.id) {
      await updateNotice(noticeForm.id, {
        title: noticeForm.title,
        content: noticeForm.content,
        type: noticeForm.type,
        status: noticeForm.status,
      })
    } else {
      await createNotice({
        title: noticeForm.title,
        content: noticeForm.content,
        type: noticeForm.type,
        status: noticeForm.status,
      })
    }
    ElMessage.success(noticeForm.id ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadNoticeList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除公告
const handleDeleteNotice = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除该公告吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await deleteNotice(id)
    ElMessage.success('删除成功')
    loadNoticeList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除公告失败:', error)
    }
  }
}

onMounted(() => {
  loadNoticeList()
})
</script>
