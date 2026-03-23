<template>
  <div class="subscription-page" v-loading="loading">
    <div class="card-container">
      <el-card shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-left">
              <el-icon size="20"><svg viewBox="0 0 1024 1024" width="20" height="20"><path fill="currentColor" d="M128 192h768v128H192v640h640v-64h64v128H128V192z"></path><path fill="currentColor" d="M384 384h384v64H384v320h320v-64h64v128H384V384z"></path></svg></el-icon>
              <span>订阅链接</span>
            </div>
            <el-button
              type="primary"
              :icon="RefreshRight"
              :loading="refreshing"
              @click="handleRefreshToken"
              size="small"
            >
              刷新Token
            </el-button>
          </div>
        </template>

        <el-form label-width="100px">
          <el-form-item label="Clash订阅">
            <el-input
              :model-value="clashUrl"
              readonly
            >
              <template #prepend>
                <el-tag type="success" size="small">Clash</el-tag>
              </template>
              <template #append>
                <el-button @click="copyClashUrl">复制</el-button>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="Sing-Box订阅">
            <el-input
              :model-value="singboxUrl"
              readonly
            >
              <template #prepend>
                <el-tag type="warning" size="small">Sing-Box</el-tag>
              </template>
              <template #append>
                <el-button @click="copySingboxUrl">复制</el-button>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item label="订阅Token">
            <el-input
              :model-value="token"
              readonly
            >
              <template #append>
                <el-button @click="copyToken">复制</el-button>
              </template>
            </el-input>
          </el-form-item>
        </el-form>

        <el-alert
          title="使用说明"
          type="info"
          :closable="false"
          show-icon
        >
          <p>1. 根据客户端类型复制对应的订阅链接</p>
          <p>2. <strong>Clash订阅</strong>：适用于 Clash、Clash Verge、Clash Meta 等客户端</p>
          <p>3. <strong>Sing-Box订阅</strong>：适用于 Sing-Box、SFM 等客户端</p>
          <p>4. 如需更换订阅地址，点击右上角"刷新Token"按钮</p>
        </el-alert>
      </el-card>

      <el-card shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-left">
              <el-icon size="20"><svg viewBox="0 0 1024 1024" width="20" height="20"><path fill="currentColor" d="M128 192h768v128H192v640h640v-64h64v128H128V192z"></path><path fill="currentColor" d="M384 384h384v64H384v320h320v-64h64v128H384V384z"></path></svg></el-icon>
              <span>二维码</span>
            </div>
            <el-radio-group v-model="qrcodeType" size="small">
              <el-radio-button label="clash">Clash</el-radio-button>
              <el-radio-button label="singbox">Sing-Box</el-radio-button>
            </el-radio-group>
          </div>
        </template>

        <div class="qrcode-container">
          <img :src="qrcodeUrl" alt="订阅二维码" />
          <p>{{ qrcodeType === 'clash' ? '使用Clash客户端扫描二维码导入' : '使用Sing-Box客户端扫描二维码导入' }}</p>
        </div>
      </el-card>
    </div>

    <el-card shadow="never" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon size="20"><svg viewBox="0 0 1024 1024" width="20" height="20"><path fill="currentColor" d="M128 192h768v128H192v640h640v-64h64v128H128V192z"></path><path fill="currentColor" d="M384 384h384v64H384v320h320v-64h64v128H384V384z"></path></svg></el-icon>
            <span>配置预览</span>
          </div>
          <div style="display: flex; gap: 10px; align-items: center;">
            <el-radio-group v-model="previewFormat" size="small">
              <el-radio-button label="clash">Clash</el-radio-button>
              <el-radio-button label="singbox">Sing-Box</el-radio-button>
            </el-radio-group>
            <el-button type="success" :icon="CircleCheck" @click="validateConfig" :loading="validating">校验配置</el-button>
            <el-button type="warning" :icon="Delete" @click="cleanupRules" :loading="cleaning">清理无效规则</el-button>
            <el-button type="primary" :icon="View" @click="loadConfig">加载配置</el-button>
          </div>
        </div>
      </template>

      <pre v-if="configContent" class="config-content">{{ configContent }}</pre>
      <el-empty v-else description="点击上方按钮加载配置" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { View, RefreshRight, CircleCheck, Delete } from '@element-plus/icons-vue'
import { getSubscriptionURL, refreshToken, validateConfig as validateConfigAPI, cleanupInvalidRules } from '@/api/subscription'
import axios from 'axios'

const subscriptionUrl = ref('')
const clashUrl = ref('')
const singboxUrl = ref('')
const token = ref('')
const qrcodeType = ref('clash')
const previewFormat = ref('clash')
const configContent = ref('')
const loading = ref(false)
const refreshing = ref(false)
const validating = ref(false)
const cleaning = ref(false)

