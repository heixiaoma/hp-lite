<template>
  <a-layout id="layout">
    <!-- 头部导航 -->
    <a-layout-header class="header">
      <div class="logo"><a href="/" style="color: #ffffff"> <img src="/logo-back.png" style="width:45px">HP-Lite内网穿透</a>
      </div>
      <div class="user">
        <a-dropdown trigger="click">
          <span class="name">{{ userInfo.email }}</span>
          <template #overlay>
            <a-menu slot="overlay" @click="handleMenuClick">
              <a-menu-item key="logout">退出登录</a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
        <span style="margin-left: 10px" class="name">{{ userInfo.auth === "NO_AUTH" ? "未认证" : "已认证" }}</span>
      </div>

      <div class="user-mini">
        <span @click="handleLoginOut">退出</span>
      </div>


    </a-layout-header>
    <!-- 左右布局菜单页面系统 -->
    <a-layout>
      <a-layout-sider style="background-color: #4b6ff6" :width="130" :collapsed-width="0" breakpoint="md" collapsible
                      :collapsed="isCollapsed"
                      @collapse="collapsedTrigger"
                      class="sider">
        <a-menu style="background-color: #4b6ff6" :theme="theme" mode="inline" :default-selected-keys="[selectedKey]"
                @select="handleMenuSelect">
          <a-menu-item key="user" v-if="userInfo&&userInfo.role==='ADMIN'">
            <router-link to="/client/user">穿透用户</router-link>
          </a-menu-item>
          <a-menu-item key="device">
            <router-link to="/client/device">穿透设备</router-link>
          </a-menu-item>
          <a-menu-item key="domain">
            <router-link to="/client/domain">穿透域名</router-link>
          </a-menu-item>
          <a-menu-item key="config">
            <router-link to="/client/config">穿透配置</router-link>
          </a-menu-item>
          <a-menu-item key="monitor">
            <router-link to="/client/monitor">穿透监控</router-link>
          </a-menu-item>
          <a-menu-item key="teach">
            <router-link to="/client/teach">穿透教程</router-link>
          </a-menu-item>
        </a-menu>
      </a-layout-sider>
      <a-layout-content class="content">
        <router-view></router-view>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script>
import userInfo from "../../data/userInfo";
import {router} from "../../router";
import {notification} from "ant-design-vue";

export default {
  data() {
    return {
      userInfo: {},
      theme: 'dark',
      selectedKey: this.$route.path,
      username: 'John Doe',
      isCollapsed: false
    }
  },
  mounted() {
    var userInfo1 = userInfo.getUserInfo();
    if (userInfo1 && userInfo1.expTime > Date.parse(new Date())) {
      this.userInfo = userInfo1
    } else {
      //校验失败前往重新登录
      notification.open({
        message: "校验异常",
        description: "未获取到登录信息，请重新登录"
      })
      router.push("/home/login")
    }
  },
  methods: {
    handleLoginOut() {
      // 执行退出登录逻辑
      userInfo.removeUserInfo()
      notification.open({
        message: "退出成功",
      })
      router.push("/")
    },
    handleMenuClick({key}) {
      if (key === 'logout') {
        this.handleLoginOut()
      }
    },
    handleMenuSelect({key}) {
      this.selectedKey = key;
    },
    collapsedTrigger(e) {
      this.isCollapsed = !this.isCollapsed;
    }
  },
  watch: {
    $route() {
      this.selectedKey = this.$route.path;
      // this.isCollapsed = false;
    }
  }
}
</script>

<style scoped>
.header {
  background-color: #4b6ff6;
  display: flex;
  align-items: center;
}

.logo {
  font-size: 24px;
  color: #ffffff;
  margin-right: 50px;
}

.menu {
  flex-grow: 1;
}

.user {
  margin-left: auto;
  text-align: right;
}

.user-mini {
  display: none;
}

.name {
  color: #ffffff;
  line-height: 44px;
  cursor: pointer;
}

.sider {
  min-height: calc(100vh - 64px);
}

:deep(.ant-layout-sider-zero-width-trigger) {
  background-color: #4b6ff6;
}

.content {
  padding: 24px;
}

@media only screen and (max-width: 767px) {
  .logo {
    margin-right: 0;
  }

  .user {
    display: none;
  }

  .user-mini {
    margin-left: auto;
    text-align: right;
    color: #ffffff;
    cursor: pointer;
    display: block;
  }
}

</style>
