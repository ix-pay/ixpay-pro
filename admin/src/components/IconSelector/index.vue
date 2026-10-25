<template>
  <div class="icon-selector">
    <el-popover placement="bottom-start" :width="520" trigger="click" :teleported="true">
      <template #reference>
        <el-input
          v-model="selectedIcon"
          placeholder="点击选择图标"
          readonly
          clearable
          @clear="handleClear"
        >
          <template #prefix>
            <el-icon v-if="selectedIcon" :size="18">
              <component :is="selectedIcon" />
            </el-icon>
          </template>
          <template #suffix>
            <el-icon>
              <ArrowDown />
            </el-icon>
          </template>
        </el-input>
      </template>

      <div class="icon-panel">
        <!-- 搜索框 -->
        <el-input
          v-model="searchText"
          placeholder="搜索图标名称"
          size="small"
          clearable
          class="mb-3"
        >
          <template #prefix>
            <el-icon>
              <Search />
            </el-icon>
          </template>
        </el-input>

        <!-- 图标网格 -->
        <div class="icon-grid">
          <div
            v-for="icon in paginatedIcons"
            :key="icon"
            class="icon-item"
            :class="{ active: selectedIcon === icon }"
            @click="handleSelect(icon)"
          >
            <el-icon :size="20">
              <component :is="icon" />
            </el-icon>
            <span class="icon-name">{{ icon }}</span>
          </div>
        </div>

        <!-- 分页 -->
        <div class="icon-pagination">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :page-sizes="[20, 40, 60, 80]"
            :total="filteredIcons.length"
            layout="total, sizes, prev, pager, next"
            size="small"
          />
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable @typescript-eslint/no-unused-vars */
import { ref, computed } from 'vue'

defineOptions({
  name: 'IconSelector',
})
import {
  ArrowDown,
  Search,
  // 方向类
  ArrowUp,
  ArrowLeft,
  ArrowRight,
  Top,
  Bottom,
  Right,
  Back,
  Share,
  // 操作类
  Plus,
  Minus,
  Close,
  Check,
  Edit,
  Delete,
  Refresh,
  Setting,
  More,
  MoreFilled,
  Star,
  StarFilled,
  Upload,
  UploadFilled,
  Download,
  // 菜单类
  Menu,
  Operation,
  Folder,
  FolderAdd,
  FolderOpened,
  FolderChecked,
  Document,
  Files,
  Notebook,
  Tickets,
  // 用户类
  User,
  UserFilled,
  Avatar,
  // 系统类
  Tools,
  Monitor,
  DataBoard,
  DataAnalysis,
  DataLine,
  PieChart,
  List,
  Grid,
  Connection,
  Link,
  Unlock,
  Lock,
  Key,
  // 通讯类
  ChatDotRound,
  ChatLineRound,
  ChatLineSquare,
  ChatDotSquare,
  View,
  Hide,
  Picture,
  PictureFilled,
  // 状态类
  Warning,
  WarningFilled,
  CircleCheck,
  CircleCheckFilled,
  CircleClose,
  CircleCloseFilled,
  CirclePlus,
  CirclePlusFilled,
  SuccessFilled,
  InfoFilled,
  QuestionFilled,
  // 时间类
  Clock,
  Timer,
  Calendar,
  // 其他常用
  HomeFilled,
  Location,
  LocationFilled,
  LocationInformation,
  Rank,
  Medal,
  Trophy,
  ShoppingBag,
  ShoppingCart,
  ShoppingCartFull,
  CreditCard,
  Money,
  Coin,
  Wallet,
  WalletFilled,
  Phone,
  PhoneFilled,
  VideoCamera,
  VideoCameraFilled,
  Microphone,
  Camera,
  CameraFilled,
  Bell,
  Printer,
  Filter,
  Sort,
  SortDown,
  SortUp,
  Loading,
  Aim,
  AlarmClock,
  Apple,
  Bicycle,
  Basketball,
  Box,
  Briefcase,
  Brush,
  Burger,
  Cherry,
  Chicken,
  ChromeFilled,
  Coffee,
  ColdDrink,
  Collection,
  CollectionTag,
  Comment,
  Compass,
  CopyDocument,
  Crop,
  Dessert,
  Dish,
  DishDot,
  Discount,
  Drizzling,
  EditPen,
  ElementPlus,
  Expand,
  Failed,
  Female,
  Film,
  Finished,
  FirstAidKit,
  Flag,
  Fold,
  Food,
  Football,
  ForkSpoon,
  Fries,
  FullScreen,
  Grape,
  Headset,
  Help,
  HelpFilled,
  Histogram,
  HotWater,
  House,
  IceCream,
  IceCreamRound,
  IceDrink,
  IceTea,
  Iphone,
  KnifeFork,
  Lightning,
  Lollipop,
  MagicStick,
  Magnet,
  Male,
  Management,
  MapLocation,
  MessageBox,
  Mic,
  MilkTea,
  Moon,
  MoonNight,
  Mouse,
  Mug,
  Mute,
  MuteNotification,
  NoSmoking,
  Notification,
  Odometer,
  OfficeBuilding,
  Open,
  Opportunity,
  Orange,
  Paperclip,
  PartlyCloudy,
  Pear,
  PictureRounded,
  Place,
  Platform,
  Pointer,
  Present,
  PriceTag,
  Promotion,
  QuartzWatch,
  Reading,
  ReadingLamp,
  RefreshLeft,
  RefreshRight,
  Refrigerator,
  Remove,
  ScaleToOriginal,
  School,
  Scissor,
  Select,
  Sell,
  SemiSelect,
  Service,
  Ship,
  Shop,
  Soccer,
  SoldOut,
  Stopwatch,
  Sugar,
  Suitcase,
  SuitcaseLine,
  Sunny,
  Sunrise,
  Sunset,
  Switch,
  SwitchButton,
  TakeawayBox,
  Ticket,
  ToiletPaper,
  TrendCharts,
  TurnOff,
  Umbrella,
  Van,
  VideoPause,
  VideoPlay,
  Watch,
  Watermelon,
  WindPower,
  ZoomIn,
  ZoomOut,
} from '@element-plus/icons-vue'
/* eslint-enable @typescript-eslint/no-unused-vars */

