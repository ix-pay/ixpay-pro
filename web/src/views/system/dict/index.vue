<template>
  <div
    class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md transition-colors duration-300"
  >
    <!-- 搜索栏 -->
    <div class="flex flex-col gap-3 p-4 border-b">
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.keyword"
          placeholder="请输入字典名称或编码"
          clearable
          style="width: 192px"
          @keyup.enter="loadDictList"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-select
          v-model="searchForm.status"
          placeholder="选择状态"
          clearable
          style="width: 192px"
          @change="loadDictList"
        >
          <el-option label="启用" :value="1" />
          <el-option label="禁用" :value="0" />
        </el-select>
        <el-button type="primary" @click="loadDictList" :loading="loading">
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
        <el-button type="primary" v-auth-btn="'system:dict:add'" @click="handleAddDict">
          <el-icon>
            <Plus />
          </el-icon>
          添加字典
        </el-button>
      </div>
    </div>

    <!-- 统计信息 -->
    <div
      class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex items-center gap-6 text-sm">
        <span class="flex items-center gap-1">
          <el-icon class="text-blue-500">
            <Document />
          </el-icon>
          字典总数：<span class="font-medium">{{ pagination.total }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <SuccessFilled />
          </el-icon>
          启用：<span class="font-medium">{{ dictList.filter((d) => d.status === 1).length }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <CircleClose />
          </el-icon>
          禁用：<span class="font-medium">{{ dictList.filter((d) => d.status === 0).length }}</span>
        </span>
      </div>
    </div>

    <!-- 字典表格 -->
    <div class="flex-1 overflow-hidden">
      <el-table v-loading="loading" :data="dictList" stripe class="w-full h-full" :height="'100%'">
        <el-table-column prop="dictName" label="字典名称" min-width="150" sortable="custom" />
        <el-table-column prop="dictCode" label="字典编码" min-width="150" sortable="custom" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              @change="handleStatusChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="itemCount" label="明细数量" width="100" align="center" />
        <el-table-column label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="scope">
            <div class="flex flex-wrap gap-2">
              <el-button type="primary" size="small" @click="openDictItems(scope.row)">
                <el-icon>
                  <List />
                </el-icon>
                管理明细
              </el-button>
              <el-button
                v-auth-btn="'system:dict:edit'"
                type="primary"
                size="small"
                @click="handleEditDict(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:dict:delete'"
                type="danger"
                size="small"
                @click="handleDeleteDict(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 分页 -->
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

    <!-- 右侧抽屉：字典明细管理 -->
    <el-drawer
      v-model="drawerVisible"
      :title="`${selectedDict?.dictName} - 明细管理`"
      size="65%"
      :close-on-click-modal="false"
    >
      <template #header>
        <div class="flex items-center justify-between w-full pr-4">
          <div class="flex items-center gap-3">
            <span class="text-lg font-medium">{{ selectedDict?.dictName }}</span>
            <el-tag>{{ selectedDict?.dictCode }}</el-tag>
            <el-tag v-if="selectedDict?.status === 1" type="success" size="small">启用</el-tag>
            <el-tag v-else type="danger" size="small">禁用</el-tag>
          </div>
        </div>
      </template>

      <!-- 明细操作栏 -->
      <div class="flex items-center justify-between mb-4">
        <div class="flex items-center gap-3">
          <el-input
            v-model="itemSearchForm.keyword"
            placeholder="搜索标签或值"
            clearable
            class="w-48"
            @keyup.enter="loadDictItems"
          >
            <template #prefix>
              <el-icon>
                <Search />
              </el-icon>
            </template>
          </el-input>
          <el-button type="primary" @click="loadDictItems">
            <el-icon>
              <Search />
            </el-icon>
            搜索
          </el-button>
        </div>
        <el-button type="primary" v-auth-btn="'system:dict:item:add'" @click="handleAddDictItem">
          <el-icon>
            <Plus />
          </el-icon>
          添加明细
        </el-button>
      </div>

      <!-- 统计信息 -->
      <div
        class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 mb-4"
      >
        <div class="flex items-center gap-6 text-sm">
          <span class="flex items-center gap-1">
            <el-icon class="text-blue-500">
              <List />
            </el-icon>
            明细总数：<span class="font-medium">{{ itemPagination.total }}</span>
          </span>
          <span class="flex items-center gap-1">
            <el-icon class="text-green-500">
              <SuccessFilled />
            </el-icon>
            启用：<span class="font-medium">{{
              dictItemList.filter((i) => i.status === 1).length
            }}</span>
          </span>
          <span class="flex items-center gap-1">
            <el-icon class="text-orange-500">
              <CircleClose />
            </el-icon>
            禁用：<span class="font-medium">{{
              dictItemList.filter((i) => i.status === 0).length
            }}</span>
          </span>
        </div>
      </div>

      <!-- 明细表格 -->
      <el-table
        v-loading="itemLoading"
        :data="dictItemList"
        stripe
        border
        class="w-full"
        :height="'calc(100vh - 280px)'"
      >
        <el-table-column prop="itemKey" label="标签" min-width="120" />
        <el-table-column prop="itemValue" label="值" min-width="120" />
        <el-table-column prop="sort" label="排序" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              size="small"
              @change="handleItemStatusChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="scope">
            <div class="flex flex-wrap gap-2 justify-center">
              <el-button
                v-auth-btn="'system:dict:item:edit'"
                type="primary"
                size="small"
                @click="handleEditDictItem(scope.row)"
              >
                编辑
              </el-button>
              <el-button
                v-auth-btn="'system:dict:item:delete'"
                type="danger"
                size="small"
                @click="handleDeleteDictItem(scope.row.id)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 明细分页 -->
      <div class="flex items-center justify-end mt-4">
        <el-pagination
          v-model:current-page="itemPagination.page"
          v-model:page-size="itemPagination.pageSize"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          :total="itemPagination.total"
          @size-change="handleItemSizeChange"
          @current-change="handleItemCurrentChange"
          small
        />
      </div>
    </el-drawer>

    <!-- 字典类型表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form ref="dictFormRef" :model="dictForm" :rules="formRules" label-width="100px">
        <el-form-item label="字典名称" prop="dictName">
          <el-input v-model="dictForm.dictName" placeholder="请输入字典名称" />
        </el-form-item>
        <el-form-item label="字典编码" prop="dictCode">
          <el-input
            v-model="dictForm.dictCode"
            placeholder="请输入字典编码"
            :disabled="!!dictForm.id"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="dictForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入字典描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="dictForm.status"
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

    <!-- 字典明细表单对话框 -->
    <el-dialog
      v-model="itemDialogVisible"
      :title="itemDialogTitle"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form ref="itemFormRef" :model="itemForm" :rules="itemFormRules" label-width="100px">
        <el-form-item label="标签" prop="itemKey">
          <el-input v-model="itemForm.itemKey" placeholder="请输入标签（显示文本）" />
        </el-form-item>
        <el-form-item label="值" prop="itemValue">
          <el-input v-model="itemForm.itemValue" placeholder="请输入值（存储值）" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="itemForm.sort" :min="0" :max="999" class="w-full" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="itemForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="itemForm.status"
            :active-value="1"
            :inactive-value="0"
            active-color="#13ce66"
            inactive-color="#ff4949"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="itemDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleItemSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import {
  Plus,
  Search,
  Refresh,
  List,
  Document,
  SuccessFilled,
  CircleClose,
} from '@element-plus/icons-vue'
import {
  getDictList,
  createDict,
  updateDict,
  deleteDict,
  getDictItemsByDictId,
  createDictItem,
  updateDictItem,
  deleteDictItem,
} from '@/api/modules/dict'
import { formatDate } from '@/utils/format'

defineOptions({
  name: 'DictManagement',
})

interface Dict {
  id: string
  dictName: string
  dictCode: string
  description: string
  status: number
  itemCount?: number
  createdAt: string
  updatedAt: string
}

interface DictItem {
  id: string
  dictId: string
  itemKey: string
  itemValue: string
  sort: number
  description: string
  status: number
  createdAt: string
  updatedAt: string
}

// 字典列表状态
const dictList = ref<Dict[]>([])
const loading = ref(false)

const searchForm = reactive({
  keyword: '',
  status: undefined as number | undefined,
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 字典表单
const dialogVisible = ref(false)
const dialogTitle = ref('')
const dictFormRef = ref<FormInstance>()
const dictForm = reactive({
  id: '',
  dictName: '',
  dictCode: '',
  description: '',
  status: 1,
})

const formRules = reactive({
  dictName: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
  dictCode: [{ required: true, message: '请输入字典编码', trigger: 'blur' }],
})

// 抽屉和明细状态
const drawerVisible = ref(false)
const selectedDict = ref<Dict | null>(null)
const dictItemList = ref<DictItem[]>([])
const itemLoading = ref(false)

const itemSearchForm = reactive({
  keyword: '',
})

const itemPagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
})

// 明细表单
const itemDialogVisible = ref(false)
const itemDialogTitle = ref('')
const itemFormRef = ref<FormInstance>()
const itemForm = reactive({
  id: '',
  dictId: '',
  itemKey: '',
  itemValue: '',
  sort: 0,
  description: '',
  status: 1,
})

const itemFormRules = reactive({
  itemKey: [{ required: true, message: '请输入标签', trigger: 'blur' }],
  itemValue: [{ required: true, message: '请输入值', trigger: 'blur' }],
})

const loadDictList = async () => {
  loading.value = true
  try {
    const response = await getDictList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...(searchForm.keyword ? { keyword: searchForm.keyword } : {}),
      ...(searchForm.status !== undefined ? { status: searchForm.status } : {}),
    })
    const pageData = response.data as Record<string, unknown>
    dictList.value = (pageData?.list as Dict[]) || []
    pagination.total = (pageData?.total as number) || 0
  } catch (error) {
    console.error('获取字典列表失败:', error)
    dictList.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.status = undefined
  pagination.page = 1
  loadDictList()
}

