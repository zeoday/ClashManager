<template>
  <div class="rules-page">
    <el-card shadow="never" class="rules-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <div class="header-icon">
              <el-icon><DocumentCopy /></el-icon>
            </div>
            <span>规则列表</span>
          </div>
          <div class="header-actions">
            <el-button :icon="Upload" @click="showImportDialog">导入规则</el-button>
            <el-button type="primary" :icon="Plus" @click="showCreateDialog">新增规则</el-button>
          </div>
        </div>
      </template>

      <!-- 搜索和过滤区域 -->
      <div class="filter-section">
        <div class="filter-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索匹配内容、目标或备注"
            clearable
            class="search-input"
            @clear="handleSearchChange"
            @keyup.enter="handleSearchChange"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>

          <el-select
            v-model="filterType"
            placeholder="规则类型"
            clearable
            class="filter-select"
            @change="handleFilterTypeChange"
            @clear="handleFilterTypeClear"
          >
            <el-option label="DOMAIN-SUFFIX" value="DOMAIN-SUFFIX" />
            <el-option label="DOMAIN" value="DOMAIN" />
            <el-option label="DOMAIN-KEYWORD" value="DOMAIN-KEYWORD" />
            <el-option label="IP-CIDR" value="IP-CIDR" />
            <el-option label="GEOIP" value="GEOIP" />
            <el-option label="MATCH" value="MATCH" />
          </el-select>

          <el-select
            v-model="filterTarget"
            placeholder="目标类型"
            clearable
            class="filter-select-wide"
            @change="handleFilterTargetChange"
            @clear="handleFilterTargetClear"
          >
            <el-option-group label="固定出口">
              <el-option label="DIRECT - 直连" value="DIRECT">
                <div class="option-content">
                  <el-tag type="success" size="small">DIRECT</el-tag>
                  <span class="option-text">直连</span>
                </div>
              </el-option>
              <el-option label="PROXY - 代理" value="PROXY">
                <div class="option-content">
                  <el-tag type="primary" size="small">PROXY</el-tag>
                  <span class="option-text">代理</span>
                </div>
              </el-option>
              <el-option label="REJECT - 拒绝" value="REJECT">
                <div class="option-content">
                  <el-tag type="danger" size="small">REJECT</el-tag>
                  <span class="option-text">拒绝</span>
                </div>
              </el-option>
            </el-option-group>
            <el-option-group label="代理节点" v-if="nodes.length > 0">
              <el-option
                v-for="node in nodes"
                :key="'node-' + node.ID"
                :label="node.Name"
                :value="'node:' + node.Name"
              >
                <div class="option-content-flex">
                  <el-icon><Connection /></el-icon>
                  <span class="option-name">{{ node.Name }}</span>
                  <el-tag size="small" class="option-type-tag">{{ node.Type }}</el-tag>
                </div>
              </el-option>
            </el-option-group>
            <el-option-group label="代理组" v-if="groups.length > 0">
              <el-option
                v-for="group in groups"
                :key="'group-' + group.ID"
                :label="group.Name"
                :value="'group:' + group.Name"
              >
                <div class="option-content-flex">
                  <el-icon><Grid /></el-icon>
                  <span class="option-name">{{ group.Name }}</span>
                  <el-tag size="small" class="option-type-tag">{{ group.Type }}</el-tag>
                </div>
              </el-option>
            </el-option-group>
          </el-select>

          <el-button @click="resetFilter" :icon="RefreshLeft">重置</el-button>
        </div>
        <div class="filter-right">
          <span class="total-count">共 {{ total }} 条规则</span>
        </div>
      </div>

      <el-table :data="displayRules" stripe class="rules-table" v-loading="loading">
        <el-table-column prop="ID" label="ID" min-width="60" />
        <el-table-column prop="Priority" label="序号" min-width="70">
          <template #default="{ row }">
            <span class="priority-badge">{{ row.Priority ?? 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="Type" label="规则类型" min-width="140">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.Type)" size="small">{{ row.Type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="Payload" label="匹配内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="Target" label="目标" min-width="140">
          <template #default="{ row }">
            <div class="target-cell">
              <!-- 固定出口 -->
              <el-tag v-if="row.Target === 'PROXY'" type="primary" size="small" class="target-tag proxy">
                PROXY
              </el-tag>
              <el-tag v-else-if="row.Target === 'DIRECT'" type="success" size="small" class="target-tag direct">
                DIRECT
              </el-tag>
              <el-tag v-else-if="row.Target === 'REJECT'" type="danger" size="small" class="target-tag reject">
                REJECT
              </el-tag>
              <!-- 其他目标（节点名或代理组名） -->
              <div v-else class="target-tag other">
                <span>{{ row.Target }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="Remark" label="备注" min-width="150" show-overflow-tooltip />
        <el-table-column prop="CreatedAt" label="创建时间" min-width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.CreatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-section">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑规则对话框 -->
    <el-dialog v-model="formDialogVisible" :title="isEdit ? '编辑规则' : '新增规则'" width="520px" class="rule-dialog">
      <el-form :model="ruleForm" label-width="110px" class="rule-form">
        <el-form-item label="序号">
          <el-input-number v-model="ruleForm.Priority" :min="0" :max="9999" placeholder="数字越小优先级越高" class="form-input" />
          <div class="form-hint">数字越小优先级越高，0为最高优先级</div>
        </el-form-item>
        <el-form-item label="规则类型">
          <el-select v-model="ruleForm.Type" placeholder="请选择规则类型" class="form-input">
            <el-option label="DOMAIN-SUFFIX - 域名后缀匹配" value="DOMAIN-SUFFIX" />
            <el-option label="DOMAIN - 完整域名匹配" value="DOMAIN" />
            <el-option label="DOMAIN-KEYWORD - 域名关键字匹配" value="DOMAIN-KEYWORD" />
            <el-option label="IP-CIDR - IP段匹配" value="IP-CIDR" />
            <el-option label="GEOIP - 地理位置匹配" value="GEOIP" />
            <el-option label="MATCH - 全匹配（默认规则）" value="MATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配内容">
          <el-input v-model="ruleForm.Payload" placeholder="如: google.com 或 192.168.1.0/24" class="form-input" />
        </el-form-item>
        <el-form-item label="目标">
          <el-select v-model="ruleForm.Target" placeholder="请选择目标" class="form-input" @change="handleTargetChange" filterable>
            <el-option-group label="内置目标">
              <el-option label="PROXY - 代理" value="PROXY">
                <span>PROXY</span>
                <span class="option-desc">代理</span>
              </el-option>
              <el-option label="DIRECT - 直连" value="DIRECT">
                <span>DIRECT</span>
                <span class="option-desc">直连</span>
              </el-option>
              <el-option label="REJECT - 拒绝" value="REJECT">
                <span>REJECT</span>
                <span class="option-desc">拒绝</span>
              </el-option>
            </el-option-group>
            <el-option-group label="代理节点">
              <el-option
                v-for="node in nodes"
                :key="node.ID"
                :label="node.Name"
                :value="`node:${node.ID}:${node.Name}`"
              >
                <div class="option-content-flex">
                  <el-icon><Connection /></el-icon>
                  <span class="option-name">{{ node.Name }}</span>
                  <el-tag size="small" class="option-type-tag">{{ node.Type }}</el-tag>
                </div>
              </el-option>
            </el-option-group>
            <el-option-group label="代理组">
              <el-option
                v-for="group in groups"
                :key="group.ID"
                :label="group.Name"
                :value="`group:${group.ID}:${group.Name}`"
              >
                <div class="option-content-flex">
                  <el-icon><Grid /></el-icon>
                  <span class="option-name">{{ group.Name }}</span>
                  <el-tag size="small" class="option-type-tag">{{ group.Type }}</el-tag>
                </div>
              </el-option>
            </el-option-group>
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="ruleForm.Remark" placeholder="可选，用于记录规则用途" maxlength="200" show-word-limit class="form-input" />
          <div class="form-hint">备注仅用于展示，不会生成到配置文件中</div>
        </el-form-item>
        <el-form-item label="No Resolve">
          <el-switch v-model="ruleForm.NoResolve" />
          <span class="switch-label">不反查域名</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 导入规则对话框 -->
    <el-dialog v-model="importDialogVisible" title="导入规则" width="600px" class="import-dialog">
      <el-form label-width="100px" class="import-form">
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :show-file-list="true"
            :limit="1"
            accept=".yaml,.yml"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            drag
          >
            <el-icon class="upload-icon"><UploadFilled /></el-icon>
            <div class="upload-text">拖拽文件到此处或点击上传</div>
            <template #tip>
              <div class="upload-tip">支持 .yaml 或 .yml 格式文件，将自动解析 rules 节点</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="文件内容" v-if="importContent">
          <el-input
            v-model="importContent"
            type="textarea"
            :rows="10"
            placeholder="文件内容将显示在这里"
            readonly
            class="import-textarea"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImport" :loading="importing">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Search,
  Upload,
  RefreshLeft,
  DocumentCopy,
  Connection,
  Grid,
  Edit,
  Delete,
  UploadFilled
} from '@element-plus/icons-vue'
import { getRules, createRule, updateRule, deleteRule, importRules } from '@/api/rules'
import { getGroups } from '@/api/groups'
import { getNodes } from '@/api/nodes'

const rules = ref([])
const groups = ref([])
const nodes = ref([])
const formDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const loading = ref(false)

// 导入相关
const importDialogVisible = ref(false)
const importContent = ref('')
const importing = ref(false)

const ruleForm = ref({
  Type: 'DOMAIN-SUFFIX',
  Payload: '',
  Target: 'PROXY',
  TargetID: 0,
  TargetType: '',
  Priority: 0,
  NoResolve: false,
  Remark: ''
})

// 搜索和过滤
const searchKeyword = ref('')
const filterType = ref('')
const filterTarget = ref('')

// 分页相关
const currentPage = ref(1)
const pageSize = ref(50)
const total = ref(0)

// 计算属性：显示的规则（直接使用后端返回的数据）
const displayRules = computed(() => rules.value)

// 获取规则类型标签颜色
const getTypeTagType = (type) => {
  const typeMap = {
    'DOMAIN': 'primary',
    'DOMAIN-SUFFIX': 'success',
    'DOMAIN-KEYWORD': 'warning',
    'IP-CIDR': 'info',
    'GEOIP': 'primary',
    'MATCH': 'danger'
  }
  return typeMap[type] || ''
}

// 格式化日期时间
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const loadRules = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value
    }

    // 类型过滤
    if (filterType.value) {
      params.type = filterType.value
    }

    // 关键词搜索
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }

    // 目标过滤 - 解析filterTarget
    if (filterTarget.value) {
      if (filterTarget.value.startsWith('node:')) {
        // 具体节点名筛选
        params.target = filterTarget.value.replace('node:', '')
      } else if (filterTarget.value.startsWith('group:')) {
        // 具体代理组名筛选
        params.target = filterTarget.value.replace('group:', '')
      } else {
        // 固定出口 (DIRECT, PROXY, REJECT)
        params.target = filterTarget.value
      }
    }

    const result = await getRules(params)
    rules.value = result.rules || []
    total.value = result.total || 0
  } catch (error) {
    console.error('Load rules error:', error)
    ElMessage.error('加载规则失败')
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  groups.value = await getGroups()
}

