<template>
  <div class="layout">
    <el-header class="header">
      <div class="header-left">
        <h2>HC-ITMS</h2>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="$router.push('/upload')">
          <el-icon><Upload /></el-icon> 上传应用
        </el-button>
        <el-button @click="logout">退出</el-button>
      </div>
    </el-header>

    <el-main class="main-content">
      <div class="filter-bar">
        <el-radio-group v-model="platform" @change="loadApps">
          <el-radio-button label="">全部</el-radio-button>
          <el-radio-button label="ios">iOS</el-radio-button>
          <el-radio-button label="android">Android</el-radio-button>
        </el-radio-group>
      </div>

      <div class="app-grid" v-loading="loading">
        <div
          class="app-card"
          v-for="app in apps"
          :key="app.id"
          @click="$router.push(`/app/${app.id}`)"
        >
          <div class="card-icon">
            <el-image
              :src="`/d/${app.id}/icon?t=${app.updated_at}`"
              fit="cover"
              style="width: 56px; height: 56px;"
            >
              <template #error>
                <div class="icon-fallback">
                  <el-icon :size="24"><Cellphone /></el-icon>
                </div>
              </template>
            </el-image>
          </div>
          <div class="card-body">
            <h3 class="card-title">{{ app.name }}</h3>
            <div class="card-tags">
              <span :class="['tag', 'tag-' + app.platform]">
                {{ app.platform === 'ios' ? 'iOS' : 'Android' }}
              </span>
              <span v-if="app.is_enterprise" class="tag tag-enterprise">企业版</span>
            </div>
            <div class="card-footer">
              <span class="version">v{{ app.latest_version || '-' }}</span>
              <span class="installs">
                <el-icon><Download /></el-icon>
                {{ app.install_count }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <el-empty v-if="!loading && apps.length === 0" description="暂无应用，点击右上角上传" />
    </el-main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { appApi } from '../api'

const router = useRouter()
const apps = ref<any[]>([])
const loading = ref(false)
const platform = ref('')

const loadApps = async () => {
  loading.value = true
  try {
    const { data } = await appApi.list(platform.value)
    apps.value = data.apps || []
  } catch (e) {
    // handled
  } finally {
    loading.value = false
  }
}

const logout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}

onMounted(loadApps)
</script>

<style scoped>
.layout {
  min-height: 100vh;
  background: #f0f2f5;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
  padding: 0 24px;
  height: 60px;
}

.header-left h2 {
  color: #1a1a1a;
  font-size: 20px;
  font-weight: 600;
}

.header-right {
  display: flex;
  gap: 12px;
}

.main-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
}

.filter-bar {
  margin-bottom: 24px;
}

/* Responsive grid */
.app-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}

/* Card styles */
.app-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.app-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.card-icon {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  border-radius: 12px;
  overflow: hidden;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-icon img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.icon-fallback {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f0f0;
  color: #ccc;
}

.card-body {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  margin: 0 0 8px 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-tags {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;
}

.tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.tag-ios {
  background: #e8f4fd;
  color: #1890ff;
}

.tag-android {
  background: #e8f8e8;
  color: #52c41a;
}

.tag-enterprise {
  background: #fff3e0;
  color: #ff9800;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.version {
  font-size: 13px;
  color: #666;
  font-weight: 500;
}

.installs {
  font-size: 12px;
  color: #999;
  display: flex;
  align-items: center;
  gap: 4px;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .main-content {
    padding: 16px;
  }

  .app-grid {
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 12px;
  }

  .app-card {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .app-grid {
    grid-template-columns: 1fr;
  }
}
</style>
