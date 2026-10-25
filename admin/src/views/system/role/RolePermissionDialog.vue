<template>
  <el-drawer v-model="dialogVisible" title="角色权限设置" direction="rtl" size="900px" :close-on-click-modal="false"
    @close="handleClose">
    <div class="drawer-wrapper">
      <el-tabs v-model="activeTab" class="permission-tabs">
        <!-- 菜单权限标签页 -->
        <el-tab-pane label="菜单权限" name="menu">
          <div class="tab-scroll-area">
            <el-tree ref="menuTreeRef" :data="menuTree" :props="menuTreeProps" show-checkbox node-key="id"
              :default-checked-keys="checkedMenuIds" @check="handleMenuCheck" />
          </div>
        </el-tab-pane>

        <!-- API 权限标签页 -->
        <el-tab-pane label="API 权限" name="api">
          <el-card class="api-card">
            <template #header>
              <div class="card-header">
                <span>API 路由列表（仅通用 API）</span>
                <el-button size="small" @click="toggleAllApis">
                  {{ allApisChecked ? '取消全选' : '全选' }}
                </el-button>
              </div>
            </template>
            <div class="tab-scroll-area">
              <div v-for="group in groupedApis" :key="group.name" class="api-group">
                <div class="group-title">
                  <el-checkbox :indeterminate="isGroupIndeterminate(group)" :model-value="isGroupChecked(group)"
                    :disabled="group.allDisabled" @change="handleGroupCheck(group, $event)">
                    {{ group.name }} ({{ group.apis.length }})
                  </el-checkbox>
                </div>
                <el-checkbox-group v-model="checkedApiIds" class="group-apis">
                  <el-checkbox v-for="api in group.apis" :key="`${api.id}-${api._index}`" :value="api.id"
                    :disabled="api.disabled">
                    <span class="method-tag" :class="`method-${api.method.toLowerCase()}`">{{
                      api.method
                      }}</span>
                    <span class="api-path">{{ api.path }}</span>
                    <span v-if="api.description" class="api-desc">- {{ api.description }}</span>
                  </el-checkbox>
                </el-checkbox-group>
              </div>
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </span>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { TreeInstance, TreeNodeData, CheckboxValueType } from 'element-plus'
import { getMenuTree } from '@/api/modules/menu'
import {
  getRoleAvailableApis,
  getRolePermissionDetail,
  saveRolePermissions,
} from '@/api/modules/role'
import type { Role as RoleType } from '@/types/role'
import type { ApiRoute as ApiRouteType } from '@/api/modules/api-route'

interface Props {
  visible: boolean
  roleId: string
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
})

const activeTab = ref('menu')
const saving = ref(false)
const loading = ref(false)

const menuTreeRef = ref<TreeInstance>()
const menuTree = ref<MenuItem[]>([])
const allMenus = ref<MenuItem[]>([])
const checkedMenuIds = ref<(string | number)[]>([]) // 仅叶子节点
const halfCheckedMenuIds = ref<(string | number)[]>([]) // 半选的父节点
const allApis = ref<ApiRouteWithDisabled[]>([])
const checkedApiIds = ref<string[]>([])

// API 路由类型（带 disabled 和调试字段）
interface ApiRouteWithDisabled extends ApiRouteType {
  disabled?: boolean
  _index?: number // 用于调试和唯一标识
}

// 菜单树节点类型（与后端 getMenuTree 返回结构对齐）
interface MenuItem {
  id: string | number
  parentId?: string | number
  title?: string
  name?: string
  path?: string
  component?: string
  icon?: string
  type?: number // 1-目录，2-菜单，3-按钮
  children?: MenuItem[]
}

const menuTreeProps = {
  children: 'children',
  label: (data: TreeNodeData) => {
    const menuItem = data as MenuItem
    const typeLabel = menuItem.type === 3 ? '[按钮]' : menuItem.type === 1 ? '[目录]' : '[菜单]'
    return `${typeLabel} ${menuItem.title || menuItem.name}`
  },
}

// 按组分组 API
const groupedApis = computed(() => {
  const groups: Record<string, ApiRouteWithDisabled[]> = {}

  if (!Array.isArray(allApis.value)) {
    console.warn('allApis.value 不是数组:', allApis.value)
    return []
  }

  allApis.value.forEach((api) => {
    const groupName = api.group || '未分组'
    if (!groups[groupName]) {
      groups[groupName] = []
    }
    groups[groupName].push(api)
  })

  return Object.entries(groups).map(([name, apis]) => {
    const allDisabled = apis.every((api) => api.disabled)
    return { name, apis, allDisabled }
  })
})

