<template>
  <a-layout id="layout" class="layout-container">
    <!-- 头部导航 -->
    <a-layout-header class="header">
      <!-- 头部内容保持不变 -->
      <div class="header-content">
        <div class="logo">
          <a href="/" class="logo-link">
            <img src="/logo-back.png" class="logo-img" alt="HP-Lite">
            <span class="logo-text">HP-Lite内网穿透</span>
          </a>
        </div>

        <div class="user-area">
          <a-dropdown
              trigger="click"
              :overlay-style="{ borderRadius: '8px', boxShadow: '0 4px 16px rgba(0,0,0,0.15)' }"
          >
            <div class="user-info" @click.stop>
              <a-avatar class="user-avatar">
                <template #icon>
                  <img src="/logo.png" >
                </template>
              </a-avatar>
              <span class="user-email">{{ userInfo.email || '未登录' }}</span>
            </div>
            <template #overlay>
              <a-menu @click="handleMenuClick" class="user-menu">
<!--                <a-menu-item key="profile" class="menu-item">-->
<!--                  <a-icon type="user" class="menu-icon" />-->
<!--                  <span>个人资料</span>-->
<!--                </a-menu-item>-->
<!--                <a-menu-item key="setting" class="menu-item">-->
<!--                  <a-icon type="setting" class="menu-icon" />-->
<!--                  <span>系统设置</span>-->
<!--                </a-menu-item>-->
<!--                <a-menu-divider />-->
                <a-menu-item key="logout" class="menu-item logout-item">
                  <span>退出登录</span>
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>

          <div class="user-mini" @click="handleLoginOut">
            <a-button type="text" class="mini-logout">
              <span>退出登录</span>
            </a-button>
          </div>
        </div>
      </div>
    </a-layout-header>

    <!-- 主体内容区域 - 关键优化区域 -->
    <a-layout class="main-container">
      <!-- 侧边栏保持固定 -->
      <a-layout-sider
          class="sidebar"
          :width="160"
          :collapsed-width="64"
          breakpoint="md"
          collapsible
          :collapsed="isCollapsed"
          @collapse="collapsedTrigger"
          @expand="collapsedTrigger"
          :style="{
          backgroundColor: '#4b6ff6',
          transition: 'all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1)'
        }"
      >
        <!-- 侧边栏菜单保持不变 -->
        <a-menu
            mode="inline"
            :theme="theme"
            :selected-keys="[selectedKey]"
            @select="handleMenuSelect"
            class="sidebar-menu"
        >
          <a-menu-item key="/client/user" v-if="userInfo&&userInfo.role==='ADMIN'" class="sidebar-item">
            <template #icon><svg t="1751867825889" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="6970" width="24" height="24"><path d="M592.896 589.141333c96.938667-34.816 166.570667-131.072 166.570667-244.053333 0-142.677333-110.933333-258.048-247.466667-258.048-136.533333 0-247.466667 115.712-247.466667 258.048 0 112.981333 69.632 208.896 166.570667 244.053333-162.816 40.96-284.672 202.069333-284.672 383.658667L877.226667 972.8C877.226667 791.210667 755.712 630.101333 592.896 589.141333L592.896 589.141333zM592.896 589.141333" fill="#ffffff" p-id="6971"></path></svg></template>
            <span>穿透用户</span>
          </a-menu-item>
          <a-menu-item key="/client/device" class="sidebar-item">
            <template #icon><svg t="1751867998833" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="9921" width="24" height="24"><path d="M96 896a32 32 0 0 1-32-32V592a32 32 0 0 1 32-32h832a32 32 0 0 1 32 32v272a32 32 0 0 1-32 32z m48.064-176.064a47.936 47.936 0 0 0 48 48h160a48 48 0 0 0 0-96h-160a47.936 47.936 0 0 0-48.064 48z m-48.064-256a32 32 0 0 1-32-32V160a32 32 0 0 1 32-32h832a32 32 0 0 1 32 32v271.936a32 32 0 0 1-32 32z m48.064-175.872a48.064 48.064 0 0 0 48 48h160a48 48 0 0 0 0-96h-160a48 48 0 0 0-48.064 47.936z" fill="#ffffff" p-id="9922"></path></svg></template>
            <span>穿透设备</span>
          </a-menu-item>
          <a-menu-item key="/client/domain" class="sidebar-item">
            <template #icon><svg t="1751868116673" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="12795" width="24" height="24"><path d="M178.581333 682.666667c62.848 126.442667 193.322667 213.333333 344.085334 213.333333 150.762667 0 281.237333-86.890667 344.085333-213.333333h93.226667C891.733333 857.514667 721.664 981.333333 522.666667 981.333333 323.690667 981.333333 153.6 857.493333 85.333333 682.666667h93.248zM130.730667 394.666667l29.482666 104.32 31.978667-104.32h64.298667l30.848 103.722666 28.714666-103.722666h97.685334l29.568 104.384 32-104.384h64.277333l30.848 103.722666 28.736-103.722666h111.146667l29.546666 104.384 32-104.384h64.298667l30.848 103.722666 28.714667-103.722666h78.805333l-73.514667 234.666666h-68.416l-30.037333-100.629333-32.405333 100.629333h-66.922667l-49.514667-156.949333-49.28 156.949333h-68.416l-30.016-100.629333-32.426666 100.629333h-66.901334l-42.837333-135.594666-42.496 135.594666h-68.416l-30.037333-100.629333L190.506667 629.333333H123.562667l-74.112-234.666666h81.28zM522.666667 42.666667C721.664 42.666667 891.733333 166.506667 960 341.333333H866.773333C803.925333 214.912 673.450667 128 522.666667 128c-150.784 0-281.258667 86.890667-344.085334 213.333333H85.333333C153.6 166.506667 323.669333 42.666667 522.666667 42.666667z" fill="#ffffff" p-id="12796"></path></svg></template>
            <span>穿透域名</span>
          </a-menu-item>
          <a-menu-item key="/client/config" class="sidebar-item">
            <template #icon><svg t="1751868156496" class="icon" viewBox="0 0 1034 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="13941" width="24" height="24"><path d="M734.208 196.608c97.28 0 186.368 33.792 258.048 89.088V178.176c0-38.912-37.888-68.608-83.968-68.608H83.968c-47.104 0-83.968 29.696-83.968 68.608v353.28h324.608c38.912-190.464 207.872-334.848 409.6-334.848zM390.144 851.968H75.776c-20.48 0-37.888 14.336-37.888 30.72s17.408 30.72 37.888 30.72h366.592c-18.432-18.432-36.864-39.936-52.224-61.44zM317.44 593.92h-317.44v34.816c0 38.912 37.888 69.632 83.968 69.632h240.64c-5.12-26.624-8.192-55.296-8.192-83.968 0-6.144 1.024-13.312 1.024-20.48z" fill="#ffffff" p-id="13942"></path><path d="M980.992 614.4c0-34.816 21.504-64.512 53.248-76.8-7.168-28.672-18.432-56.32-33.792-80.896-30.72 13.312-66.56 8.192-91.136-16.384-24.576-24.576-30.72-61.44-16.384-91.136-24.576-14.336-52.224-26.624-80.896-33.792-12.288 30.72-41.984 53.248-76.8 53.248-34.816 0-64.512-21.504-76.8-53.248-28.672 7.168-56.32 18.432-80.896 33.792 13.312 30.72 8.192 66.56-16.384 91.136-24.576 24.576-61.44 30.72-91.136 16.384-15.36 24.576-26.624 52.224-33.792 80.896 30.72 12.288 53.248 41.984 53.248 76.8 0 34.816-21.504 64.512-53.248 76.8 7.168 28.672 18.432 56.32 33.792 80.896 30.72-13.312 66.56-8.192 91.136 16.384 24.576 24.576 30.72 61.44 16.384 91.136 24.576 14.336 52.224 26.624 80.896 33.792 12.288-30.72 41.984-53.248 76.8-53.248 34.816 0 64.512 21.504 76.8 53.248 28.672-7.168 56.32-18.432 80.896-33.792-13.312-30.72-8.192-66.56 16.384-91.136 24.576-24.576 61.44-30.72 91.136-16.384 15.36-24.576 26.624-52.224 33.792-80.896-30.72-12.288-53.248-41.984-53.248-76.8zM734.208 759.808C654.336 759.808 588.8 694.272 588.8 614.4c0-79.872 65.536-145.408 145.408-145.408 79.872 0 145.408 65.536 145.408 145.408 0 80.896-64.512 145.408-145.408 145.408z" fill="#ffffff" p-id="13943"></path><path d="M734.208 614.4m-74.752 0a74.752 74.752 0 1 0 149.504 0 74.752 74.752 0 1 0-149.504 0Z" fill="#ffffff" p-id="13944"></path></svg></template>
            <span>穿透配置</span>
          </a-menu-item>
          <a-menu-item key="/client/waf" class="sidebar-item">
            <template #icon><svg t="1751891575635" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="6191" width="24" height="24"><path d="M1002.55552 200.832c-21.888-126.08-52.16-156.096-179.904-178.432C741.30752 8.128 626.42752 0.064 507.70752 0.064c-117.76 0-226.688 7.68-306.688 21.568C74.93952 43.584 44.98752 73.792 22.65152 201.536c-29.632 169.152-30.016 453.824-1.024 621.632 21.888 126.08 52.096 156.032 179.904 178.432 81.536 14.272 196.544 22.464 315.712 22.464 117.44 0 226.176-7.616 306.112-21.44 126.08-21.888 156.096-52.16 178.368-179.968 29.568-169.216 30.016-453.888 0.896-621.696z m-31.936 616.32c-20.032 114.56-40 134.72-152.832 154.368-78.144 13.504-184.896 20.992-300.608 20.992-117.312 0-230.4-8-310.272-21.952-114.56-20.096-134.784-40.128-154.304-152.768-28.544-164.864-28.16-444.608 0.96-610.816 20.032-114.496 40-134.656 152.768-154.24C284.60352 39.104 391.61152 31.616 507.57952 31.616c117.056 0 229.888 8 309.504 21.952 114.496 20.096 134.656 40.064 154.304 152.832 28.608 164.864 28.224 444.608-0.896 610.944z" fill="#ffffff" p-id="6192"></path><path d="M513.01952 192h-2.368v0.32L229.37152 266.88l16.448 297.856c0 1.984 0 39.616 12.032 62.208 60.544 113.728 203.968 182.848 249.92 203.328a30.592 30.592 0 0 0 2.88 1.152V832c0.32 0 0.704-0.128 1.088-0.256l1.28 0.256h0.064v-0.704c1.152-0.512 2.176-0.96 2.24-1.088 45.952-20.48 189.376-89.536 249.92-203.328 12.032-22.656 13.248-60.224 13.312-62.208l16.512-297.856-281.92-74.88zM327.09952 330.368c0-8.128 6.592-14.656 14.656-14.656h139.584c8.128 0 14.656 6.592 14.656 14.656v66.112a14.656 14.656 0 0 1-14.656 14.656H341.75552a14.656 14.656 0 0 1-14.656-14.656V330.368z m0 124.928c0-8.128 6.592-14.656 14.656-14.656h44.096c8.128 0 14.656 6.592 14.656 14.656v66.112a14.656 14.656 0 0 1-14.656 14.656h-44.096a14.656 14.656 0 0 1-14.656-14.656V455.296z m169.024 190.976a14.656 14.656 0 0 1-14.656 14.656H341.88352a14.656 14.656 0 0 1-14.656-14.656V580.16c0-8.128 6.592-14.656 14.656-14.656h139.584c8.128 0 14.656 6.592 14.656 14.656v66.112z m198.336 0a14.656 14.656 0 0 1-14.656 14.656H540.21952a14.656 14.656 0 0 1-14.656-14.656V580.16c0-8.128 6.592-14.656 14.656-14.656h139.584c8.128 0 14.656 6.592 14.656 14.656v66.112z m0-124.864a14.656 14.656 0 0 1-14.656 14.656h-44.096a14.592 14.592 0 0 1-14.656-14.656V455.296c0-8.128 6.592-14.656 14.656-14.656h44.096c8.128 0 14.656 6.592 14.656 14.656v66.112z m-168.96-191.04c0-8.128 6.592-14.656 14.656-14.656h139.584c8.128 0 14.656 6.592 14.656 14.656v66.112a14.656 14.656 0 0 1-14.656 14.656H540.15552a14.656 14.656 0 0 1-14.656-14.656V330.368z m51.392 110.272c8.128 0 14.656 6.592 14.656 14.656v66.112a14.592 14.592 0 0 1-14.656 14.656H444.66752a14.592 14.592 0 0 1-14.656-14.656V455.296c0-8.128 6.592-14.656 14.656-14.656h132.224z" fill="#ffffff" p-id="6193"></path></svg></template>
            <span>穿透安全</span>
          </a-menu-item>
          <a-menu-item key="/client/monitor" class="sidebar-item">
            <template #icon><svg t="1751868202326" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="15169" width="24" height="24"><path d="M764.74 955.37c-5.12 0.38-9.22 0.51-13.18 0.51-161.2 0.12-322.65 0.12-483.84 0.12-13.7 0-13.83-0.13-13.7-13.83v-34.05c0-9.6 3.97-13.57 13.83-13.57 33.67 0 67.47 0.12 101.15 0 14.08-0.13 15.37-1.67 15.37-15.23v-34.06c-0.13-12.68-3.2-15.75-15.75-15.75H60.56c-26.76 0-45.84-11.91-55.44-37.13-2.69-6.91-4.74-14.47-4.74-21.64-0.26-214.97 0-429.93-0.38-644.77 0-24.96 13.44-40.84 32.13-53.13 6.4-4.23 15.62-5.12 23.56-6.15 7.29-1.03 15.1-0.38 22.79-0.38H955.12c21 0 39.56 6.01 52.49 22.66 8.58 11.01 15.75 22.02 15.62 38.53-0.77 213.68-0.51 427.5-0.51 641.18 0 10.5 0 19.84-8.96 28.93-6.28 6.53-9.99 16.65-18.82 21.25-8.2 4.35-17.03 9.86-25.73 10.11-53.65 1.03-107.42 0.51-161.07 0.51-52.37 0.13-104.86 0-157.23 0-12.93 0-13.32 0.38-13.44 13.57-0.13 13.44 0.51 26.89-0.26 40.45-0.64 9.86 4.74 11.27 12.29 11.27 33.29-0.13 66.32-0.13 99.36-0.13 14.47 0 15.87 1.54 15.75 15.62 0.13 14.63 0.13 29.23 0.13 45.11zM63.51 128.28c-0.38 6.53-1.03 11.65-1.03 16.9-0.12 45.71 0 91.29 0 137.12v341.33c0 17.03 0 16.9 17.03 16.9h863.58c13.7 0 16.64-2.69 16.64-16.52V142.74c0-13.7-1.03-14.6-14.72-14.6H63.51v0.14zM436.46 370c-13.32 1.28-25.22-1.66-36.1-9.86-1.66-1.15-6.4 0.13-8.58 1.79-24.32 19.2-48.27 38.79-72.59 58-15.37 12.16-31.11 23.82-46.48 35.85-6.01 4.86-11.9 9.98-4.73 18.43 1.02 1.15 1.28 3.71 0.77 5.25-5.9 19.46-11.27 38.79-32.26 48.52-26.76 12.3-51.86 4.99-70.04-17.02-10.75-13.19-13.18-28.68-10.11-44.17 3.59-17.67 13.83-32.26 31.24-39.05 9.35-3.59 19.97-3.97 29.96-5.76 2.04-0.38 4.73-0.26 6.27 0.77 14.72 9.35 25.74 4.1 37.64-6.79 13.83-12.8 29.71-23.56 44.69-35.21 22.15-17.15 44.42-34.18 66.06-51.85 3.59-2.81 6.14-8.83 6.02-13.44-0.51-20.74 9.35-37.13 25.22-47.24 14.98-9.6 33.29-12.04 51.34-3.97 9.86 4.23 18.05 9.73 24.07 18.56 11.27 16.77 14.47 33.93 6.02 53.26-1.66 3.58 2.43 10.88 5.76 14.85 13.7 16.9 29.19 32.13 41.23 49.8 6.15 9.34 12.55 7.94 20.49 8.45 10.37 0.64 21.25-3.07 30.59 5.12 1.03 1.16 6.66-0.77 8.97-2.69 16-14.72 31.88-29.58 47.76-44.55 26.89-25.09 53.78-50.32 80.91-75.41 9.22-8.45 18.95-16.26 27.66-25.22 3.07-3.2 5.12-8.32 5.76-12.93 0.77-7.17-1.28-14.72-0.39-22.02 2.18-16.91 8.32-32.26 24.58-40.46 8.84-4.48 18.44-10.12 27.79-9.98 16 0.26 31.88 5.38 43.65 17.54 16.26 16.91 15.62 38.41 11.4 58.52-3.2 14.85-15.62 25.99-30.86 31.11-8.32 2.82-17.67 3.71-26.5 5.13-1.66 0.38-4.09 0.38-5.38-0.51-15.88-11.02-24.07 1.41-33.8 10.5-15.49 14.85-31.5 29.06-47.38 43.79-22.91 21.38-45.83 43.02-68.75 64.4-11.27 10.63-22.79 21.12-33.55 32.26-2.69 2.69-4.74 8.06-4.23 11.78 2.56 23.81-6.53 43.4-26.38 55.05-12.93 7.56-28.42 11.78-45.19 4.48-11.52-5.12-22.41-9.73-29.06-20.61-10.12-16-14.72-32.78-6.02-51.34 2.82-5.76 3.33-11.9-2.3-18.3-14.6-16.65-27.79-34.45-41.61-51.47-1.79-2.18-5.25-3.84-8.2-3.97-4.73-0.41-10.11 0.49-15.36 0.61zM229.43 477.16c-0.12-10.88-6.27-16.26-18.18-16.26-10.88 0.12-17.41 6.66-17.15 17.54 0.12 10.75 7.17 17.79 17.54 17.79 9.09-0.12 17.92-9.46 17.79-19.07z m580.12-256.95c-11.52-0.13-17.54 5.5-17.67 16.9-0.13 10.88 6.66 18.44 16.64 18.44 8.58 0 18.82-9.47 18.82-17.8 0.13-9.61-7.93-17.42-17.79-17.54zM539.4 462.19c-0.26 12.55 4.1 17.67 15.37 17.79 11.27 0.13 19.46-6.91 19.59-16.9 0.13-6.01-9.35-18.82-18.05-18.44-10.63 0.26-16.66 5.89-16.91 17.55z m-86.8-146.21c-3.97-13.7-11.27-21-26.25-16.26-8.19 2.56-8.45 10.12-10.5 15.75-1.79 5.5 7.94 16 16.26 17.28 12.8 1.92 16.64-8.33 20.49-16.77z m0 0" fill="#ffffff" p-id="15170"></path></svg></template>
            <span>穿透监控</span>
          </a-menu-item>
          <a-menu-item key="/client/teach" class="sidebar-item">
            <template #icon><svg t="1751868316938" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="22356" width="24" height="24"><path d="M896 128l0 832-672 0c-53.024 0-96-42.976-96-96s42.976-96 96-96l608 0 0-768-640 0c-70.4 0-128 57.6-128 128l0 768c0 70.4 57.6 128 128 128l768 0 0-896-64 0z" fill="#ffffff" p-id="22357"></path><path d="M224.064 832l0 0c-0.032 0-0.032 0-0.064 0-17.664 0-32 14.336-32 32s14.336 32 32 32c0.032 0 0.032 0 0.064 0l0 0 607.904 0 0-64-607.904 0z" fill="#ffffff" p-id="22358"></path></svg></template>
            <span>穿透说明</span>
          </a-menu-item>
        </a-menu>
      </a-layout-sider>

      <!-- 主内容区域 - 核心优化点 -->
      <div class="content-wrapper">
        <a-layout-content class="main-content">
          <div class="content-container">
            <router-view />
          </div>
        </a-layout-content>
      </div>
    </a-layout>
  </a-layout>
