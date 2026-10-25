<template>
  <div class="flex flex-col h-full bg-[var(--bg-color)] rounded-lg shadow-md">
    <!-- 租户切换（单租户切换管理模式：选中后本页应用/商户数据来自对应租户库） -->
    <div class="flex flex-wrap items-center gap-3 p-4 pb-0">
      <span class="text-sm font-medium">租户</span>
      <TenantSwitcher @change="onTenantChange" />
      <span class="text-xs text-gray-500 dark:text-gray-400"
        >选中后本页应用/商户数据来自对应租户库</span
      >
    </div>

    <!-- 页签导航 -->
    <el-tabs v-model="activeTab">
      <el-tab-pane label="微信应用配置" name="apps" />
      <el-tab-pane label="商户号配置" name="merchants" />
    </el-tabs>

    <!-- 应用配置 -->
    <div v-show="activeTab === 'apps'" class="flex flex-col flex-1 overflow-hidden">
      <!-- 工具栏 -->
      <div class="flex flex-wrap items-center gap-2 p-4">
        <el-button type="primary" @click="openAppDialog()">
          <el-icon><Plus /></el-icon>
          新增应用
        </el-button>
      </div>

      <!-- 统计信息 -->
      <div
        class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-y border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-center gap-6 text-sm">
          <span class="flex items-center gap-1">
            <el-icon color="var(--primary-color)"><Connection /></el-icon>
            应用总数：<span class="font-medium">{{ appTotal }}</span>
          </span>
          <span class="flex items-center gap-1">
            <el-icon color="var(--success-color)"><CircleCheck /></el-icon>
            启用：<span class="font-medium">{{ appEnabledCount }}</span>
          </span>
          <span class="flex items-center gap-1">
            <el-icon color="var(--danger-color)"><CircleClose /></el-icon>
            禁用：<span class="font-medium">{{ appDisabledCount }}</span>
          </span>
        </div>
      </div>

      <!-- 表格区域 -->
      <div class="flex-1 overflow-hidden">
        <el-table :data="apps" :height="'100%'" stripe v-loading="appLoading">
          <el-table-column type="index" label="序号" width="70" />
          <el-table-column
            prop="appKey"
            label="实例标志(app_key)"
            min-width="150"
            show-overflow-tooltip
          />
          <el-table-column prop="appType" label="类型" width="90">
            <template #default="{ row }">
              <el-tag :type="row.appType === 'mp' ? 'success' : 'warning'">
                {{ row.appType === 'mp' ? '服务号' : '小程序' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="appId" label="AppID" width="200" show-overflow-tooltip />
          <el-table-column label="默认商户" width="180" show-overflow-tooltip>
            <template #default="{ row }">
              {{
                row.defaultMerchantMchId
                  ? `${row.defaultMerchantMchId}${row.defaultMerchantName ? '（' + row.defaultMerchantName + '）' : ''}`
                  : '未设置'
              }}
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="140" show-overflow-tooltip />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="300" fixed="right">
            <template #default="{ row }">
              <div class="flex flex-wrap gap-2">
                <el-button
                  v-if="row.appType === 'mp'"
                  size="small"
                  type="primary"
                  @click="openAuthURL(row)"
                >
                  授权入口
                </el-button>
                <el-button v-else size="small" type="primary" @click="openMiniappQrcode(row)">
                  小程序码
                </el-button>
                <el-button size="small" type="success" @click="openMerchants(row)"
                  >商户关联</el-button
                >
                <el-button size="small" type="warning" @click="openUIDialog(row)">UI配置</el-button>
                <el-button size="small" @click="openAppDialog(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="removeApp(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页区域 -->
      <div class="flex items-center justify-between px-4 py-3 border-t">
        <span class="text-sm text-gray-600 dark:text-gray-400">共 {{ appTotal }} 条</span>
        <el-pagination
          v-model:current-page="appPage"
          :page-size="appPageSize"
          :total="appTotal"
          layout="total, prev, pager, next"
          @current-change="loadApps"
        />
      </div>
    </div>

    <!-- 商户配置 -->
    <div v-show="activeTab === 'merchants'" class="flex flex-col flex-1 overflow-hidden">
      <!-- 工具栏 -->
      <div class="flex flex-wrap items-center gap-2 p-4">
        <el-button type="primary" @click="openMerchantDialog()">
          <el-icon><Plus /></el-icon>
          新增商户
        </el-button>
      </div>

      <!-- 统计信息 -->
      <div
        class="px-4 py-2 bg-gray-50 dark:bg-gray-800 border-y border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-center gap-6 text-sm">
          <span class="flex items-center gap-1">
            <el-icon color="var(--primary-color)"><Coin /></el-icon>
            商户总数：<span class="font-medium">{{ merchantTotal }}</span>
          </span>
          <span class="flex items-center gap-1">
            <el-icon color="var(--success-color)"><CircleCheck /></el-icon>
            启用：<span class="font-medium">{{ merchantEnabledCount }}</span>
          </span>
          <span class="flex items-center gap-1">
            <el-icon color="var(--danger-color)"><CircleClose /></el-icon>
            禁用：<span class="font-medium">{{ merchantDisabledCount }}</span>
          </span>
        </div>
      </div>

      <!-- 表格区域 -->
      <div class="flex-1 overflow-hidden">
        <el-table :data="merchants" :height="'100%'" stripe v-loading="merchantLoading">
          <el-table-column type="index" label="序号" width="70" />
          <el-table-column prop="mchId" label="商户号" width="150" />
          <el-table-column prop="mchName" label="商户名称" width="140" show-overflow-tooltip />
          <el-table-column
            prop="mchSerialNo"
            label="证书序列号"
            min-width="240"
            show-overflow-tooltip
          />
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'info'">
                {{ row.status === 1 ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <div class="flex flex-wrap gap-2">
                <el-button size="small" @click="openMerchantDialog(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="removeMerchant(row)">删除</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页区域 -->
      <div class="flex items-center justify-between px-4 py-3 border-t">
        <span class="text-sm text-gray-600 dark:text-gray-400">共 {{ merchantTotal }} 条</span>
        <el-pagination
          v-model:current-page="merchantPage"
          :page-size="merchantPageSize"
          :total="merchantTotal"
          layout="total, prev, pager, next"
          @current-change="loadMerchants"
        />
      </div>
    </div>

    <!-- 应用编辑弹窗 -->
    <el-dialog
      v-model="appDialogVisible"
      :title="appForm.id ? '编辑应用' : '新增应用'"
      width="760px"
    >
      <el-form :model="appForm" label-width="130px">
        <el-divider content-position="left">基本信息</el-divider>
        <el-form-item label="实例标志(app_key)" required>
          <el-input
            v-model="appForm.appKey"
            placeholder="如 wxapp-main / mini-demo"
            :disabled="!!appForm.id"
          />
          <div class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">
            前端实例唯一标志，需与 wxapp / miniapp 部署时配置的 APP_KEY
            一致，登录与支付按它路由到本应用；创建后自动绑定路由到当前租户（无需再到租户详情手动绑定）
          </div>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="应用类型" required>
              <el-select v-model="appForm.appType" style="width: 100%">
                <el-option label="服务号(H5)" value="mp" />
                <el-option label="小程序" value="miniapp" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="AppID" required>
              <el-input v-model="appForm.appId" placeholder="微信 AppID，如 wx079ea8f5e3d25370" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="AppSecret">
          <el-input v-model="appForm.appSecret" placeholder="留空或******表示保留原值" />
        </el-form-item>
        <el-form-item label-width="0">
          <div class="-mt-3.5 text-xs leading-relaxed text-gray-500 dark:text-gray-400">
            默认商户不在本表单设置：保存后请通过「商户关联」绑定商户并设为默认（下单未指定商户号时使用）
          </div>
        </el-form-item>

        <!-- 网页授权配置：仅服务号(H5)需要，小程序通过 wx.login 静默登录 -->
        <template v-if="appForm.appType === 'mp'">
          <el-divider content-position="left">网页授权配置（服务号）</el-divider>
          <el-form-item label="授权scope">
            <el-select v-model="appForm.authScope" style="width: 100%">
              <el-option label="snsapi_base(静默)" value="snsapi_base" />
              <el-option label="snsapi_userinfo(弹窗)" value="snsapi_userinfo" />
            </el-select>
          </el-form-item>
          <el-form-item label="授权回调地址">
            <el-input
              v-model="appForm.oauthCallbackUrl"
              placeholder="如 https://ixpay.zcyxiaokeai.live/api/wx/auth/callback"
            />
            <div class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">
              微信网页授权回调地址（公众号后台需配置同名网页授权域名）；留空使用系统全局配置
            </div>
          </el-form-item>
          <el-form-item label="前端地址">
            <el-input
              v-model="appForm.frontendUrl"
              placeholder="如 https://ixpay.zcyxiaokeai.live/wxapp"
            />
            <div class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">
              授权登录成功后跳转的 H5 前端地址（wxapp 部署地址）；留空使用系统全局配置
            </div>
          </el-form-item>
        </template>
        <el-form-item v-else label-width="0">
          <el-alert
            type="info"
            :closable="false"
            title="小程序通过 wx.login 静默登录，无需网页授权相关配置"
            class="w-full"
          />
        </el-form-item>

        <el-divider content-position="left">其他</el-divider>
        <el-row :gutter="16">
          <el-col :span="16">
            <el-form-item label="描述">
              <el-input v-model="appForm.description" placeholder="仅用于平台内备注识别" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="状态">
              <el-switch v-model="appForm.status" :active-value="1" :inactive-value="0" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="appDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveApp">保存</el-button>
      </template>
    </el-dialog>

    <!-- 商户编辑弹窗 -->
    <el-dialog
      v-model="merchantDialogVisible"
      :title="merchantForm.id ? '编辑商户' : '新增商户'"
      width="760px"
    >
      <el-form :model="merchantForm" label-width="140px">
        <el-divider content-position="left">基础信息</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="商户号(mch_id)" required>
              <el-input v-model="merchantForm.mchId" placeholder="如 1900000109" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="商户名称">
              <el-input v-model="merchantForm.mchName" placeholder="如 测试商户" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label-width="0">
          <div
            class="-mt-3.5 text-xs leading-relaxed text-gray-500 dark:text-gray-400"
          >
            商户号：微信支付商户平台的商户号（纯数字），登录 pay.weixin.qq.com → 账户中心 →
            商户信息可查看；商户名称仅用于平台内展示识别
          </div>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="证书序列号" required>
              <el-input
                v-model="merchantForm.mchSerialNo"
                placeholder="如 5157F09EFDC096DE15EBE81A47057A725B6E1F2A"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="APIv3密钥" required>
              <el-input
                v-model="merchantForm.apiv3Key"
                placeholder="32 位 APIv3 密钥；留空或******表示保留原值"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label-width="0">
          <div
            class="-mt-3.5 text-xs leading-relaxed text-gray-500 dark:text-gray-400"
          >
            证书序列号：商户 API 证书序列号（40 位十六进制），商户平台 → API安全 → API证书
            可查看；APIv3密钥：商户平台 → API安全 → 设置APIv3密钥 时自定义的 32
            位字符串，编辑时留空或输入 ****** 表示保留原值
          </div>
        </el-form-item>

        <el-divider content-position="left">密钥凭据（AES 加密存储）</el-divider>
        <el-form-item label="商户私钥PEM" required>
          <el-input
            v-model="merchantForm.privateKey"
            type="textarea"
            :rows="5"
            placeholder="apiclient_key.pem 完整内容（含 BEGIN/END 行）；留空或******表示保留原值"
          />
          <div class="text-xs text-gray-500 dark:text-gray-400">
            商户 API 证书压缩包内的 apiclient_key.pem 文件全部内容；编辑时留空或输入 ******
            表示保留原值
          </div>
        </el-form-item>
        <el-form-item label="微信支付公钥ID">
          <el-input
            v-model="merchantForm.publicKeyId"
            placeholder="如 PUB_KEY_ID_0000000000000000000000000000；留空或******表示保留原值"
          />
          <div class="text-xs text-gray-500 dark:text-gray-400">
            使用「微信支付公钥」模式时填写（商户平台 → API安全 →
            微信支付公钥）；证书模式可留空；编辑时留空或输入 ****** 表示保留原值
          </div>
        </el-form-item>
        <el-form-item label="微信支付公钥PEM">
          <el-input
            v-model="merchantForm.publicKey"
            type="textarea"
            :rows="4"
            placeholder="公钥文件（pub_key.pem）完整内容；证书模式可留空；留空或******表示保留原值"
          />
          <div class="text-xs text-gray-500 dark:text-gray-400">
            与公钥ID配套的公钥文件内容（含 BEGIN/END
            行）；使用平台证书模式时可留空；编辑时留空或输入 ****** 表示保留原值
          </div>
        </el-form-item>

        <el-divider content-position="left">回调地址</el-divider>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="支付回调地址">
              <el-input
                v-model="merchantForm.payNotifyUrl"
                placeholder="如 https://ixpay.zcyxiaokeai.live/api/wx/pay/notify/wechat"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="退款回调地址">
              <el-input
                v-model="merchantForm.refundNotifyUrl"
                placeholder="如 https://ixpay.zcyxiaokeai.live/api/wx/pay/notify/refund"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label-width="0">
          <div
            class="-mt-3.5 text-xs leading-relaxed text-gray-500 dark:text-gray-400"
          >
            微信支付/退款结果异步通知地址，必须为 https 外网可达；留空使用系统全局配置
          </div>
        </el-form-item>

        <el-divider content-position="left">其他</el-divider>
        <el-form-item label="状态">
          <el-switch v-model="merchantForm.status" :active-value="1" :inactive-value="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="merchantDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveMerchant">保存</el-button>
      </template>
    </el-dialog>

    <!-- UI 配置编辑弹窗 -->
    <el-dialog
      v-model="uiDialogVisible"
      :title="`租户 UI 配置 - ${uiApp?.appKey || ''}`"
      width="720px"
    >
      <div class="mb-3">
        <el-alert type="info" :closable="false" class="whitespace-pre-line">
          <template #title>
            <div class="leading-relaxed text-left">
              字段说明（前端启动时动态加载生效，无需发版）：<br />
              · brand：brand.title 标题 / brand.logo_url 品牌图 / brand.primary_color 主题色<br />
              · home：home.banners 轮播图数组 [{image_url, link?}]（兼容旧版单张
              home.banner_url）/ home.notice 公告 / home.grid 宫格菜单
              [{icon,title,path,visible}]<br />
              · tabbar：tabbar.items 底部导航 [{icon,title,path}]（默认：首页 / 个人中心）<br />
              · profile：profile.menu 个人中心菜单 [{icon?,title,path}]（默认：待支付列表 /
              订单支付）
            </div>
          </template>
        </el-alert>
      </div>
      <el-input
        v-model="uiJSON"
        type="textarea"
        :rows="16"
        :placeholder="uiExample"
      />
      <template #footer>
        <el-button @click="uiDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveUI">保存</el-button>
      </template>
    </el-dialog>

    <!-- 商户关联弹窗 -->
    <el-dialog
      v-model="bindDialogVisible"
      :title="`商户关联 - ${bindApp?.appKey || ''}`"
      width="720px"
    >
      <el-alert
        title="默认商户在此唯一设置：应用下单未指定商户号时使用；一个应用可关联多个商户，仅一个默认，解除默认商户关联时剩余最早关联的商户自动接任"
        type="info"
        :closable="false"
        class="mb-4"
      />

      <!-- 已关联商户 -->
      <div class="flex flex-wrap items-center justify-between gap-2 mb-2">
        <span class="flex items-center gap-1 text-sm font-medium">
          <el-icon color="var(--primary-color)"><Coin /></el-icon>
          已关联商户（{{ bindList.length }}）
        </span>
      </div>
      <el-table :data="bindList" border size="small" max-height="320">
        <el-table-column prop="mchId" label="商户号" width="150" />
        <el-table-column prop="mchName" label="名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="默认" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.isDefault" type="success" size="small">默认</el-tag>
            <el-button v-else size="small" type="warning" @click="setDefaultMerchant(row)">
              设为默认
            </el-button>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-button size="small" type="danger" @click="doUnbind(row)">解除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-divider content-position="left">关联新商户</el-divider>
      <div class="flex flex-wrap items-center gap-3">
        <el-select
          v-model="bindMerchantID"
          placeholder="选择要关联的商户"
          filterable
          style="width: 320px"
        >
          <el-option
            v-for="m in bindableMerchants"
            :key="m.id"
            :label="`${m.mchName || m.mchId}(${m.mchId})`"
            :value="m.id"
          />
        </el-select>
        <el-button type="primary" @click="doBind">
          <el-icon><Plus /></el-icon>
          关联
        </el-button>
        <span class="text-xs text-gray-500 dark:text-gray-400">首个关联的商户自动设为默认</span>
      </div>
    </el-dialog>

    <!-- 授权入口弹窗（服务号）：链接 + 二维码 -->
    <el-dialog
      v-model="authUrlDialogVisible"
      :title="`授权入口 - ${authUrlApp?.appKey || ''}`"
      width="520px"
    >
      <el-alert
        title="微信扫码进入授权登录（回跳支付页）；也可复制链接配置到公众号菜单 / 测试号"
        type="info"
        :closable="false"
        class="mb-4"
      />
      <div class="flex flex-col items-center gap-3" v-loading="authQrcodeLoading">
        <img
          v-if="authQrcodeUrl"
          :src="authQrcodeUrl"
          alt="授权入口二维码"
          class="w-52 h-52 rounded-lg border border-gray-200 dark:border-gray-700"
        />
        <div v-if="authUrlText" class="w-full rounded bg-gray-50 dark:bg-gray-800 p-3">
          <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">授权链接</div>
          <div class="text-xs font-mono break-all">{{ authUrlText }}</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="copyAuthUrl" :disabled="!authUrlText">复制链接</el-button>
        <el-button @click="downloadAuthQrcode" :disabled="!authQrcodeUrl">下载二维码</el-button>
        <el-button @click="authUrlDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 小程序码弹窗 -->
    <el-dialog
      v-model="qrcodeDialogVisible"
      :title="`小程序入口码 - ${qrcodeApp?.appKey || ''}`"
      width="480px"
    >
      <el-alert
        title="微信扫码进入小程序默认支付页（scene 携带 app_key 自动归属本应用）；可右键或点击下方按钮保存图片，配置到线下物料"
        type="info"
        :closable="false"
        class="mb-4"
      />
      <div class="flex flex-col items-center gap-3" v-loading="qrcodeLoading">
        <img
          v-if="qrcodeUrl"
          :src="qrcodeUrl"
          alt="小程序码"
          class="w-56 h-56 rounded-lg border border-gray-200 dark:border-gray-700"
        />
        <span v-if="qrcodeApp" class="text-xs text-gray-500 dark:text-gray-400">
          AppID：{{ qrcodeApp.appId }}
        </span>
      </div>
      <template #footer>
        <el-button @click="downloadQrcode" :disabled="!qrcodeUrl">下载图片</el-button>
        <el-button @click="qrcodeDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Connection, Coin, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import {
  getWxAppList,
  createWxApp,
  updateWxApp,
  deleteWxApp,
  getWxAppAuthURL,
  getWxAppAuthQrcode,
  getWxAppMiniappQrcode,
  getWxMerchantList,
  createWxMerchant,
  updateWxMerchant,
  deleteWxMerchant,
  getWxAppMerchants,
  bindWxAppMerchant,
  unbindWxAppMerchant,
  getWxAppUI,
  saveWxAppUI,
  type WxAppConfig,
  type WxMerchantConfig,
} from '@/api/modules/wx-channel'
import TenantSwitcher from '@/components/TenantSwitcher/index.vue'
import { useTenantStore } from '@/stores/modules/tenant'

const tenant = useTenantStore()
const route = useRoute()
const router = useRouter()

// 切换租户后重载本页数据（渠道配置随 ?tenant_id= 路由到对应租户库）
function onTenantChange() {
  appPage.value = 1
  merchantPage.value = 1
  loadApps()
  loadMerchants()
}

const activeTab = ref('apps')

// 应用列表
const apps = ref<WxAppConfig[]>([])
const appLoading = ref(false)
const appPage = ref(1)
const appPageSize = ref(10)
const appTotal = ref(0)

// 商户列表
const merchants = ref<WxMerchantConfig[]>([])
const merchantLoading = ref(false)
const merchantPage = ref(1)
const merchantPageSize = ref(10)
const merchantTotal = ref(0)

// 统计信息（基于当前页数据）
const appEnabledCount = computed(() => apps.value.filter((item) => item.status === 1).length)
const appDisabledCount = computed(() => apps.value.filter((item) => item.status !== 1).length)
const merchantEnabledCount = computed(
  () => merchants.value.filter((item) => item.status === 1).length,
)
const merchantDisabledCount = computed(
  () => merchants.value.filter((item) => item.status !== 1).length,
)

// 弹窗
const appDialogVisible = ref(false)
const appForm = ref<Partial<WxAppConfig>>({})
const merchantDialogVisible = ref(false)
const merchantForm = ref<Partial<WxMerchantConfig>>({})

// 关联
const bindDialogVisible = ref(false)
const bindApp = ref<WxAppConfig | null>(null)
const bindList = ref<Record<string, unknown>[]>([])
const bindMerchantID = ref<string>('')

// 小程序码
const qrcodeDialogVisible = ref(false)
const qrcodeApp = ref<WxAppConfig | null>(null)
const qrcodeUrl = ref('')
const qrcodeLoading = ref(false)

// 服务号授权入口（链接 + 二维码）
const authUrlDialogVisible = ref(false)
const authUrlApp = ref<WxAppConfig | null>(null)
const authUrlText = ref('')
const authQrcodeUrl = ref('')
const authQrcodeLoading = ref(false)

// UI 配置
const uiDialogVisible = ref(false)
const uiApp = ref<WxAppConfig | null>(null)
const uiJSON = ref('')

// UI 配置示例（覆盖轮播/宫格/底部导航/个人中心菜单，供 placeholder 参考）
const uiExample = `{
  "version": 2,
  "brand": { "title": "xxx 支付", "logo_url": "https://.../logo.png", "primary_color": "#07c160" },
  "home": {
    "notice": "欢迎使用",
    "banners": [
      { "image_url": "https://.../banner1.jpg", "link": "/payment" },
      { "image_url": "https://.../banner2.jpg" }
    ],
    "grid": [
      { "icon": "📋", "title": "待支付列表", "path": "/pending-list", "visible": true },
      { "icon": "💳", "title": "订单支付", "path": "/payment", "visible": true }
    ]
  },
  "tabbar": {
    "items": [
      { "icon": "🏠", "title": "首页", "path": "/" },
      { "icon": "👤", "title": "个人中心", "path": "/profile" }
    ]
  },
  "profile": {
    "menu": [
      { "icon": "📋", "title": "待支付列表", "path": "/pending-list" },
      { "icon": "💳", "title": "订单支付", "path": "/payment" }
    ]
  }
}`

async function loadApps() {
  appLoading.value = true
  try {
    const res = await getWxAppList({
      page: appPage.value,
      pageSize: appPageSize.value,
    })
    apps.value = res.data!.list
    appTotal.value = res.data!.total
  } finally {
    appLoading.value = false
  }
}

async function loadMerchants() {
  merchantLoading.value = true
  try {
    const res = await getWxMerchantList({
      page: merchantPage.value,
      pageSize: merchantPageSize.value,
    })
    merchants.value = res.data!.list
    merchantTotal.value = res.data!.total
  } finally {
    merchantLoading.value = false
  }
}

function openAppDialog(row?: WxAppConfig) {
  // 默认商户不在本表单设置（响应中的 defaultMerchant* 为只读展示字段，
  // 提交时后端忽略），唯一入口为「商户关联」弹窗
  appForm.value = row
    ? { ...row }
    : { appType: 'mp', authScope: 'snsapi_base', status: 1 }
  appDialogVisible.value = true
}

async function saveApp() {
  try {
    // 小程序无网页授权概念，提交前清空服务号专属字段，避免切换类型后残留旧值
    if (appForm.value.appType === 'miniapp') {
      appForm.value.authScope = ''
      appForm.value.oauthCallbackUrl = ''
      appForm.value.frontendUrl = ''
    }
    if (appForm.value.id) {
      await updateWxApp(appForm.value.id, appForm.value)
      ElMessage.success('保存成功')
    } else {
      await createWxApp(appForm.value)
      ElMessage.success('创建成功，应用路由已自动绑定到当前租户')
    }
    appDialogVisible.value = false
    loadApps()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function removeApp(row: WxAppConfig) {
  await ElMessageBox.confirm(`确定删除应用 ${row.appKey} 吗？`, '提示', { type: 'warning' })
  await deleteWxApp(row.id)
  ElMessage.success('删除成功')
  loadApps()
}

// 打开服务号授权入口弹窗：并行拉取授权链接与二维码 PNG（Blob → ObjectURL）
async function openAuthURL(row: WxAppConfig) {
  authUrlApp.value = row
  authUrlText.value = ''
  authQrcodeUrl.value = ''
  authUrlDialogVisible.value = true
  authQrcodeLoading.value = true
  try {
    const [urlRes, pngBlob] = await Promise.all([
      getWxAppAuthURL(row.id),
      getWxAppAuthQrcode(row.id),
    ])
    authUrlText.value = urlRes.data?.authUrl ?? ''
    authQrcodeUrl.value = URL.createObjectURL(pngBlob)
  } catch (e) {
    ElMessage.error((e as Error).message || '获取授权入口失败')
    authUrlDialogVisible.value = false
  } finally {
    authQrcodeLoading.value = false
  }
}

// 复制授权链接到剪贴板
async function copyAuthUrl() {
  if (!authUrlText.value) return
  try {
    await navigator.clipboard.writeText(authUrlText.value)
    ElMessage.success('链接已复制')
  } catch {
    ElMessage.error('复制失败，请手动选择链接复制')
  }
}

// 下载授权入口二维码图片
function downloadAuthQrcode() {
  if (!authQrcodeUrl.value || !authUrlApp.value) return
  const a = document.createElement('a')
  a.href = authQrcodeUrl.value
  a.download = `auth-qrcode-${authUrlApp.value.appKey}.png`
  a.click()
}

// 打开小程序码弹窗并加载小程序码 PNG（Blob → ObjectURL）
async function openMiniappQrcode(row: WxAppConfig) {
  qrcodeApp.value = row
  qrcodeUrl.value = ''
  qrcodeDialogVisible.value = true
  qrcodeLoading.value = true
  try {
    const blob = await getWxAppMiniappQrcode(row.id)
    qrcodeUrl.value = URL.createObjectURL(blob)
  } catch (e) {
    ElMessage.error((e as Error).message || '生成小程序码失败')
    qrcodeDialogVisible.value = false
  } finally {
    qrcodeLoading.value = false
  }
}

// 下载小程序码图片
function downloadQrcode() {
  if (!qrcodeUrl.value || !qrcodeApp.value) return
  const a = document.createElement('a')
  a.href = qrcodeUrl.value
  a.download = `miniapp-qrcode-${qrcodeApp.value.appKey}.png`
  a.click()
}

function openMerchantDialog(row?: WxMerchantConfig) {
  merchantForm.value = row ? { ...row } : { status: 1 }
  merchantDialogVisible.value = true
}

async function saveMerchant() {
  try {
    if (merchantForm.value.id) {
      await updateWxMerchant(merchantForm.value.id, merchantForm.value)
    } else {
      await createWxMerchant(merchantForm.value)
    }
    ElMessage.success('保存成功')
    merchantDialogVisible.value = false
    loadMerchants()
  } catch (e) {
    ElMessage.error((e as Error).message || '保存失败')
  }
}

async function removeMerchant(row: WxMerchantConfig) {
  await ElMessageBox.confirm(`确定删除商户 ${row.mchId} 吗？`, '提示', { type: 'warning' })
  await deleteWxMerchant(row.id)
  ElMessage.success('删除成功')
  loadMerchants()
}

async function openMerchants(row: WxAppConfig) {
  bindApp.value = row
  bindDialogVisible.value = true
  await refreshBindList()
}

async function refreshBindList() {
  if (!bindApp.value) return
  const res = await getWxAppMerchants(bindApp.value.id)
  bindList.value = res.data!.list as unknown as Record<string, unknown>[]
}

// 可关联商户 = 全部启用商户 - 已关联商户（避免重复关联报错）
const bindableMerchants = computed(() =>
  merchants.value.filter((m) => !bindList.value.some((b) => String(b.mchId) === String(m.mchId))),
)

async function doBind() {
  if (!bindApp.value || !bindMerchantID.value) {
    ElMessage.warning('请选择要关联的商户')
    return
  }
  // 默认商户兜底由后端保证：应用当前无默认时本次关联自动设为默认
  const isFirstBind = bindList.value.length === 0
  await bindWxAppMerchant(bindApp.value.id, {
    merchantId: bindMerchantID.value,
    isDefault: false,
  })
  ElMessage.success(isFirstBind ? '关联成功，已自动设为默认商户' : '关联成功')
  bindMerchantID.value = ''
  refreshBindList()
}

// 设为默认商户（先清除旧默认，再绑定 isDefault=true）
async function setDefaultMerchant(row: Record<string, unknown>) {
  if (!bindApp.value) return
  await bindWxAppMerchant(bindApp.value.id, {
    merchantId: row.merchantConfigId as string,
    isDefault: true,
  })
  ElMessage.success(`商户 ${row.mchId} 已设为默认`)
  refreshBindList()
}

async function doUnbind(row: Record<string, unknown>) {
  if (!bindApp.value) return
  await unbindWxAppMerchant(bindApp.value.id, row.merchantConfigId as string)
  ElMessage.success('已解除')
  refreshBindList()
}

async function openUIDialog(row: WxAppConfig) {
  uiApp.value = row
  try {
    const res = await getWxAppUI(row.id)
    uiJSON.value = res.data?.uiJson || ''
  } catch {
    uiJSON.value = ''
  }
  uiDialogVisible.value = true
}

async function saveUI() {
  if (!uiApp.value) return
  try {
    JSON.parse(uiJSON.value)
  } catch {
    ElMessage.error('UI 配置不是合法 JSON')
    return
  }
  await saveWxAppUI(uiApp.value.id, { uiJson: uiJSON.value })
  ElMessage.success('保存成功，前端下次启动生效')
  uiDialogVisible.value = false
}

onMounted(async () => {
  // 支持从租户详情「前往渠道配置」携带 tenant_id 跳转：自动选中该租户（免二次切换）后加载数据
  const queryTenantId = route.query.tenant_id as string
  if (queryTenantId) {
    await tenant.loadTenants()
    tenant.setCurrentTenant(queryTenantId)
    router.replace({ query: { ...route.query, tenant_id: undefined } })
  }
  loadApps()
  loadMerchants()
})
</script>