const allApisChecked = computed(() => {
  const availableApis = allApis.value.filter((api) => !api.disabled)
  return checkedApiIds.value.length === availableApis.length
})

// 加载数据
const loadData = async () => {
  if (!props.roleId) return
  loading.value = true
  try {
    // 并行加载菜单树、API 列表和角色详情
    // 注意：roleId 保持字符串类型，避免 Number() 转换导致精度丢失
    const [menuRes, apiRes, roleRes] = await Promise.all([
      getMenuTree(),
      getRoleAvailableApis(props.roleId),
      getRolePermissionDetail(props.roleId),
    ])

    // 处理菜单数据
    allMenus.value = (menuRes.data as MenuItem[]) || []
    menuTree.value = buildMenuTree(allMenus.value)

    // 设置已勾选的菜单 ID（仅叶子节点，父节点状态由 el-tree 自动计算半选）
    const role = roleRes.data as unknown as RoleType
    const leafIds = getLeafNodeIds(menuTree.value)
    const allRoleMenuIds = (role.menus || []).map((m) => m.id)
    checkedMenuIds.value = allRoleMenuIds.filter((id) => leafIds.has(id))
    // 非叶子节点的菜单 ID 作为半选父节点初始值
    halfCheckedMenuIds.value = allRoleMenuIds.filter((id) => !leafIds.has(id))

    // 处理 API 数据（带 disabled 标记）
    // 后端返回 PascalCase 字段（ID, Path, Method, Group 等），需转换为前端 camelCase
    let rawApiData: unknown[] = []
    if (Array.isArray(apiRes.data)) {
      rawApiData = apiRes.data
    } else if (apiRes.data && typeof apiRes.data === 'object' && 'list' in apiRes.data) {
      rawApiData = (apiRes.data as { list?: unknown[] }).list || []
    }
    const apiData: ApiRouteWithDisabled[] = rawApiData
      .map((item: unknown, index: number) => {
        const raw = item as Record<string, unknown>
        // 优先使用 PascalCase 字段，兼容 camelCase
        const id = String(raw.ID || raw.id || '')
        return {
          id,
          path: (raw.Path || raw.path || '') as string,
          method: (raw.Method || raw.method || '') as string,
          name: (raw.Name || raw.name || '') as string,
          group: (raw.Group || raw.group || '') as string,
          description: (raw.Description || raw.description || '') as string,
          authRequired: (raw.AuthRequired ?? raw.authRequired ?? false) as boolean,
          status: (raw.Status ?? raw.status ?? 1) as number,
          createdAt: (raw.CreatedAt || raw.createdAt || '') as string,
          updatedAt: (raw.UpdatedAt || raw.updatedAt || '') as string,
          disabled: false,
          _index: index, // 用于调试和唯一标识
        }
      })
      // 过滤掉没有有效 id 的数据
      .filter((api) => api.id && api.id !== '')
    console.log('API 数据映射完成:', apiData.length, '条，分组:', [...new Set(apiData.map((a) => a.group))])
    allApis.value = apiData

    // 设置已勾选的 API ID
    checkedApiIds.value = (role.routes || []).map((api) => String(api.id))
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 获取树中所有叶子节点 ID（无 children 的节点）
const getLeafNodeIds = (nodes: MenuItem[]): Set<string | number> => {
  const leafIds = new Set<string | number>()
  const traverse = (items: MenuItem[]) => {
    items.forEach((item) => {
      if (item.children && item.children.length > 0) {
        traverse(item.children)
      } else {
        leafIds.add(item.id)
      }
    })
  }
  traverse(nodes)
  return leafIds
}

// 构建菜单树（后端数据已包含嵌套 children 和按钮节点（type=3））
const buildMenuTree = (menus: MenuItem[]): MenuItem[] => {
  // 后端返回的数据已包含完整的树结构和按钮节点
  // 直接返回即可，el-tree 会自动处理嵌套 children
  return menus
}

// 处理菜单树勾选（叶子节点存 checkedMenuIds，半选父节点存 halfCheckedMenuIds）
// 按钮（type=3）已作为菜单项存储在 base_menus 表中，统一通过 menuIds 授权
const handleMenuCheck = () => {
  const checkedKeys = menuTreeRef.value?.getCheckedKeys(false) as (string | number)[]
  const halfCheckedKeys = menuTreeRef.value?.getHalfCheckedKeys() as (string | number)[]

  // 所有叶子节点（包括按钮）都作为菜单 ID 保存
  checkedMenuIds.value = checkedKeys
  halfCheckedMenuIds.value = halfCheckedKeys
}

// 检查组是否全选
const isGroupChecked = (group: {
  name: string
  apis: ApiRouteWithDisabled[]
  allDisabled: boolean
}) => {
  const availableApis = group.apis.filter((api) => !api.disabled)
  if (availableApis.length === 0) return false
  return availableApis.every((api) => checkedApiIds.value.includes(api.id))
}

// 检查组是否半选
const isGroupIndeterminate = (group: {
  name: string
  apis: ApiRouteWithDisabled[]
  allDisabled: boolean
}) => {
  const availableApis = group.apis.filter((api) => !api.disabled)
  if (availableApis.length === 0) return false
  const count = availableApis.filter((api) => checkedApiIds.value.includes(api.id)).length
  return count > 0 && count < availableApis.length
}

// 处理组勾选
const handleGroupCheck = (
  group: { name: string; apis: ApiRouteWithDisabled[]; allDisabled: boolean },
  checked: CheckboxValueType,
) => {
  const availableApiIds = group.apis.filter((api) => !api.disabled).map((api) => api.id)

  if (checked) {
    checkedApiIds.value = Array.from(new Set([...checkedApiIds.value, ...availableApiIds]))
  } else {
    checkedApiIds.value = checkedApiIds.value.filter((id) => !availableApiIds.includes(id))
  }
}

// 切换全选状态
const toggleAllApis = () => {
  const availableApis = allApis.value.filter((api) => !api.disabled)
  if (allApisChecked.value) {
    // 取消全选
    const availableIds = availableApis.map((api) => api.id)
    checkedApiIds.value = checkedApiIds.value.filter((id) => !availableIds.includes(id))
  } else {
    // 全选
    checkedApiIds.value = Array.from(
      new Set([...checkedApiIds.value, ...availableApis.map((api) => api.id)]),
    )
  }
}

// 保存权限
const handleSave = async () => {
  if (!props.roleId) return
  saving.value = true
  try {
    // 合并叶子节点和半选父节点，作为完整的菜单权限保存
    // 按钮（type=3）已作为菜单项存储，通过 menuIds 统一授权
    const allMenuIds = [...checkedMenuIds.value, ...halfCheckedMenuIds.value]
    await saveRolePermissions(props.roleId, {
      menuIds: allMenuIds.map((id) => String(id)),
      apiRouteIds: checkedApiIds.value,
    })

    ElMessage.success('保存成功')
    // 【新增】提示用户权限变更
    ElMessage.info('角色权限已更新，相关用户需重新登录以应用最新权限')

    emit('success')
    dialogVisible.value = false
  } catch (error) {
    ElMessage.error('保存失败')
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

const handleClose = () => {
  dialogVisible.value = false
}

// 监听 visible 变化，加载数据
watch(
  () => props.visible,
  (val) => {
    if (val) {
      loadData()
    }
  },
  { immediate: true },
)
</script>

<style scoped lang="scss">
/* 阻止 drawer body 自身的滚动条，所有滚动交给内部 .tab-scroll-area */
:deep(.el-drawer__body) {
  overflow: hidden !important;
}

/* 外层包裹器：固定高度 = 视口 - drawer 头部 - drawer 底部 - body 内边距 */
.drawer-wrapper {
  height: calc(100vh - 180px);
  overflow: hidden;
}

/* tabs 填满包裹器 */
.permission-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* tabs 内容区占满剩余高度，不产生滚动 */
:deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

/* 每个 tab-pane 填满内容区 */
:deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
}

/* 可滚动内容区域：高度 100% + 溢出滚动 */
.tab-scroll-area {
  height: 100%;
  overflow-y: auto;
  padding: 4px 0;
}

/* API 卡片填满 tab-pane，内部 flex 布局 */
.api-card {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  :deep(.el-card__body) {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.api-group {
  margin-bottom: 16px;

  .group-title {
    font-weight: bold;
    margin-bottom: 8px;
  }

  .group-apis {
    padding-left: 20px;

    .el-checkbox {
      display: block;
      margin-bottom: 4px;
    }
  }
}

.method-tag {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: bold;
  margin-right: 8px;

  &.method-get {
    background-color: #67c23a;
    color: white;
  }

  &.method-post {
    background-color: #409eff;
    color: white;
  }

  &.method-put {
    background-color: #e6a23c;
    color: white;
  }

  &.method-delete {
    background-color: #f56c6c;
    color: white;
  }
}

.api-path {
  font-family: 'Courier New', monospace;
  color: #606266;
}

.api-desc {
  color: #909399;
  font-size: 12px;
}

.mb-4 {
  margin-bottom: 16px;
}

.text-sm {
  font-size: 12px;
  color: #606266;

  li {
    margin-bottom: 4px;
  }
}
</style>