// 定义 props
interface Props {
  modelValue?: string
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
})

// 定义 emits
interface Emits {
  (e: 'update:modelValue', value: string): void
}

const emit = defineEmits<Emits>()

// 选中的图标
const selectedIcon = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

// 搜索文本
const searchText = ref('')

// 当前页码
const currentPage = ref(1)

// 每页显示数量
const pageSize = ref(40)

// 所有可用图标列表
const allIcons = [
  // 方向类
  'ArrowUp',
  'ArrowDown',
  'ArrowLeft',
  'ArrowRight',
  'Top',
  'Bottom',
  'Right',
  'Back',
  'Share',
  // 操作类
  'Plus',
  'Minus',
  'Close',
  'Check',
  'Edit',
  'Delete',
  'Refresh',
  'Setting',
  'More',
  'MoreFilled',
  'Star',
  'StarFilled',
  'Upload',
  'UploadFilled',
  'Download',
  // 菜单类
  'Menu',
  'Operation',
  'Folder',
  'FolderAdd',
  'FolderOpened',
  'FolderChecked',
  'Document',
  'Files',
  'Notebook',
  'Tickets',
  // 用户类
  'User',
  'UserFilled',
  'Avatar',
  // 系统类
  'Tools',
  'Monitor',
  'DataBoard',
  'DataAnalysis',
  'DataLine',
  'PieChart',
  'List',
  'Grid',
  'Connection',
  'Link',
  'Unlock',
  'Lock',
  'Key',
  // 通讯类
  'ChatDotRound',
  'ChatLineRound',
  'ChatLineSquare',
  'ChatDotSquare',
  'View',
  'Hide',
  'Picture',
  'PictureFilled',
  // 状态类
  'Warning',
  'WarningFilled',
  'CircleCheck',
  'CircleCheckFilled',
  'CircleClose',
  'CircleCloseFilled',
  'CirclePlus',
  'CirclePlusFilled',
  'SuccessFilled',
  'InfoFilled',
  'QuestionFilled',
  // 时间类
  'Clock',
  'Timer',
  'Calendar',
  // 其他常用
  'HomeFilled',
  'Location',
  'LocationFilled',
  'LocationInformation',
  'Rank',
  'Medal',
  'Trophy',
  'ShoppingBag',
  'ShoppingCart',
  'ShoppingCartFull',
  'CreditCard',
  'Money',
  'Coin',
  'Wallet',
  'WalletFilled',
  'Phone',
  'PhoneFilled',
  'VideoCamera',
  'VideoCameraFilled',
  'Microphone',
  'Camera',
  'CameraFilled',
  'Bell',
  'Printer',
  'Filter',
  'Sort',
  'SortDown',
  'SortUp',
  'Loading',
  'Aim',
  'AlarmClock',
  'Apple',
  'Bicycle',
  'Basketball',
  'Box',
  'Briefcase',
  'Brush',
  'Burger',
  'Cherry',
  'Chicken',
  'ChromeFilled',
  'Coffee',
  'ColdDrink',
  'Collection',
  'CollectionTag',
  'Comment',
  'Compass',
  'CopyDocument',
  'Crop',
  'Dessert',
  'Dish',
  'DishDot',
  'Discount',
  'Drizzling',
  'EditPen',
  'ElementPlus',
  'Expand',
  'Failed',
  'Female',
  'Film',
  'Finished',
  'FirstAidKit',
  'Flag',
  'Fold',
  'Food',
  'Football',
  'ForkSpoon',
  'Fries',
  'FullScreen',
  'Grape',
  'Headset',
  'Help',
  'HelpFilled',
  'Histogram',
  'HotWater',
  'House',
  'IceCream',
  'IceCreamRound',
  'IceDrink',
  'IceTea',
  'Iphone',
  'KnifeFork',
  'Lightning',
  'Lollipop',
  'MagicStick',
  'Magnet',
  'Male',
  'Management',
  'MapLocation',
  'MessageBox',
  'Mic',
  'MilkTea',
  'Moon',
  'MoonNight',
  'Mouse',
  'Mug',
  'Mute',
  'MuteNotification',
  'NoSmoking',
  'Notification',
  'Odometer',
  'OfficeBuilding',
  'Open',
  'Opportunity',
  'Orange',
  'Paperclip',
  'PartlyCloudy',
  'Pear',
  'PictureRounded',
  'Place',
  'Platform',
  'Pointer',
  'Present',
  'PriceTag',
  'Promotion',
  'QuartzWatch',
  'Reading',
  'ReadingLamp',
  'RefreshLeft',
  'RefreshRight',
  'Refrigerator',
  'Remove',
  'ScaleToOriginal',
  'School',
  'Scissor',
  'Select',
  'Sell',
  'SemiSelect',
  'Service',
  'Ship',
  'Shop',
  'Soccer',
  'SoldOut',
  'Stopwatch',
  'Sugar',
  'Suitcase',
  'SuitcaseLine',
  'Sunny',
  'Sunrise',
  'Sunset',
  'Switch',
  'SwitchButton',
  'TakeawayBox',
  'Ticket',
  'ToiletPaper',
  'TrendCharts',
  'TurnOff',
  'Umbrella',
  'Van',
  'VideoPause',
  'VideoPlay',
  'Watch',
  'Watermelon',
  'WindPower',
  'ZoomIn',
  'ZoomOut',
]

