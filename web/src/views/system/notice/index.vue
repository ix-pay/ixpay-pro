<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.keyword"
          placeholder="请输入公告标题"
          clearable
          style="width: 192px"
          @keyup.enter="loadNoticeList"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-select v-model="searchForm.type" placeholder="公告类型" clearable style="width: 192px">
          <el-option
            v-for="item in noticeTypeOptions"
            :key="item.itemKey"
            :label="item.itemValue"
            :value="item.itemKey"
          />
        </el-select>
        <el-select v-model="searchForm.status" placeholder="状态" clearable style="width: 192px">
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="loadNoticeList">
          <el-icon>
            <Search />
          </el-icon>
          搜索
        </el-button>
        <el-button @click="resetSearch">
          <el-icon>
            <Refresh />
          </el-icon>
          重置
        </el-button>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <el-button type="primary" v-auth-btn="'system:notice:add'" @click="handleAddNotice">
          <el-icon><Plus /></el-icon>
          添加公告
        </el-button>
      </div>
    </div>

    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Bell />
          </el-icon>
          公告总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <SuccessFilled />
          </el-icon>
          启用：<span class="font-medium">{{
            noticeList.filter((n) => n.status === 1).length
          }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <CircleClose />
          </el-icon>
          禁用：<span class="font-medium">{{
            noticeList.filter((n) => n.status === 0).length
          }}</span>
        </span>
      </div>
    </div>

    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="noticeList"
        stripe
        class="w-full h-full"
        :height="'100%'"
      >
        <el-table-column prop="title" label="公告标题" width="250" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="scope">
            <el-tag
              :type="
                scope.row.type === 1
                  ? 'primary'
                  : scope.row.type === 2
                    ? 'success'
                    : scope.row.type === 3
                      ? 'warning'
                      : 'danger'
              "
              size="small"
            >
              {{
                scope.row.type === 1
                  ? '系统公告'
                  : scope.row.type === 2
                    ? '活动公告'
                    : scope.row.type === 3
                      ? '普通通知'
                      : '紧急通知'
              }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag
              :type="
                scope.row.status === 1 ? 'success' : scope.row.status === 2 ? 'info' : 'warning'
              "
              size="small"
            >
              {{ scope.row.status === 1 ? '已发布' : scope.row.status === 2 ? '已归档' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发布时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.publishTime) }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button
                v-auth-btn="'system:notice:edit'"
                type="primary"
                size="small"
                @click="handleEditNotice(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:notice:delete'"
                type="danger"
                size="small"
                @click="handleDeleteNotice(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

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

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="600px">
      <el-form ref="noticeFormRef" :model="noticeForm" :rules="formRules" label-width="100px">
        <el-form-item label="公告标题" prop="title">
          <el-input v-model="noticeForm.title" placeholder="请输入公告标题" />
        </el-form-item>
        <el-form-item label="公告类型" prop="type">
          <el-select v-model="noticeForm.type" placeholder="请选择公告类型" class="w-full">
            <el-option
              v-for="item in noticeTypeOptions"
              :key="item.itemKey"
              :label="item.itemValue"
              :value="item.itemKey"
            />
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
        <el-form-item label="发布时间" prop="publishTime">
          <el-date-picker
            v-model="noticeForm.publishTime"
            type="datetime"
            placeholder="请选择发布时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            class="w-full"
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
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { Plus, Search, Refresh, Bell, SuccessFilled, CircleClose } from '@element-plus/icons-vue'
import { getNoticeList, createNotice, updateNotice, deleteNotice } from '@/api/modules/notice'
import { getDictItemsByCode } from '@/api/modules/dict'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'NoticeManagement',
})

interface Notice {
  id: string
  title: string
  content: string
  type: number
  status: number
  publishTime: string
  createdAt: string
}

interface DictItem {
  itemKey: string
  itemValue: string
  sort: number
}

const noticeList = ref<Notice[]>([])
const loading = ref(false)
const isLoading = ref(false)
const noticeTypeOptions = ref<DictItem[]>([])

const searchForm = reactive({
  keyword: '',
  type: undefined as string | undefined,
  status: undefined as number | undefined,
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const noticeFormRef = ref<FormInstance>()

const noticeForm = reactive({
  id: '',
  title: '',
  content: '',
  type: 1,
  publishTime: '',
  status: 1,
})

const formRules = reactive({
  title: [{ required: true, message: '请输入公告标题', trigger: 'blur' }],
  content: [{ required: true, message: '请输入公告内容', trigger: 'blur' }],
  type: [{ required: true, message: '请选择公告类型', trigger: 'change' }],
})

// 加载字典数据
const loadNoticeTypeOptions = async () => {
  try {
    const res = await getDictItemsByCode('notice_type')
    noticeTypeOptions.value = res.data || []
  } catch (error) {
    console.error('加载公告类型字典失败:', error)
    noticeTypeOptions.value = []
  }
}

const loadNoticeList = async () => {
  if (isLoading.value) return

  isLoading.value = true
  loading.value = true
  try {
    const response = await getNoticeList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.keyword ? { title: searchForm.keyword } : {}),
      ...(searchForm.type ? { type: searchForm.type } : {}),
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    noticeList.value = (pageData?.list as Notice[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    console.error('获取公告列表失败:', error)
    noticeList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
    isLoading.value = false
  }
}

const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.type = undefined
  searchForm.status = undefined
  pagination.page = 1
  loadNoticeList()
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadNoticeList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadNoticeList()
}

const handleAddNotice = () => {
  dialogTitle.value = '添加公告'
  Object.assign(noticeForm, {
    id: '',
    title: '',
    content: '',
    type: 1,
    publishTime: '',
    status: 1,
  })
  dialogVisible.value = true
}

const handleEditNotice = (notice: Notice) => {
  dialogTitle.value = '编辑公告'
  Object.assign(noticeForm, notice)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    await noticeFormRef.value?.validate()
    if (noticeForm.id) {
      await updateNotice(noticeForm.id, {
        title: noticeForm.title,
        content: noticeForm.content,
        type: noticeForm.type,
        status: noticeForm.status,
        publishTime: noticeForm.publishTime,
      })
      ElMessage.success('更新成功')
    } else {
      await createNotice({
        title: noticeForm.title,
        content: noticeForm.content,
        type: noticeForm.type,
        status: noticeForm.status,
        publishTime: noticeForm.publishTime,
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadNoticeList()
  } catch {
    // 验证失败
  }
}

const handleDeleteNotice = async (id: string) => {
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
  loadNoticeTypeOptions()
  loadNoticeList()
})
</script>
