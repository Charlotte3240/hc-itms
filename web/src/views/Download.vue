<template>
  <div class="download-page">
    <div class="download-card" v-loading="loading">
      <template v-if="app">
        <el-image
          :src="`/d/${app.id}/icon`"
          style="width: 80px; height: 80px; border-radius: 16px;"
          fit="cover"
        >
          <template #error>
            <div style="width: 80px; height: 80px; background: #eee; border-radius: 16px;" />
          </template>
        </el-image>
        <h1>{{ app.name }}</h1>
        <p class="version">v{{ version?.version || '-' }} ({{ version?.build_number || '-' }})</p>
        <p class="size" v-if="version?.file_size">{{ formatSize(version.file_size) }}</p>
        <p class="changelog" v-if="version?.changelog">{{ version.changelog }}</p>

        <div class="install-section">
          <template v-if="app.platform === 'ios'">
            <el-alert v-if="!isSafari" type="warning" :closable="false" style="margin-bottom: 16px;">
              请在 Safari 浏览器中打开此页面以安装应用
            </el-alert>
            <a :href="installUrl" class="install-btn" v-if="version">
              <el-icon><Download /></el-icon> 安装
            </a>
          </template>
          <template v-else>
            <a :href="`/d/${app.id}/v/${version?.id}/apk`" class="install-btn" v-if="version">
              <el-icon><Download /></el-icon> 下载 APK
            </a>
          </template>
        </div>

        <div class="qr-section" v-if="version">
          <img :src="`/d/${app.id}/v/${version.id}/qrcode`" class="qr-code" />
          <p>扫码下载</p>
        </div>
      </template>
      <el-empty v-else-if="!loading" description="应用不存在" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { downloadApi } from '../api'

const route = useRoute()
const app = ref<any>(null)
const version = ref<any>(null)
const loading = ref(false)

const isSafari = /Safari/.test(navigator.userAgent) && !/Chrome/.test(navigator.userAgent)

const installUrl = computed(() => {
  if (!version.value) return '#'
  const plistUrl = `${window.location.origin}/d/${app.value.id}/v/${version.value.id}/plist`
  return `itms-services://?action=download-manifest&url=${encodeURIComponent(plistUrl)}`
})

const formatSize = (bytes: number) => {
  if (!bytes) return '-'
  const mb = bytes / 1024 / 1024
  return mb >= 1 ? `${mb.toFixed(1)} MB` : `${(bytes / 1024).toFixed(0)} KB`
}

const loadData = async () => {
  loading.value = true
  try {
    const { data } = await downloadApi.getLatest(Number(route.params.id))
    app.value = data.app
    version.value = data.version
  } catch (e) {
    // handled
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.download-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}
.download-card {
  background: #fff;
  border-radius: 16px;
  padding: 40px;
  text-align: center;
  max-width: 400px;
  width: 100%;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}
.download-card h1 {
  margin: 16px 0 8px;
  font-size: 24px;
}
.version {
  color: #666;
  font-size: 14px;
}
.size {
  color: #999;
  font-size: 13px;
}
.changelog {
  color: #555;
  font-size: 14px;
  margin: 12px 0;
  padding: 12px;
  background: #f9f9f9;
  border-radius: 8px;
  text-align: left;
}
.install-section {
  margin: 24px 0;
}
.install-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 48px;
  background: #409eff;
  color: #fff;
  border-radius: 8px;
  text-decoration: none;
  font-size: 16px;
  font-weight: 500;
  transition: background 0.2s;
}
.install-btn:hover {
  background: #337ecc;
}
.qr-section {
  margin-top: 24px;
}
.qr-code {
  width: 120px;
  height: 120px;
}
.qr-section p {
  color: #999;
  font-size: 12px;
  margin-top: 8px;
}
</style>