const handleSizeChange = (val: number) => {
  pagination.pageSize = val
  pagination.page = 1
  loadDictList()
}

const handleCurrentChange = (val: number) => {
  pagination.page = val
  loadDictList()
}

// 状态切换
const handleStatusChange = async (dict: Dict) => {
  try {
    await updateDict(dict.id, {
      dictName: dict.dictName,
      dictCode: dict.dictCode,
      description: dict.description,
      status: dict.status,
    })
    ElMessage.success(`字典${dict.status === 1 ? '已启用' : '已禁用'}`)
    loadDictList()
  } catch (error) {
    dict.status = dict.status === 1 ? 0 : 1
    ElMessage.error('状态更新失败')
    console.error('更新字典状态失败:', error)
  }
}

// 打开明细抽屉
const openDictItems = (dict: Dict) => {
  selectedDict.value = dict
  itemPagination.page = 1
  itemPagination.pageSize = 10
  itemPagination.total = 0
  itemSearchForm.keyword = ''
  loadDictItems()
  drawerVisible.value = true
}

// 加载字典明细
const loadDictItems = async () => {
  if (!selectedDict.value?.id) {
    dictItemList.value = []
    return
  }

  itemLoading.value = true
  try {
    const response = await getDictItemsByDictId(selectedDict.value.id)
    const data = response.data as { list: DictItem[]; total: number }
    dictItemList.value = data?.list || []
    itemPagination.total = data?.total || 0
  } catch (error) {
    console.error('获取字典明细失败:', error)
    dictItemList.value = []
    itemPagination.total = 0
  } finally {
    itemLoading.value = false
  }
}