</template>

<script>
// 脚本部分保持不变
import userInfo from "../../data/userInfo";
import { router } from "../../router";
import { notification } from "ant-design-vue";
import { onMounted, reactive, toRefs, watch } from "vue";

export default {
  setup() {
    const state = reactive({
      userInfo: {},
      theme: 'dark',
      selectedKey: '',
      isCollapsed: false,
      isMobile: false
    });

    onMounted(() => {
      const userData = userInfo.getUserInfo();
      if (userData && userData.expTime > Date.now()) {
        state.userInfo = userData;
      } else {
        notification.error({
          message: "校验异常",
          description: "未获取到登录信息，请重新登录",
          duration: 3
        });
        router.push("/home/login");
      }

      state.selectedKey = router.currentRoute.value.path;
      window.addEventListener('resize', handleResize);
      handleResize();
    });

    const handleResize = () => {
      state.isMobile = window.innerWidth < 768;
    };

    const handleLoginOut = () => {
      userInfo.removeUserInfo();
      notification.success({
        message: "退出成功",
        description: "您已安全退出系统",
        duration: 2
      });
      router.push("/");
    };

    const handleMenuClick = ({ key }) => {
      if (key === 'logout') {
        handleLoginOut();
      }
    };

    const handleMenuSelect = ({ key }) => {
      state.selectedKey = key;
      router.push(key);
    };

    const collapsedTrigger = () => {
      state.isCollapsed = !state.isCollapsed;
    };

    watch(
        () => router.currentRoute.value.path,
        (newPath) => {
          state.selectedKey = newPath;
        }
    );

    return {
      ...toRefs(state),
      handleLoginOut,
      handleMenuClick,
      handleMenuSelect,
      collapsedTrigger
    };
  }
};
</script>

