<template>
  <div class="console">
    <!-- 顶部 -->
    <header class="console__header">
      <div class="console__header-inner">
        <a class="brand" href="/">
          <img src="/logo.png" alt="HP-Lite">
          <span class="brand__text">HP-Lite<span class="brand__sub">内网穿透</span></span>
        </a>

        <div class="console__header-right">
          <div class="user" v-if="!isMobile">
            <t-avatar size="32px" shape="circle" image="/logo.png"/>
            <div class="user__meta">
              <span class="user__email">{{ userInfo.email || '未登录' }}</span>
              <t-tag v-if="userInfo.role === 'ADMIN'" theme="primary" variant="light" size="small">管理员</t-tag>
              <t-tag v-else size="small" variant="light">子用户</t-tag>
            </div>
          </div>

          <t-dropdown
              trigger="click"
              placement="bottom-right"
              :options="userOptions"
              @click="onUserAction"
          >
            <t-button variant="text" shape="square">
              <component :is="isMobile ? PoweroffIcon : ArrowDownIcon"/>
            </t-button>
          </t-dropdown>
        </div>
      </div>
    </header>

    <div class="console__body">
      <!-- 侧边栏 -->
      <aside class="console__aside" :class="{ 'is-collapsed': isCollapsed }">
        <t-menu
            :value="selectedKey"
            theme="light"
            :collapsed="isCollapsed"
            :width="['100%', '100%']"
            :expanded="expanded"
            @change="onMenuChange"
            @expand="onMenuExpand"
        >
          <t-menu-item v-if="userInfo && userInfo.role === 'ADMIN'" value="/client/user">
            <template #icon><user-icon/></template>
            系统用户
          </t-menu-item>

          <t-menu-item value="/client/domain">
            <template #icon><link-icon/></template>
            域名管理
          </t-menu-item>

          <t-menu-item value="/client/device">
            <template #icon><desktop-icon/></template>
            穿透设备
          </t-menu-item>

          <t-menu-item value="/client/config">
            <template #icon><setting-icon/></template>
            穿透配置
          </t-menu-item>

          <t-submenu value="safe" title="穿透安全">
            <template #icon><lock-on-icon/></template>
            <t-menu-item value="/client/safe">
              <template #icon><secured-icon/></template>
              防护规则
            </t-menu-item>
            <t-menu-item value="/client/waf">
              <template #icon><filter-icon/></template>
              穿透限制
            </t-menu-item>
          </t-submenu>

          <t-submenu value="monitor" title="穿透监控">
            <template #icon><chart-bubble-icon/></template>
            <t-menu-item value="/client/monitor">
              <template #icon><chart-line-icon/></template>
              流量统计
            </t-menu-item>
          </t-submenu>

          <t-submenu value="ext" title="扩展功能">
            <template #icon><extension-icon/></template>
            <t-menu-item value="/client/forward">
              <template #icon><arrow-right-icon/></template>
              正向代理
            </t-menu-item>
            <t-menu-item value="/client/reverse">
              <template #icon><arrow-left-icon/></template>
              反向代理
            </t-menu-item>
          </t-submenu>

          <t-menu-item value="/client/teach">
            <template #icon><chat-icon/></template>
            穿透交流
          </t-menu-item>
        </t-menu>

        <div class="console__aside-toggle" @click="toggleCollapse">
          <component :is="isCollapsed ? ArrowRightIcon : ArrowLeftIcon"/>
          <span v-if="!isCollapsed">收起菜单</span>
        </div>
      </aside>

      <!-- 主内容 -->
      <main class="console__main">
        <div class="console__content">
          <router-view/>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import {onMounted, onUnmounted, ref, watch} from "vue";
import {router} from "../../router";
import {NotifyPlugin} from "tdesign-vue-next";
import userInfoStore from "../../data/userInfo";
import {
  ArrowDownIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ChartBubbleIcon,
  ChartLineIcon,
  ChatIcon,
  DesktopIcon,
  ExtensionIcon,
  FilterIcon,
  LinkIcon,
  LockOnIcon,
  PoweroffIcon,
  SecuredIcon,
  SettingIcon,
  UserIcon
} from 'tdesign-icons-vue-next';

const userInfo = ref({});
const selectedKey = ref('');
const isCollapsed = ref(false);
const isMobile = ref(false);
const expanded = ref(['safe', 'monitor', 'ext']);

// 子菜单所属分组，用于路由变化时自动展开
const groupOf = {
  '/client/safe': 'safe',
  '/client/waf': 'safe',
  '/client/monitor': 'monitor',
  '/client/forward': 'ext',
  '/client/reverse': 'ext',
};

const userOptions = [
  {content: '退出登录', value: 'logout', theme: 'error'},
];