// 计算当前显示的二维码URL
const qrcodeUrl = computed(() => {
  const url = qrcodeType.value === 'clash' ? clashUrl.value : singboxUrl.value
  return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(url)}`
})

// 加载订阅信息
const loadSubscription = async () => {
  loading.value = true
  try {
    const data = await getSubscriptionURL()
    subscriptionUrl.value = data.url
    clashUrl.value = data.clash_url || data.url
    singboxUrl.value = data.singbox_url || (data.url + '?format=singbox')
    token.value = data.token
  } catch {
    ElMessage.error('加载订阅信息失败')
  } finally {
    loading.value = false
  }
}

// 刷新Token
const handleRefreshToken = async () => {
  refreshing.value = true
  try {
    const data = await refreshToken()
    ElMessage.success('Token已刷新')
    await loadSubscription()
  } catch {
    ElMessage.error('刷新Token失败')
  } finally {
    refreshing.value = false
  }
}

// 复制Clash URL
const copyClashUrl = async () => {
  try {
    await navigator.clipboard.writeText(clashUrl.value)
    ElMessage.success('Clash订阅链接已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 复制Sing-Box URL
const copySingboxUrl = async () => {
  try {
    await navigator.clipboard.writeText(singboxUrl.value)
    ElMessage.success('Sing-Box订阅链接已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 复制Token
const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(token.value)
    ElMessage.success('Token已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 加载配置预览
const loadConfig = async () => {
  if (!token.value) {
    ElMessage.error('Token不存在，请刷新页面')
    return
  }
  try {
    let url = `/sub/${token.value}`
    let headers = {}

    if (previewFormat.value === 'singbox') {
      url += '?format=singbox'
      headers['Accept'] = 'application/json'
    } else {
      headers['Accept'] = 'application/yaml'
    }

    const response = await axios.get(url, { headers })
    configContent.value = typeof response.data === 'object'
      ? JSON.stringify(response.data, null, 2)
      : response.data
    ElMessage.success('配置加载成功')
  } catch (err) {
    console.error('加载配置失败:', err)
    ElMessage.error('加载配置失败: ' + (err.response?.data?.error || err.message))
  }
}

// 校验配置
const validateConfig = async () => {
  validating.value = true
  try {
    const result = await validateConfigAPI()

    if (result.valid) {
      if (result.errors && result.errors.length > 0) {
        // 有警告但没有错误
        const warnings = result.errors.filter(e => e.type === 'warning')
        if (warnings.length > 0) {
          ElMessage.warning(`配置校验通过，但有 ${warnings.length} 个警告，下方查看详情`)

          // 显示警告详情
          const warningDetails = warnings.map((e, i) => {
            return `⚠️ 警告 ${i + 1}: ${e.message}\n${e.suggestion ? '💡 建议: ' + e.suggestion : ''}`
          }).join('\n\n')

          configContent.value = `配置校验结果：✅ 通过，但有 ${warnings.length} 个警告\n\n${warningDetails}`
        } else {
          ElMessage.success('配置校验通过！')
          configContent.value = ''
        }
      } else {
        ElMessage.success('配置校验通过！')
        configContent.value = ''
      }
    } else {
      // 有错误
      const errors = result.errors || []
      const errorCount = errors.filter(e => e.type === 'error').length
      const warningCount = errors.filter(e => e.type === 'warning').length

      let message = `配置校验失败：发现 ${errorCount} 个错误`
      if (warningCount > 0) {
        message += `，${warningCount} 个警告`
      }

      ElMessage.error(message)

      // 显示详细的错误信息
      const errorDetails = errors.map((e, i) => {
        const icon = e.type === 'error' ? '❌' : '⚠️'
        const typeLabel = e.type === 'error' ? '错误' : '警告'
        return `${icon} ${typeLabel} ${i + 1}: ${e.message}\n${e.suggestion ? '💡 建议: ' + e.suggestion : ''}`
      }).join('\n\n')

      configContent.value = `配置校验结果：\n\n${errorDetails}`
    }
  } catch (err) {
    console.error('校验配置失败:', err)
    ElMessage.error('校验失败：' + (err.response?.data?.error || err.message))
  } finally {
    validating.value = false
  }
}

// 清理无效规则
const cleanupRules = async () => {
  try {
    await ElMessageBox.confirm(
      '这将删除所有引用不存在节点或分组的规则，操作不可恢复。确定继续吗？',
      '确认清理',
      { type: 'warning' }
    )

    cleaning.value = true
    const result = await cleanupInvalidRules()

    if (result.deletedCount > 0) {
      ElMessage.success(`已清理 ${result.deletedCount} 条无效规则`)
    } else {
      ElMessage.info('没有发现无效规则')
    }
  } catch (err) {
    if (err !== 'cancel') {
      console.error('清理规则失败:', err)
      ElMessage.error('清理失败：' + (err.response?.data?.error || err.message))
    }
  } finally {
    cleaning.value = false
  }
}

onMounted(() => {
  loadSubscription()
})
</script>

<style scoped>
.subscription-page {
  height: 100%;
}

.card-container {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

:deep(.el-card__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

.qrcode-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
  gap: 15px;
}

.qrcode-container img {
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 10px;
  background: #fff;
}

.qrcode-container p {
  color: #909399;
  font-size: 14px;
  margin: 0;
}

.config-content {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.6;
  max-height: 500px;
  overflow-y: auto;
  color: #303133;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

:deep(.el-alert) {
  margin-top: 15px;
}

:deep(.el-alert p) {
  margin: 5px 0;
}

:deep(.el-empty) {
  padding: 40px 0;
}
</style>