<style scoped>
/* 关键样式优化 */
.layout-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden; /* 防止整体页面滚动 */
}

.header {
  background: linear-gradient(135deg, #4b6ff6 0%, #1890ff 100%);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  height: 64px;
  padding: 0;
  position: relative;
  z-index: 100; /* 确保头部在最上层 */
}

.main-container {
  display: flex;
  flex: 1;
  overflow: hidden; /* 关键：隐藏容器溢出内容 */
}

.sidebar {
  /* 固定侧边栏高度，使其不随内容滚动 */
  height: calc(100vh - 64px);
  position: sticky;
  top: 64px;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
  z-index: 10;
  overflow-y: auto; /* 侧边栏内容过多时自身滚动 */
}

/* 内容容器 - 核心滚动样式 */
.content-wrapper {
  flex: 1;
  overflow: hidden; /* 限制内容区域 */
}

.main-content {
  padding: 12px;
  height: calc(100vh - 64px); /* 高度 = 视口高度 - 头部高度 */
  overflow-y: auto; /* 内容超出时滚动 */
  box-sizing: border-box; /* 确保padding不影响整体高度计算 */
}

.content-container {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  padding: 24px;
  min-height: 100%; /* 确保容器至少填满内容区域 */
}

/* 其他样式保持不变 */
.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 24px;
}