const loadNodes = async () => {
  nodes.value = await getNodes()
}

const handleTargetChange = (value) => {
  // Check if it's a node reference (format: "node:ID:Name")
  if (value && value.startsWith('node:')) {
    const parts = value.split(':')
    if (parts.length >= 3) {
      ruleForm.value.TargetID = parseInt(parts[1])
      ruleForm.value.TargetType = 'node'
      ruleForm.value.Target = parts[2] // Store name for display
    }
  } else if (value && value.startsWith('group:')) {
    // Check if it's a group reference (format: "group:ID:Name")
    const parts = value.split(':')
    if (parts.length >= 3) {
      ruleForm.value.TargetID = parseInt(parts[1])
      ruleForm.value.TargetType = 'group'
      ruleForm.value.Target = parts[2] // Store name for display
    }
  } else {
    // Built-in target (PROXY, DIRECT, REJECT)
    ruleForm.value.TargetID = 0
    ruleForm.value.TargetType = ''
    ruleForm.value.Target = value
  }
}

const showCreateDialog = async () => {
  isEdit.value = false
  editId.value = null
  ruleForm.value = {
    Type: 'DOMAIN-SUFFIX',
    Payload: '',
    Target: 'PROXY',
    TargetID: 0,
    TargetType: '',
    Priority: 0,
    NoResolve: false,
    Remark: ''
  }
  await Promise.all([loadGroups(), loadNodes()])
  formDialogVisible.value = true
}

