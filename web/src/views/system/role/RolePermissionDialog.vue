<template>
  <el-dialog
    v-model="dialogVisible"
    title="角色权限设置"
    width="900px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-tabs v-model="activeTab">
      <!-- 菜单权限标签页 -->
      <el-tab-pane label="菜单权限" name="menu">
        <el-tree
          ref="menuTreeRef"
          :data="menuTree"
          :props="menuTreeProps"
          show-checkbox
          node-key="id"
          :default-checked-keys="checkedMenuIds"
          @check="handleMenuCheck"
        />
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
          <el-checkbox-group v-model="checkedApiIds">
            <div v-for="group in groupedApis" :key="group.name" class="api-group">
              <div class="group-title">
                <el-checkbox
                  :indeterminate="isGroupIndeterminate(group)"
                  :checked="isGroupChecked(group)"
                  :disabled="group.allDisabled"
                  @change="handleGroupCheck(group, $event)"
                >
                  {{ group.name }} ({{ group.apis.length }})
                </el-checkbox>
              </div>
              <div class="group-apis">
                <el-checkbox
                  v-for="api in group.apis"
                  :key="api.id"
                  :label="api.id"
                  :disabled="api.disabled"
                >
                  <span class="method-tag" :class="`method-${api.method.toLowerCase()}`">{{
                    api.method
                  }}</span>
                  <span class="api-path">{{ api.path }}</span>
                  <span v-if="api.description" class="api-desc">- {{ api.description }}</span>
                </el-checkbox>
              </div>
            </div>
          </el-checkbox-group>
        </el-card>
      </el-tab-pane>
    </el-tabs>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { TreeInstance, TreeNodeData, CheckboxValueType } from 'element-plus'
import { getMenuList } from '@/api/modules/menu'
import {
  getRoleAvailableApis,
  getRolePermissionDetail,
  saveRolePermissions,
} from '@/api/modules/role'
import type { Role as RoleType, BtnPerm } from '@/types/role'
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
const checkedMenuIds = ref<(string | number)[]>([])
const checkedBtnPermIds = ref<(string | number)[]>([])

const allApis = ref<ApiRouteWithDisabled[]>([])
const checkedApiIds = ref<string[]>([])

// API 路由类型（带 disabled 字段）
interface ApiRouteWithDisabled extends ApiRouteType {
  disabled?: boolean
}

// 菜单树节点类型
interface MenuItem {
  id: string | number
  parentId?: number
  menuName: string
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
    return `${typeLabel} ${menuItem.menuName || menuItem.name}`
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
      getMenuList(),
      getRoleAvailableApis(props.roleId),
      getRolePermissionDetail(props.roleId),
    ])

    // 处理菜单数据
    allMenus.value = (menuRes.data as MenuItem[]) || []
    menuTree.value = buildMenuTreeWithBtnPerms(allMenus.value, roleRes.data?.btnPerms || [])

    // 设置已勾选的菜单 ID（包含按钮的上级菜单）
    const role = roleRes.data as unknown as RoleType
    checkedMenuIds.value = (role.menus || []).map((m) => m.id)

    // 处理 API 数据（带 disabled 标记）
    let apiData: ApiRouteWithDisabled[] = []
    if (Array.isArray(apiRes.data)) {
      apiData = apiRes.data
    } else if (apiRes.data && typeof apiRes.data === 'object' && 'list' in apiRes.data) {
      apiData = (apiRes.data as { list?: ApiRouteWithDisabled[] }).list || []
    }
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

// 构建菜单树（包含按钮权限）
const buildMenuTreeWithBtnPerms = (menus: MenuItem[], btnPerms: BtnPerm[]): MenuItem[] => {
  const map: Record<string, MenuItem & { children: MenuItem[] }> = {}
  const roots: MenuItem[] = []

  // 创建菜单节点
  menus.forEach((menu) => {
    map[menu.id] = { ...menu, children: [] }
  })

  // 添加按钮权限作为菜单的子节点
  btnPerms.forEach((btn) => {
    const btnNode: MenuItem = {
      id: `btn-${btn.id}`,
      parentId: Number(btn.menuId),
      menuName: btn.name,
      name: btn.code,
      type: 3, // 按钮类型
      children: [],
    }

    // 将按钮添加到父菜单的 children
    if (map[btn.menuId]) {
      map[btn.menuId].children.push(btnNode)
    }
  })

  // 构建树形结构
  menus.forEach((menu) => {
    if (menu.parentId && map[menu.parentId]) {
      map[menu.parentId].children.push(map[menu.id])
    } else {
      roots.push(map[menu.id])
    }
  })

  return roots
}

// 处理菜单树勾选
const handleMenuCheck = () => {
  const checkedKeys = menuTreeRef.value?.getCheckedKeys(false) as (string | number)[]
  const halfCheckedKeys = menuTreeRef.value?.getHalfCheckedKeys() as (string | number)[]

  // 过滤出菜单 ID（排除按钮）
  const menuIds = [...checkedKeys, ...halfCheckedKeys].filter(
    (id) => typeof id === 'number' || !String(id).startsWith('btn-'),
  )

  // 过滤出按钮 ID
  const btnIds = checkedKeys.filter((id) => String(id).startsWith('btn-'))

  checkedMenuIds.value = menuIds
  checkedBtnPermIds.value = btnIds
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
    // 使用统一的保存接口
    // 注意：roleId 保持字符串类型，避免精度丢失
    await saveRolePermissions(props.roleId, {
      menuIds: checkedMenuIds.value.map((id) => String(id)),
      btnPermIds: checkedBtnPermIds.value.map((id) => String(id).replace('btn-', '')),
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
.api-card {
  max-height: 500px;
  overflow-y: auto;
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
