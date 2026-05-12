<template>
  <div class="layout">
    <el-header class="header">
      <div class="header-left">
        <el-button @click="$router.push('/')">返回</el-button>
        <h2>{{ app?.name || '应用详情' }}</h2>
      </div>
    </el-header>

    <el-main class="page-container" v-loading="loading">
      <template v-if="app">
        <el-card style="margin-bottom: 20px;">
          <div class="app-header">
            <el-image
              :src="`/d/${app.id}/icon`"
              style="width: 80px; height: 80px; border-radius: 16px;"
              fit="cover"
            >
              <template #error>
                <div style="width: 80px; height: 80px; background: #eee; border-radius: 16px;" />
              </template>
            </el-image>
            <div class="app-meta">
              <h2>{{ app.name }}</h2>
              <p>Bundle ID: {{ app.bundle_id }}</p>
              <p>
                <span :class="['platform-tag', app.platform === 'ios' ? 'platform-ios' : 'platform-android']">
                  {{ app.platform === 'ios' ? 'iOS' : 'Android' }}
                </span>
                <span v-if="app.is_enterprise" class="platform-tag" style="background: #fff3e0; color: #ff9800; margin-left: 4px;">
                  企业签名
                </span>
              </p>
              <p>安装次数: {{ app.install_count }}</p>
            </div>
            <div class="app-actions">
              <el-button type="primary" @click="showUpload = true">上传新版本</el-button>
              <el-button type="danger" @click="handleDeleteApp">删除应用</el-button>
            </div>
          </div>
        </el-card>

        <el-card>
          <template #header>
            <span>版本历史</span>
          </template>
          <el-table :data="app.versions || []" style="width: 100%">
            <el-table-column prop="version" label="版本号" width="120" />
            <el-table-column prop="build_number" label="Build" width="100" />
            <el-table-column prop="file_size" label="大小" width="100">
              <template #default="{ row }">
                {{ formatSize(row.file_size) }}
              </template>
            </el-table-column>
            <el-table-column prop="min_os_version" label="最低系统" width="100" />
            <el-table-column prop="changelog" label="更新日志" />
            <el-table-column prop="created_at" label="上传时间" width="180">
              <template #default="{ row }">
                {{ new Date(row.created_at).toLocaleString() }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="copyDownloadLink(row)">复制链接</el-button>
                <el-button size="small" type="danger" @click="handleDeleteVersion(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card style="margin-top: 20px;">
          <template #header>
            <span>下载二维码</span>
          </template>
          <div style="text-align: center;">
            <img :src="`/d/${app.id}/v/${app.versions?.[0]?.id}/qrcode`" v-if="app.versions?.length" style="width: 200px;" />
            <p style="margin-top: 10px; color: #666;">扫描二维码下载应用</p>
          </div>
        </el-card>
      </template>
    </el-main>

    <el-dialog v-model="showUpload" title="上传新版本" width="500px">
      <el-upload
        drag
        :auto-upload="false"
        :on-change="handleFileChange"
        accept=".ipa,.apk"
        :limit="1"
      >
        <el-icon size="48"><Upload /></el-icon>
        <div>拖拽文件到此处或<em>点击选择</em></div>
        <template #tip>
          <div style="color: #999;">支持 .ipa 和 .apk 文件</div>
        </template>
      </el-upload>
      <el-input v-model="changelog" type="textarea" placeholder="更新日志（可选）" style="margin-top: 10px;" :rows="3" />
      <el-progress v-if="uploadProgress > 0" :percentage="uploadProgress" style="margin-top: 10px;" />
      <template #footer>
        <el-button @click="showUpload = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { appApi } from '../api'

const route = useRoute()
const router = useRouter()
const app = ref<any>(null)
const loading = ref(false)
const showUpload = ref(false)
const uploading = ref(false)
const uploadFile = ref<File | null>(null)
const changelog = ref('')
const uploadProgress = ref(0)

const loadApp = async () => {
  loading.value = true
  try {
    const { data } = await appApi.get(Number(route.params.id))
    app.value = data.app
  } catch (e) {
    // handled
  } finally {
    loading.value = false
  }
}

const formatSize = (bytes: number) => {
  if (!bytes) return '-'
  const mb = bytes / 1024 / 1024
  return mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
}

const handleFileChange = (file: any) => {
  uploadFile.value = file.raw
}

const handleUpload = async () => {
  if (!uploadFile.value) {
    ElMessage.warning('请选择文件')
    return
  }
  uploading.value = true
  uploadProgress.value = 0
  try {
    await appApi.uploadVersion(
      Number(route.params.id),
      uploadFile.value,
      changelog.value,
      (e) => {
        if (e.total) uploadProgress.value = Math.round((e.loaded / e.total) * 100)
      }
    )
    ElMessage.success('上传成功')
    showUpload.value = false
    uploadFile.value = null
    changelog.value = ''
    uploadProgress.value = 0
    loadApp()
  } catch (e) {
    // handled
  } finally {
    uploading.value = false
  }
}

const copyDownloadLink = (version: any) => {
  const url = `${window.location.origin}/d/${app.value.id}`
  navigator.clipboard.writeText(url)
  ElMessage.success('链接已复制')
}

const handleDeleteVersion = async (version: any) => {
  await ElMessageBox.confirm('确定删除此版本？', '提示')
  try {
    await appApi.deleteVersion(version.id)
    ElMessage.success('已删除')
    loadApp()
  } catch (e) {
    // handled
  }
}

const handleDeleteApp = async () => {
  await ElMessageBox.confirm('确定删除此应用及所有版本？', '警告', { type: 'warning' })
  try {
    await appApi.delete(Number(route.params.id))
    ElMessage.success('已删除')
    router.push('/')
  } catch (e) {
    // handled
  }
}

onMounted(loadApp)
</script>

<style scoped>
.layout { min-height: 100vh; }
.header {
  display: flex; align-items: center;
  background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,0.08); padding: 0 20px;
}
.header-left { display: flex; gap: 12px; align-items: center; }
.header-left h2 { margin: 0; }
.app-header { display: flex; gap: 20px; align-items: flex-start; flex-wrap: wrap; }
.app-meta { flex: 1; }
.app-meta h2 { margin-bottom: 8px; }
.app-meta p { color: #666; margin: 4px 0; }
.app-actions { display: flex; flex-direction: column; gap: 8px; }
</style>