const handleEdit = async (row) => {
  isEdit.value = true
  editId.value = row.ID
  // Build Target value from TargetID and TargetType
  let targetValue = row.Target || 'PROXY'
  if (row.TargetID > 0 && row.TargetType === 'node') {
    targetValue = `node:${row.TargetID}:${row.Target}`
  } else if (row.TargetID > 0 && row.TargetType === 'group') {
    targetValue = `group:${row.TargetID}:${row.Target}`
  }
  ruleForm.value = {
    Type: row.Type,
    Payload: row.Payload,
    Target: targetValue,
    TargetID: row.TargetID || 0,
    TargetType: row.TargetType || '',
    Priority: row.Priority ?? 0,
    NoResolve: row.NoResolve,
    Remark: row.Remark || ''
  }
  await Promise.all([loadGroups(), loadNodes()])
  formDialogVisible.value = true
}

const handleSave = async () => {
  if (!ruleForm.value.Payload || !ruleForm.value.Target) {
    ElMessage.warning('请填写完整信息')
    return
  }
  const data = {
    Type: ruleForm.value.Type,
    Payload: ruleForm.value.Payload,
    Target: ruleForm.value.Target,
    TargetID: ruleForm.value.TargetID || 0,
    TargetType: ruleForm.value.TargetType || '',
    Priority: ruleForm.value.Priority ?? 0,
    NoResolve: ruleForm.value.NoResolve,
    Remark: ruleForm.value.Remark || ''
  }
  if (isEdit.value) {
    await updateRule(editId.value, data)
    ElMessage.success('更新成功')
  } else {
    await createRule(data)
    ElMessage.success('创建成功')
  }
  formDialogVisible.value = false
  loadRules()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该规则吗？', '提示', { type: 'warning' })
  await deleteRule(row.ID)
  ElMessage.success('删除成功')
  loadRules()
}

