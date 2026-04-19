<template>
  <!-- 菜单管理页面 - 专业的树形表格设计 -->
  <div class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md">
    <!-- 顶部操作栏 -->
    <div
      class="flex flex-wrap items-center justify-between gap-3 p-4 border-b border-gray-200 dark:border-gray-700"
    >
      <div class="flex flex-wrap items-center gap-3">
        <el-input
          v-model="searchForm.keyword"
          placeholder="搜索菜单名称、路径或权限标识"
          clearable
          size="default"
          class="w-64"
          @keyup.enter="loadMenuList"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>
        <el-button type="primary" @click="loadMenuList">
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
        <el-button type="primary" v-auth-btn="'system:menu:add'" @click="handleAddDirectory">
          <el-icon>
            <FolderAdd />
          </el-icon>
          新增目录
        </el-button>
        <el-button
          type="success"
          v-auth-btn="'system:menu:add'"
          @click="(e) => handleAddMenu(e as MouseEvent)"
        >
          <el-icon>
            <Menu />
          </el-icon>
          新增菜单
        </el-button>
        <el-button
          type="warning"
          v-auth-btn="'system:menu:add'"
          @click="(e) => handleAddButton(e as MouseEvent)"
        >
          <el-icon>
            <Operation />
          </el-icon>
          新增按钮
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
            <Folder />
          </el-icon>
          目录：<span class="font-medium">{{ statistics.directoryCount }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-green-500">
            <Menu />
          </el-icon>
          菜单：<span class="font-medium">{{ statistics.menuCount }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon class="text-orange-500">
            <Operation />
          </el-icon>
          按钮：<span class="font-medium">{{ statistics.buttonCount }}</span>
        </span>
        <span class="flex items-center gap-1">
          <el-icon>
            <DataLine />
          </el-icon>
          总计：<span class="font-medium">{{ statistics.totalCount }}</span>
        </span>
      </div>
    </div>

    <!-- 表格区域 -->
    <div class="flex-1 overflow-hidden">
      <el-table
        v-loading="loading"
        :data="menuList"
        height="100%"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
        :default-expand-all="expandAll"
        :indent="48"
      >
        <!-- 菜单名称列 - 带图标 -->
        <el-table-column prop="title" label="菜单名称" width="220" fixed="left">
          <template #default="scope">
            <div style="display: inline-flex; align-items: center; gap: 8px; flex-wrap: nowrap">
              <!-- 图标 -->
              <el-icon v-if="scope.row.icon" :size="16" style="flex-shrink: 0">
                <component :is="scope.row.icon" />
              </el-icon>
              <span
                :class="{ 'font-medium': scope.row.type === 1 }"
                style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis"
              >
                {{ scope.row.title }}
              </span>
            </div>
          </template>
        </el-table-column>

        <!-- 权限标识 -->
        <el-table-column prop="permission" label="权限标识" min-width="180">
          <template #default="scope">
            <span
              v-if="scope.row.permission"
              style="color: var(--text-secondary); font-size: 13px; font-family: monospace"
            >
              {{ scope.row.permission }}
            </span>
            <span v-else style="color: var(--text-placeholder)">-</span>
          </template>
        </el-table-column>

        <!-- 文件路径 -->
        <el-table-column prop="component" label="文件路径" min-width="180">
          <template #default="scope">
            <span v-if="scope.row.component" style="color: var(--text-secondary); font-size: 13px">
              {{ scope.row.component }}
            </span>
            <span v-else style="color: var(--text-placeholder)">-</span>
          </template>
        </el-table-column>

        <!-- 排序 -->
        <el-table-column prop="sort" label="排序" width="70" align="center">
          <template #default="scope">
            <span>{{ scope.row.sort }}</span>
          </template>
        </el-table-column>

        <!-- 状态 -->
        <el-table-column prop="status" label="状态" width="70" align="center">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              size="small"
              @change="handleStatusChange(scope.row)"
            />
          </template>
        </el-table-column>

        <!-- 创建时间 -->
        <el-table-column prop="createdAt" label="创建时间" width="150">
          <template #default="scope">
            <span style="color: var(--text-secondary)">
              {{ formatDate(scope.row.createdAt) }}
            </span>
          </template>
        </el-table-column>

        <!-- 操作列 -->
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <div style="display: flex; align-items: center; gap: 8px">
              <!-- 目录的操作 -->
              <template v-if="scope.row.type === 1">
                <el-button
                  v-auth-btn="'system:menu:edit'"
                  type="primary"
                  @click="handleEditMenu(scope.row)"
                >
                  编辑
                </el-button>
                <el-button
                  v-auth-btn="'system:menu:add'"
                  type="success"
                  @click="(e) => handleAddMenu(e as MouseEvent, scope.row)"
                >
                  添加菜单
                </el-button>
              </template>

              <!-- 菜单的操作 -->
              <template v-else-if="scope.row.type === 2">
                <el-button
                  v-auth-btn="'system:menu:edit'"
                  type="primary"
                  @click="handleEditMenu(scope.row)"
                >
                  编辑
                </el-button>
                <el-button
                  v-auth-btn="'system:menu:add'"
                  type="warning"
                  @click="(e) => handleAddButton(e as MouseEvent, scope.row)"
                >
                  添加按钮
                </el-button>
              </template>

              <!-- 按钮的操作 -->
              <template v-else-if="scope.row.type === 3">
                <el-button
                  v-auth-btn="'system:menu:edit'"
                  type="primary"
                  @click="handleEditMenu(scope.row)"
                >
                  编辑
                </el-button>
              </template>

              <!-- 删除按钮 -->
              <el-popconfirm title="确定要删除吗？" @confirm="handleDeleteMenu(scope.row.id)">
                <template #reference>
                  <el-button v-auth-btn="'system:menu:delete'" type="danger"> 删除 </el-button>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 菜单表单对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form ref="menuFormRef" :model="menuForm" :rules="formRules" label-width="100px">
        <!-- 菜单类型 -->
        <el-form-item label="菜单类型" prop="type">
          <el-radio-group v-model="menuForm.type" :disabled="isEdit || isQuickAdd">
            <el-radio :label="1">
              <el-icon>
                <Folder />
              </el-icon>
              目录
            </el-radio>
            <el-radio :label="2">
              <el-icon>
                <Menu />
              </el-icon>
              菜单
            </el-radio>
            <el-radio :label="3">
              <el-icon>
                <Operation />
              </el-icon>
              按钮
            </el-radio>
          </el-radio-group>
          <div v-if="isEdit || isQuickAdd" class="text-xs text-gray-500 mt-1">
            菜单类型创建后不可修改
          </div>
        </el-form-item>

        <!-- 菜单名称 -->
        <el-form-item label="菜单名称" prop="Title">
          <el-input v-model="menuForm.Title" placeholder="请输入菜单名称" />
        </el-form-item>

        <!-- 路由名称 -->
        <el-form-item label="路由名称" prop="name">
          <el-input v-model="menuForm.name" placeholder="请输入路由名称（组件名）" />
        </el-form-item>

        <!-- 路由路径 -->
        <el-form-item v-if="menuForm.type !== 3" label="路由路径" prop="path">
          <el-input v-model="menuForm.path" placeholder="请输入路由路径" />
        </el-form-item>

        <!-- 组件路径 -->
        <el-form-item v-if="menuForm.type === 2" label="组件路径" prop="component">
          <el-input
            v-model="menuForm.component"
            placeholder="请输入组件路径，如：views/system/user/index"
          />
        </el-form-item>

        <!-- 权限标识 -->
        <el-form-item v-if="menuForm.type === 3" label="权限标识" prop="permission">
          <el-input
            v-model="menuForm.permission"
            placeholder="请输入权限标识，如：system:user:add"
          />
          <div class="text-xs text-gray-500 mt-1">
            格式：模块：功能：操作（如：system:user:add）
          </div>
        </el-form-item>

        <!-- 图标 -->
        <el-form-item label="图标" prop="Icon">
          <IconSelector v-model="menuForm.Icon" />
          <div class="text-xs text-gray-500 mt-1">点击输入框选择 Element Plus 图标</div>
        </el-form-item>

        <!-- 父菜单 -->
        <el-form-item label="父菜单" prop="parentId">
          <el-tree-select
            v-model="menuForm.parentId"
            :data="menuList"
            :props="{ label: 'title', value: 'id', children: 'children' }"
            placeholder="请选择父菜单（顶级菜单不选）"
            clearable
            check-strictly
            value-key="id"
            :render-after-expand="false"
            class="w-full"
          />
          <div v-if="menuForm.parentId === ''" class="text-xs text-gray-500 mt-1">
            不选择则为顶级菜单
          </div>
        </el-form-item>

        <!-- 关联 API -->
        <el-form-item
          v-if="menuForm.type === 2 || menuForm.type === 3"
          label="关联 API"
          prop="apiIds"
        >
          <el-tree-select
            v-model="menuForm.apiIds"
            :data="apiTreeData"
            :props="{ label: 'label', value: 'value', children: 'children' }"
            placeholder="请输入关键词搜索 API"
            multiple
            filterable
            remote
            :remote-method="searchApi"
            check-strictly
            clearable
            class="w-full"
            :reserve-keyword="false"
          />
          <div class="text-xs text-gray-500 mt-1">选择该菜单/按钮关联的 API 接口权限标识</div>
        </el-form-item>

        <!-- 排序、状态、缓存（一行显示） -->
        <div class="flex gap-4">
          <el-form-item label="排序" prop="sort" class="flex-1 min-w-[150px]">
            <el-input-number
              v-model="menuForm.sort"
              :min="0"
              :max="999"
              class="w-full"
              controls-position="right"
            />
          </el-form-item>

          <el-form-item label="状态" prop="status" class="flex-1 min-w-[120px]">
            <el-switch
              v-model="menuForm.status"
              active-color="#13ce66"
              inactive-color="#ff4949"
              active-value="1"
              inactive-value="0"
            />
          </el-form-item>

          <el-form-item
            v-if="menuForm.type === 2"
            label="缓存"
            prop="KeepAlive"
            class="flex-1 min-w-[120px]"
          >
            <el-switch
              v-model="menuForm.KeepAlive"
              active-color="#13ce66"
              inactive-color="#ff4949"
            />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, type FormInstance } from 'element-plus'
import {
  Search,
  Refresh,
  FolderAdd,
  Menu,
  Operation,
  Folder,
  DataLine,
} from '@element-plus/icons-vue'
import {
  getMenuTree,
  deleteMenu,
  createMenu,
  updateMenu,
  searchApiList,
  type MenuItem,
} from '@/api/modules/menu'
import IconSelector from '@/components/IconSelector/index.vue'

defineOptions({
  name: 'MenuManagement',
})

// 菜单列表数据
const menuList = ref<MenuItem[]>([])
// 加载状态
const loading = ref(false)
// 搜索表单
const searchForm = reactive({
  keyword: '',
})
// 对话框状态
const dialogVisible = ref(false)
const dialogTitle = ref('')
// 表单引用
const menuFormRef = ref<FormInstance | null>(null)
// 当前操作类型：1-目录，2-菜单，3-按钮
const currentType = ref(2)
// 是否展开所有层级
const expandAll = ref(true)
// 是否是快速新增模式（从操作列按钮打开）
const isQuickAdd = ref(false)

// 统计数据
const statistics = computed(() => {
  let directoryCount = 0
  let menuCount = 0
  let buttonCount = 0

  const count = (menus: MenuItem[]) => {
    for (const menu of menus) {
      if (menu.type === 1) directoryCount++
      else if (menu.type === 2) menuCount++
      else if (menu.type === 3) buttonCount++
      if (menu.children && menu.children.length > 0) {
        count(menu.children)
      }
    }
  }
  count(menuList.value)

  return {
    directoryCount,
    menuCount,
    buttonCount,
    totalCount: directoryCount + menuCount + buttonCount,
  }
})

// 是否编辑模式
const isEdit = ref(false)

// 菜单表单数据
const menuForm = reactive({
  id: '',
  name: '',
  Title: '',
  path: '',
  component: '',
  Icon: '',
  permission: '',
  sort: 0,
  parentId: '',
  status: 1,
  KeepAlive: false,
  type: 2,
  apiIds: [] as string[],
  hidden: false,
  isExt: false,
  redirect: '',
  defaultMenu: false,
  breadcrumb: true,
  activeMenu: '',
  affix: false,
  frameLoading: false,
})

// API 树形数据
const apiTreeData = ref<
  Array<{
    label: string
    value: string
    children: Array<{
      label: string
      value: string
      path?: string
      method?: string
      group?: string
      description?: string
    }>
  }>
>([])

// 加载 API 列表数据
const loadApiList = async (keyword?: string) => {
  try {
    const response = await searchApiList({ keyword })
    console.log('API 列表响应:', response)

    let apiList: {
      id: string | number
      group?: string
      method?: string
      path?: string
      description?: string
    }[] = []
    const responseData = response.data as
      | { data?: typeof apiList; list?: typeof apiList }
      | typeof apiList

    // 处理不同的返回结构
    if (Array.isArray(responseData)) {
      // 直接返回数组
      apiList = responseData
    } else if (responseData?.list && Array.isArray(responseData.list)) {
      // 分页结构：{ list: [], total: 0 }
      apiList = responseData.list
    } else if (responseData?.data && Array.isArray(responseData.data)) {
      // 嵌套结构：{ data: { data: [] } }
      apiList = responseData.data
    } else {
      console.error('未知的 API 响应格式:', response)
      ElMessage.warning('API 数据格式异常')
      return
    }

    console.log('API 列表数据:', apiList)

    // 按分组转换为树形结构
    interface ApiGroup {
      label: string
      value: string
      children: Array<{
        label: string
        value: string
        path?: string
        method?: string
        group?: string
        description?: string
      }>
    }

    const groupMap = new Map<string, ApiGroup>()
    apiList.forEach((api) => {
      const group = api.group || '未分组'
      if (!groupMap.has(group)) {
        groupMap.set(group, {
          label: group,
          value: group,
          children: [],
        })
      }
      groupMap.get(group)!.children.push({
        label: `${api.method || ''} ${api.path || ''} - ${api.description || '无描述'}`,
        value: String(api.id),
        path: api.path,
        method: api.method,
        group: api.group,
        description: api.description,
      })
    })

    apiTreeData.value = Array.from(groupMap.values())
  } catch (error) {
    console.error('获取 API 列表失败:', error)
    ElMessage.error('获取 API 列表失败')
  }
}

// 搜索 API
const searchApi = async (keyword: string) => {
  if (!keyword || keyword.length < 1) {
    await loadApiList()
    return
  }
  await loadApiList(keyword)
}

// 表单验证规则
const formRules = reactive({
  Title: [
    { required: true, message: '请输入菜单名称', trigger: 'blur' },
    { min: 1, max: 50, message: '菜单名称长度在 1 到 50 个字符', trigger: 'blur' },
  ],
  name: [
    { required: true, message: '请输入路由名称', trigger: 'blur' },
    { min: 1, max: 100, message: '路由名称长度在 1 到 100 个字符', trigger: 'blur' },
  ],
  path: [{ required: true, message: '请输入路由路径', trigger: 'blur' }],
  component: [{ required: true, message: '请输入组件路径', trigger: 'blur' }],
  permission: [{ required: true, message: '请输入权限标识', trigger: 'blur' }],
})

// 获取菜单列表（树形结构）
const loadMenuList = async () => {
  loading.value = true
  try {
    const response = await getMenuTree()
    const allMenus = (response.data as MenuItem[]) || []

    // 如果有搜索条件，进行过滤
    if (searchForm.keyword) {
      menuList.value = filterMenus(allMenus, searchForm.keyword)
    } else {
      menuList.value = allMenus
    }
  } catch (error) {
    ElMessage.error('获取菜单列表失败')
    console.error('获取菜单列表失败:', error)
  } finally {
    loading.value = false
  }
}

// 递归过滤菜单树
const filterMenus = (menus: MenuItem[], keyword: string): MenuItem[] => {
  const result: MenuItem[] = []
  for (const menu of menus) {
    // 如果当前菜单匹配
    if (
      menu.title?.toLowerCase().includes(keyword.toLowerCase()) ||
      menu.path?.toLowerCase().includes(keyword.toLowerCase()) ||
      menu.permission?.toLowerCase().includes(keyword.toLowerCase())
    ) {
      // 复制菜单并保留子菜单
      const clonedMenu: MenuItem = { ...menu }
      if (menu.children && menu.children.length > 0) {
        clonedMenu.children = filterMenus(menu.children, keyword)
      }
      result.push(clonedMenu)
    } else if (menu.children && menu.children.length > 0) {
      // 递归过滤子菜单
      const filteredChildren = filterMenus(menu.children, keyword)
      if (filteredChildren.length > 0) {
        // 如果子菜单有匹配的，保留父菜单
        result.push({ ...menu, children: filteredChildren })
      }
    }
  }
  return result
}

// 重置搜索
const resetSearch = () => {
  searchForm.keyword = ''
  loadMenuList()
}

// 格式化日期
const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 添加目录
const handleAddDirectory = () => {
  currentType.value = 1
  dialogTitle.value = '新增目录'
  resetForm()
  menuForm.type = 1
  isEdit.value = false
  isQuickAdd.value = false // 顶部按钮，可以选择类型
  dialogVisible.value = true
}

// 添加菜单
const handleAddMenu = (event?: MouseEvent, parentMenu?: MenuItem) => {
  currentType.value = 2
  dialogTitle.value = parentMenu ? '添加菜单' : '新增菜单'
  resetForm()
  menuForm.type = 2
  isEdit.value = false
  isQuickAdd.value = !!parentMenu // 有父菜单表示从操作列打开，类型不可选
  if (parentMenu) {
    menuForm.parentId = String(parentMenu.id)
  }
  dialogVisible.value = true
}

// 添加按钮
const handleAddButton = (event?: MouseEvent, parentMenu?: MenuItem) => {
  currentType.value = 3
  dialogTitle.value = parentMenu ? '添加按钮' : '新增按钮'
  resetForm()
  menuForm.type = 3
  isEdit.value = false
  isQuickAdd.value = !!parentMenu // 有父菜单表示从操作列打开，类型不可选
  if (parentMenu) {
    menuForm.parentId = String(parentMenu.id)
  }
  dialogVisible.value = true
}

// 重置表单
const resetForm = () => {
  Object.assign(menuForm, {
    id: '',
    name: '',
    Title: '',
    path: '',
    component: '',
    Icon: '',
    permission: '',
    sort: 0,
    parentId: '',
    status: 1,
    KeepAlive: false,
    type: 2,
    apiIds: [],
  })
  isEdit.value = false
  isQuickAdd.value = false
}

// 编辑菜单
const handleEditMenu = (menu: MenuItem) => {
  dialogTitle.value = '编辑菜单'
  isEdit.value = true
  Object.assign(menuForm, {
    id: String(menu.id),
    name: menu.name,
    Title: menu.title, // 后端返回小写 title
    path: menu.path,
    component: menu.component,
    Icon: menu.icon, // 后端返回驼峰命名图标（如 UserFilled）
    permission: menu.permission,
    sort: menu.sort,
    parentId: menu.parentId === '0' ? '' : String(menu.parentId || ''),
    status: String(menu.status), // 转为字符串以匹配 el-switch 的 active-value
    KeepAlive: menu.keepAlive, // 后端返回小写 keepAlive
    type: menu.type,
    apiIds: menu.apiIds?.map((id) => String(id)) || [],
  })
  dialogVisible.value = true
}

// 状态变更
const handleStatusChange = async (menu: MenuItem) => {
  try {
    await updateMenu(menu)
    ElMessage.success('状态更新成功')
  } catch (error) {
    ElMessage.error('状态更新失败')
    console.error('状态更新失败:', error)
  }
}

// 提交表单
const handleSubmit = async () => {
  try {
    await menuFormRef.value?.validate()
    // 构建提交数据，使用小写字段名匹配后端的 json 标签
    const submitData = {
      id: menuForm.id,
      name: menuForm.name,
      title: menuForm.Title, // 注意：后端 json 标签是 title
      path: menuForm.path,
      component: menuForm.component,
      icon: menuForm.Icon, // 直接使用驼峰命名（与 Element Plus 组件名一致）
      permission: menuForm.permission,
      sort: menuForm.sort,
      parentId: menuForm.parentId || '0',
      status: menuForm.status,
      type: menuForm.type,
      keepAlive: menuForm.KeepAlive,
      hidden: menuForm.hidden || false,
      isExt: menuForm.isExt || false,
      redirect: menuForm.redirect || '',
      defaultMenu: menuForm.defaultMenu || false,
      breadcrumb: menuForm.breadcrumb !== undefined ? menuForm.breadcrumb : true,
      activeMenu: menuForm.activeMenu || '',
      affix: menuForm.affix || false,
      frameLoading: menuForm.frameLoading || false,
      apiIds: menuForm.apiIds || [],
    }

    if (submitData.id) {
      // 编辑菜单
      await updateMenu(submitData)
      ElMessage.success('更新成功')
    } else {
      // 添加菜单
      await createMenu(submitData)
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    loadMenuList()
  } catch (error) {
    console.error('提交表单失败:', error)
  }
}

// 删除菜单
const handleDeleteMenu = async (id: string) => {
  try {
    await deleteMenu(id)
    ElMessage.success('删除成功')
    loadMenuList()
  } catch (error) {
    ElMessage.error('删除失败')
    console.error('删除菜单失败:', error)
  }
}

onMounted(() => {
  loadMenuList()
  loadApiList() // 加载 API 列表数据
})
</script>

<style scoped>
/* 
 * 菜单管理页面样式说明：
 * - 布局使用 Tailwind CSS
 * - Element Plus 组件使用原生样式 + 必要的 Tailwind 辅助类
 */

/* 表格容器高度 */
.flex-1 {
  min-height: 0;
  /* 允许 flex 子项滚动 */
}

/* 表格样式修正 */
:deep(.el-table) {
  font-size: 14px;
}

/* 固定列背景色 - 使用项目主题变量 */
:deep(.el-table__header th) {
  background-color: var(--bg-dark) !important;
  color: var(--text-primary) !important;
  font-weight: 600 !important;
}

/* 树形表格展开图标垂直居中 - 关键修复 */
:deep(.el-table__expand-icon) {
  display: inline-flex !important;
  align-items: center !important;
  justify-content: center !important;
  height: 100% !important;
  vertical-align: middle !important;
  margin-right: 4px !important;
}

/* 表格单元格内容 */
:deep(.el-table .cell) {
  white-space: normal;
  word-wrap: break-word;
}

/* 菜单名称列单元格 */
:deep(.el-table .el-table-column__Title .cell) {
  display: inline-flex !important;
  align-items: center !important;
  gap: 8px !important;
  flex-wrap: nowrap !important;
}

/* 修复固定列的宽度问题 */
:deep(.el-table__fixed) {
  width: auto !important;
}

/* 修复表格滚动条 */
:deep(.el-table__body-wrapper) {
  overflow: auto !important;
}
</style>