// 过滤后的图标列表
const filteredIcons = computed(() => {
  let icons = allIcons
  if (searchText.value) {
    icons = icons.filter((icon) => icon.toLowerCase().includes(searchText.value.toLowerCase()))
  }
  return icons
})

// 分页后的图标列表
const paginatedIcons = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredIcons.value.slice(start, end)
})

// 选择图标
const handleSelect = (icon: string) => {
  selectedIcon.value = icon
}

// 清空图标
const handleClear = () => {
  selectedIcon.value = ''
}
</script>

<style scoped>
.icon-selector {
  width: 100%;
}

.icon-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.icon-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  max-height: 180px; /* 两行图标的高度：2 行 * (图标 + 文字 + 间距) ≈ 180px */
  overflow-y: auto;
  padding: 4px;
}

.icon-grid::-webkit-scrollbar {
  width: 6px;
}

.icon-grid::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.icon-grid::-webkit-scrollbar-thumb:hover {
  background-color: rgba(0, 0, 0, 0.3);
}

.icon-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 8px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  gap: 4px;
}

.icon-item:hover {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.icon-item.active {
  border-color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.icon-name {
  font-size: 10px;
  color: var(--el-text-color-secondary);
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  width: 100%;
}

.icon-item.active .icon-name {
  color: var(--el-color-primary);
}

.icon-pagination {
  display: flex;
  justify-content: center;
  padding-top: 8px;
  border-top: 1px solid var(--el-border-color-light);
}
</style>