const handleSearchChange = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTypeChange = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTypeClear = () => {
  filterType.value = ''
  currentPage.value = 1
  loadRules()
}

const handleFilterTargetChange = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTargetClear = () => {
  filterTarget.value = ''
  currentPage.value = 1
  loadRules()
}

const resetFilter = () => {
  searchKeyword.value = ''
  filterType.value = ''
  filterTarget.value = ''
  currentPage.value = 1
  loadRules()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadRules()
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadRules()
}

// 导入相关函数
const showImportDialog = () => {
  importContent.value = ''
  importDialogVisible.value = true
}

const handleFileChange = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    importContent.value = e.target.result
  }
  reader.readAsText(file.raw)
}

const handleFileRemove = () => {
  importContent.value = ''
}

const handleImport = async () => {
  if (!importContent.value) {
    ElMessage.warning('请先选择文件')
    return
  }

  importing.value = true
  try {
    const result = await importRules(importContent.value)
    ElMessage.success(`成功导入 ${result.count} 条规则`)
    importDialogVisible.value = false
    importContent.value = ''
    loadRules()
  } catch (error) {
    ElMessage.error('导入失败: ' + (error.message || '未知错误'))
  } finally {
    importing.value = false
  }
}

onMounted(() => {
  loadRules()
  loadGroups()
  loadNodes()
})
</script>