const handleItemSizeChange = (val: number) => {
  itemPagination.pageSize = val
  itemPagination.page = 1
  loadDictItems()
}

const handleItemCurrentChange = (val: number) => {
  itemPagination.page = val
  loadDictItems()
}

// 明细状态切换
const handleItemStatusChange = async (item: DictItem) => {
  try {
    await updateDictItem(item.id, {
      dictId: item.dictId,
      itemKey: item.itemKey,
      itemValue: item.itemValue,
      sort: item.sort,
      description: item.description,
      status: item.status,
    })
    ElMessage.success(`明细${item.status === 1 ? '已启用' : '已禁用'}`)
  } catch (error) {
    item.status = item.status === 1 ? 0 : 1
    ElMessage.error('状态更新失败')
    console.error('更新明细状态失败:', error)
  }
}

// 字典 CRUD
const handleAddDict = () => {
  dialogTitle.value = '添加字典'
  Object.assign(dictForm, { id: '', dictName: '', dictCode: '', description: '', status: 1 })
  dialogVisible.value = true
}

const handleEditDict = (dict: Dict) => {
  dialogTitle.value = '编辑字典'
  Object.assign(dictForm, dict)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  try {
    await dictFormRef.value?.validate()
    if (dictForm.id) {
      await updateDict(dictForm.id, {
        dictName: dictForm.dictName,
        dictCode: dictForm.dictCode,
        description: dictForm.description,
        status: dictForm.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createDict({
        dictName: dictForm.dictName,
        dictCode: dictForm.dictCode,
        description: dictForm.description,
        status: dictForm.status,
      })
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadDictList()
  } catch {
    // 验证失败
  }
}

const handleDeleteDict = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该字典吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteDict(id)
    ElMessage.success('删除成功')
    if (selectedDict.value?.id === id) {
      drawerVisible.value = false
      selectedDict.value = null
    }
    loadDictList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除字典失败:', error)
    }
  }
}

