<template>
  <div class="layout">
    <el-header class="header">
      <div class="header-left">
        <el-button @click="$router.push('/')">返回</el-button>
        <h2>上传应用</h2>
      </div>
    </el-header>

    <el-main class="page-container">
      <el-card>
        <el-upload
          ref="uploadRef"
          drag
          :auto-upload="false"
          :on-change="handleFileChange"
          :on-exceed="handleExceed"
          accept=".ipa,.apk"
          :limit="1"
          style="margin-bottom: 20px;"
        >
          <el-icon size="64"><Upload /></el-icon>
          <div style="font-size: 16px;">拖拽 IPA/APK 文件到此处</div>
          <div style="color: #999;">或点击选择文件</div>
          <template #tip>
            <div v-if="file" style="margin-top: 8px; color: #409eff;">
              已选择: {{ file.name }}
              <el-button type="danger" text size="small" @click.stop="clearFile">移除</el-button>
            </div>
          </template>
        </el-upload>

        <el-input v-model="changelog" type="textarea" placeholder="更新日志（可选）" :rows="4" />

        <el-progress v-if="progress > 0" :percentage="progress" style="margin: 20px 0;" />

        <div style="text-align: center; margin-top: 20px;">
          <el-button type="primary" size="large" :loading="uploading" :disabled="!file" @click="handleUpload">
            上传并解析
          </el-button>
        </div>
      </el-card>

      <el-card v-if="result" style="margin-top: 20px;">
        <template #header>
          <span>上传成功</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="应用名称">{{ result.app.name }}</el-descriptions-item>
          <el-descriptions-item label="平台">{{ result.app.platform === 'ios' ? 'iOS' : 'Android' }}</el-descriptions-item>
          <el-descriptions-item label="Bundle ID">{{ result.app.bundle_id }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ result.version.version }}</el-descriptions-item>
          <el-descriptions-item label="Build">{{ result.version.build_number }}</el-descriptions-item>
          <el-descriptions-item label="大小">{{ formatSize(result.version.file_size) }}</el-descriptions-item>
        </el-descriptions>
        <div style="text-align: center; margin-top: 20px;">
          <el-button type="primary" @click="$router.push(`/app/${result.app.id}`)">查看应用详情</el-button>
          <el-button @click="$router.push('/')">返回首页</el-button>
        </div>
      </el-card>
    </el-main>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, type UploadFile, type UploadInstance } from 'element-plus'
import { appApi } from '../api'

const uploadRef = ref<UploadInstance>()
const file = ref<File | null>(null)
const changelog = ref('')
const uploading = ref(false)
const progress = ref(0)
const result = ref<any>(null)

const handleFileChange = (f: UploadFile) => {
  file.value = f.raw ?? null
  result.value = null
}

const handleExceed = () => {
  ElMessage.warning('只能选择一个文件，请先移除已选文件')
}

const clearFile = () => {
  file.value = null
  uploadRef.value?.clearFiles()
}

const formatSize = (bytes: number) => {
  if (!bytes) return '-'
  const mb = bytes / 1024 / 1024
  return mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
}

const handleUpload = async () => {
  if (!file.value) {
    ElMessage.warning('请选择文件')
    return
  }

  // We need an app ID. For new uploads, we'll create app ID 0 and let the backend handle it.
  // Actually, the backend expects an app ID in the URL. Let's upload to app ID 0 for new apps.
  // The backend will create a new app if bundle_id doesn't exist.
  uploading.value = true
  progress.value = 0
  try {
    const { data } = await appApi.uploadVersion(
      0,
      file.value,
      changelog.value,
      (e) => {
        if (e.total) progress.value = Math.round((e.loaded / e.total) * 100)
      }
    )
    result.value = data
    file.value = null
    uploadRef.value?.clearFiles()
    ElMessage.success('上传成功')
  } catch (e) {
    // handled
  } finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.layout { min-height: 100vh; }
.header {
  display: flex; align-items: center;
  background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,0.08); padding: 0 20px;
}
.header-left { display: flex; gap: 12px; align-items: center; }
.header-left h2 { margin: 0; }
</style>