<style scoped>
.rules-page {
  height: 100%;
}

.rules-card {
  border-radius: 16px;
  border: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  height: 100%;
  display: flex;
  flex-direction: column;
}

:deep(.rules-card .el-card__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
}

:deep(.rules-card .el-card__body) {
  padding: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* 过滤区域 */
.filter-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 24px;
  background: #f9f9f9;
  border-bottom: 1px solid #f0f2f5;
}

.filter-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.search-input {
  width: 280px;
}

.filter-select {
  width: 180px;
}

.filter-select-wide {
  width: 200px;
}

.filter-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.total-count {
  color: #606266;
  font-size: 14px;
  font-weight: 500;
}

/* 下拉选项样式 */
.option-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.option-content-flex {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.option-name {
  flex: 1;
  color: #303133;
  font-size: 14px;
  font-weight: 500;
}

.option-type-tag {
  font-size: 11px;
  padding: 2px 6px;
  height: 18px;
  line-height: 14px;
  background: #f0f2f5;
  border: 1px solid #dcdfe6;
  color: #909399;
}

.option-text {
  color: #606266;
  font-size: 13px;
}

/* 表格样式 */
.rules-table {
  flex: 1;
}

:deep(.rules-table.el-table) {
  border: none;
}

:deep(.rules-table .el-table__header-wrapper) {
  background: #fafafa;
}

:deep(.rules-table .el-table__th) {
  background: #fafafa;
  color: #606266;
  font-weight: 500;
  font-size: 13px;
}

:deep(.rules-table .el-table__body tr:hover > td) {
  background: #f5f7fa;
}

.priority-badge {
  display: inline-block;
  min-width: 24px;
  text-align: center;
  padding: 2px 8px;
  background: #f0f2f5;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

/* 目标列样式 */
.target-cell {
  display: flex;
  align-items: center;
}

.target-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

/* 固定出口标签 */
.target-tag.proxy {
  background: linear-gradient(135deg, #e1f3ff 0%, #d4e9ff 100%);
  color: #409eff;
  border: 1px solid #b3d8ff;
}

.target-tag.direct {
  background: linear-gradient(135deg, #e1f9e8 0%, #d4f1e4 100%);
  color: #67c23a;
  border: 1px solid #b3e19d;
}

.target-tag.reject {
  background: linear-gradient(135deg, #fee 0%, #fde2e2 100%);
  color: #f56c6c;
  border: 1px solid #fbc4c4;
}

/* 其他目标（节点名或代理组名）标签 */
.target-tag.other {
  background: linear-gradient(135deg, #f4f4f5 0%, #e8e8e9 100%);
  color: #606266;
  border: 1px solid #dcdfe6;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
}

.target-tag.other span {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 分页 */
.pagination-section {
  display: flex;
  justify-content: flex-end;
  padding: 16px 24px;
  border-top: 1px solid #f0f2f5;
}

/* 对话框样式 */
.rule-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}

.rule-form .form-input {
  width: 100%;
}

.rule-form .form-hint {
  color: #909399;
  font-size: 12px;
  margin-top: 6px;
  line-height: 1.5;
}

.rule-form :deep(.el-input-group__append .el-select) {
  width: 100%;
}

:deep(.el-select .option-desc) {
  color: #909399;
  font-size: 12px;
  margin-left: 12px;
  float: right;
}

.rule-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

.rule-form :deep(.el-switch) {
  margin-right: 12px;
}

.switch-label {
  color: #909399;
  font-size: 13px;
}

/* 导入对话框 */
.import-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}

.import-form :deep(.el-upload-dragger) {
  padding: 30px;
}

.upload-icon {
  font-size: 48px;
  color: #667eea;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.upload-tip {
  font-size: 12px;
  color: #909399;
}

.import-textarea {
  width: 100%;
}

.import-textarea :deep(.el-textarea__inner) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
}
</style>