.logo {
  display: flex;
  align-items: center;
}

.logo-link {
  display: flex;
  align-items: center;
  color: #fff;
  text-decoration: none;
  transition: all 0.3s;
}

.logo-link:hover {
  opacity: 0.9;
}

.logo-img {
  width: 36px;
  height: 36px;
  margin-right: 12px;
  border-radius: 6px;
  padding: 3px;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.user-area {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  color: #fff;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 20px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.user-avatar {
  width: 32px;
  height: 32px;
  margin-right: 8px;
  border: 2px solid rgba(255, 255, 255, 0.2);
}

.user-email {
  font-size: 14px;
  max-width: 180px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-arrow {
  font-size: 12px;
  margin-left: 4px;
  transition: transform 0.2s;
}

.user-info:hover .user-arrow {
  transform: translateY(1px);
}

.user-menu {
  border-radius: 8px;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 8px 16px !important;
  transition: all 0.2s;
}

.menu-item:hover {
  background-color: #f5f7fa !important;
}

.menu-icon {
  margin-right: 8px;
  font-size: 16px;
}

.logout-item {
  color: #f5222d !important;
}

.logout-item:hover {
  background-color: #fff5f5 !important;
}

.user-mini {
  display: none;
}

.mini-logout {
  color: #fff !important;
  padding: 4px 8px !important;
}

.sidebar-menu {
  border-right: none !important;
  background-color: transparent !important;
  padding-top: 16px !important;
}

.sidebar-item {
  margin: 4px 0 !important;
  border-radius: 6px !important;
  color: rgba(255, 255, 255, 0.9) !important;
  padding: 12px 24px !important;
  transition: all 0.2s !important;
}

.sidebar-item:hover {
  background-color: rgba(255, 255, 255, 0.1) !important;
  color: #fff !important;
}

:deep(.ant-menu-item-selected) {
  background-color: rgba(255, 255, 255, 0.2) !important;
  color: #fff !important;
  font-weight: 500 !important;
}

:deep(.ant-menu-item-selected::after) {
  border-right: 3px solid #fff !important;
  border-radius: 0 3px 3px 0 !important;
}

:deep(.ant-layout-sider-zero-width-trigger) {
  background-color: #4b6ff6 !important;
  top: 16px !important;
  right: -32px !important;
  width: 32px !important;
  height: 32px !important;
  border-radius: 0 8px 8px 0 !important;
  box-shadow: 2px 0 4px rgba(0, 0, 0, 0.1) !important;
  transition: all 0.2s !important;
}

:deep(.ant-layout-sider-zero-width-trigger:hover) {
  background-color: #3a5ede !important;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .header-content {
    padding: 0 16px;
  }

  .logo-text {
    font-size: 16px;
  }

  .user-info {
    display: none;
  }

  .user-mini {
    display: block;
  }

  .main-content {
    padding: 6px;
    height: calc(100vh - 64px);
  }

  .content-container {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .logo-img {
    margin-right: 0;
  }
}

/* 滚动条美化 */
:deep(.ant-layout-content::-webkit-scrollbar) {
  width: 6px;
  height: 6px;
}

:deep(.ant-layout-content::-webkit-scrollbar-track) {
  background: #f5f5f5;
  border-radius: 3px;
}

:deep(.ant-layout-content::-webkit-scrollbar-thumb) {
  background: #ccc;
  border-radius: 3px;
}

:deep(.ant-layout-content::-webkit-scrollbar-thumb:hover) {
  background: #aaa;
}

:deep(.ant-layout-sider-trigger){
  background-color: #5577f8;
}

</style>