onMounted(() => {
  const userData = userInfoStore.getUserInfo();
  if (userData && userData.expTime > Date.now()) {
    userInfo.value = userData;
  } else {
    NotifyPlugin.error({
      title: "校验异常",
      content: "未获取到登录信息，请重新登录",
    });
    router.push("/home/login");
  }

  selectedKey.value = router.currentRoute.value.path;
  syncExpanded(selectedKey.value);
  window.addEventListener('resize', handleResize);
  handleResize();
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
});

const handleResize = () => {
  isMobile.value = window.innerWidth < 768;
  if (isMobile.value) isCollapsed.value = true;
};

const syncExpanded = (path) => {
  const group = groupOf[path];
  if (group && !expanded.value.includes(group)) {
    expanded.value = [...expanded.value, group];
  }
};

const handleLoginOut = () => {
  userInfoStore.removeUserInfo();
  NotifyPlugin.success({
    title: "退出成功",
    content: "您已安全退出系统",
  });
  router.push("/");
};

const onUserAction = (item) => {
  if (item.value === 'logout') {
    handleLoginOut();
  }
};

const onMenuChange = (value) => {
  selectedKey.value = value;
  router.push(value);
};

const onMenuExpand = (value) => {
  expanded.value = value;
};

const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value;
};

watch(
    () => router.currentRoute.value.path,
    (newPath) => {
      selectedKey.value = newPath;
      syncExpanded(newPath);
    }
);
</script>

<style scoped>
.console {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--hp-bg);
}

/* ---------- 头部 ---------- */
.console__header {
  position: sticky;
  top: 0;
  z-index: 60;
  height: 60px;
  flex: none;
  background: rgba(255, 255, 255, .9);
  backdrop-filter: saturate(180%) blur(12px);
  -webkit-backdrop-filter: saturate(180%) blur(12px);
  border-bottom: 1px solid var(--hp-border);
}

.console__header-inner {
  height: 100%;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand img {
  width: 34px;
  height: 34px;
}

.brand__text {
  font-size: 17px;
  font-weight: 700;
  color: var(--hp-text);
  letter-spacing: .2px;
}

.brand__sub {
  margin-left: 8px;
  font-size: 12px;
  font-weight: 500;
  color: var(--hp-text-3);
}

.console__header-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 12px 4px 4px;
  border-radius: 999px;
  background: #f6f8fc;
}

.user__meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user__email {
  font-size: 13px;
  color: var(--hp-text);
  max-width: 190px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* ---------- 主体 ---------- */
.console__body {
  flex: 1;
  display: flex;
  min-height: 0;
}

.console__aside {
  flex: none;
  width: 232px;
  display: flex;
  flex-direction: column;
  background: #fff;
  border-right: 1px solid var(--hp-border);
  height: calc(100vh - 60px);
  position: sticky;
  top: 60px;
  transition: width .25s cubic-bezier(.4, 0, .2, 1);
  overflow: hidden;
}

.console__aside.is-collapsed {
  width: 64px;
}

.console__aside :deep(.t-menu) {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding-top: 12px;
  border-right: none;
}

.console__aside :deep(.t-menu__item) {
  margin: 3px 10px;
  border-radius: 10px;
  height: 40px;
  transition: background .2s ease, color .2s ease;
}

.console__aside :deep(.t-menu__item:hover) {
  background: #f3f6fd;
}

.console__aside :deep(.t-menu__item.t-is-active) {
  background: var(--td-brand-color-1);
  color: var(--td-brand-color);
  font-weight: 600;
}

.console__aside :deep(.t-menu__sub) {
  margin: 3px 10px;
}

.console__aside :deep(.t-menu__sub .t-menu__item) {
  margin: 3px 0;
}

.console__aside-toggle {
  flex: none;
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-top: 1px solid var(--hp-border);
  color: var(--hp-text-3);
  font-size: 13px;
  cursor: pointer;
  transition: color .2s ease, background .2s ease;
}

.console__aside-toggle:hover {
  color: var(--td-brand-color);
  background: #f8fafd;
}

/* ---------- 内容 ---------- */
.console__main {
  flex: 1;
  min-width: 0;
  height: calc(100vh - 60px);
  overflow-y: auto;
  padding: 18px;
}

.console__content {
  background: #fff;
  border: 1px solid var(--hp-border);
  border-radius: var(--hp-radius);
  box-shadow: var(--hp-shadow-sm);
  padding: 22px;
  min-height: 100%;
}

@media (max-width: 768px) {
  .console__main {
    padding: 10px;
  }

  .console__content {
    padding: 14px;
  }

  .brand__sub {
    display: none;
  }
}
</style>