// 明细 CRUD
const handleAddDictItem = () => {
  if (!selectedDict.value) return
  itemDialogTitle.value = '添加字典明细'
  Object.assign(itemForm, {
    id: '',
    dictId: selectedDict.value.id,
    itemKey: '',
    itemValue: '',
    sort: 0,
    description: '',
    status: 1,
  })
  itemDialogVisible.value = true
}

const handleEditDictItem = (item: DictItem) => {
  itemDialogTitle.value = '编辑字典明细'
  Object.assign(itemForm, item)
  itemDialogVisible.value = true
}

const handleItemSubmit = async () => {
  try {
    await itemFormRef.value?.validate()
    if (itemForm.id) {
      await updateDictItem(itemForm.id, {
        dictId: itemForm.dictId,
        itemKey: itemForm.itemKey,
        itemValue: itemForm.itemValue,
        sort: itemForm.sort,
        description: itemForm.description,
        status: itemForm.status,
      })
      ElMessage.success('更新成功')
    } else {
      await createDictItem(itemForm.dictId, {
        itemKey: itemForm.itemKey,
        itemValue: itemForm.itemValue,
        sort: itemForm.sort,
        description: itemForm.description,
        status: itemForm.status,
      })
      ElMessage.success('添加成功')
    }
    itemDialogVisible.value = false
    loadDictItems()
  } catch {
    // 验证失败
  }
}

const handleDeleteDictItem = async (id: string) => {
  try {
    await ElMessageBox.confirm('确定要删除该字典明细吗？', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    await deleteDictItem(id)
    ElMessage.success('删除成功')
    loadDictItems()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
      console.error('删除字典明细失败:', error)
    }
  }
}

onMounted(() => {
  loadDictList()
})
</script>

<style scoped>
:deep(.el-drawer__header) {
  margin-bottom: 0;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}
</style>
