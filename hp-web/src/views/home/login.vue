<template>
  <div class="login">
    <div class="login__bg">
      <span class="login__blob login__blob--1"></span>
      <span class="login__blob login__blob--2"></span>
      <span class="login__blob login__blob--3"></span>
    </div>

    <div class="login__card">
      <!-- 左侧品牌区 -->
      <aside class="login__brand">
        <div class="login__brand-top">
          <div class="login__logo">
            <img src="/logo-back.png" alt="HP-Lite Logo"/>
            <span>HP-Lite</span>
          </div>
        </div>

        <div class="login__brand-body">
          <h2 class="login__brand-title">内网穿透</h2>
          <p class="login__brand-desc">无需公网 IP，轻松实现内网服务外网访问</p>

          <ul class="login__features">
            <li v-for="f in features" :key="f">
              <check-circle-icon/>
              <span>{{ f }}</span>
            </li>
          </ul>
        </div>

        <div class="login__brand-foot">
          <a href="javascript:void(0)" @click="goHome">← 返回首页</a>
        </div>
      </aside>

      <!-- 右侧表单区 -->
      <section class="login__form">
        <div class="login__form-inner">
          <h2 class="login__title">欢迎回来</h2>
          <p class="login__subtitle">请登录您的账户继续使用</p>

          <t-form
              ref="loginFormRef"
              :data="form"
              :rules="rules"
              layout="vertical"
              size="large"
              :label-width="0"
          >
            <t-form-item name="email">
              <t-input v-model="form.email" clearable placeholder="用户名或邮箱" size="large">
                <template #prefix-icon><user-icon/></template>
              </t-input>
            </t-form-item>

            <t-form-item name="password">
              <t-input
                  v-model="form.password"
                  type="password"
                  clearable
                  placeholder="密码"
                  size="large"
                  @enter="handleSubmit"
              >
                <template #prefix-icon><lock-on-icon/></template>
              </t-input>
            </t-form-item>

            <div class="login__remember">
              <t-checkbox v-model="rememberMe">记住我</t-checkbox>
            </div>

            <t-button theme="primary" size="large" block :loading="loading" @click="handleSubmit">
              登录
            </t-button>
          </t-form>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import {onMounted, reactive, ref} from 'vue';
import {useRouter} from 'vue-router';
import {MessagePlugin} from 'tdesign-vue-next';
import {login} from "../../api/client/user";
import userInfo from "../../data/userInfo";
import {CheckCircleIcon, LockOnIcon, UserIcon} from 'tdesign-icons-vue-next';

const router = useRouter();
const loginFormRef = ref(null);
const loading = ref(false);
const rememberMe = ref(false);

const features = [
  '云端集中管控',
  '多设备多租户管理',
  '域名与证书自动化管理',
  '全协议流量可视化分析',
  '穿透安全防护体系',
  '正反向代理一体化支持',
];

const form = reactive({
  email: '',
  password: '',
});

const rules = reactive({
  email: [
    {required: true, message: '请输入用户名或邮箱', trigger: 'blur'},
  ],
  password: [
    {required: true, message: '请输入密码', trigger: 'blur'},
    {min: 6, message: '密码长度不能少于6位', trigger: 'blur'},
  ],
});

onMounted(() => {
  const savedUser = localStorage.getItem('hp-lite-user');
  if (savedUser) {
    const {email, password, expTime} = JSON.parse(savedUser);
    if (expTime > Date.now()) {
      form.email = email;
      form.password = password;
      rememberMe.value = true;
    }
  }

  const userData = userInfo.getUserInfo();
  if (userData && userData.expTime > Date.now()) {
    router.push("/client");
  }
});

const goHome = () => {
  router.push('/');
};

const handleSubmit = async () => {
  const result = await loginFormRef.value?.validate();
  if (result !== true) return;

  loading.value = true;
  login(form).then(res => {
    loading.value = false;
    if (res.code === 200) {
      MessagePlugin.success('登录成功，正在为您跳转...');
      // 存储用户信息
      userInfo.setUserInfo(res.data);
      // 记住登录状态
      if (rememberMe.value) {
        localStorage.setItem('hp-lite-user', JSON.stringify({
          email: form.email,
          password: form.password,
          expTime: Date.now() + 30 * 24 * 60 * 60 * 1000 // 30天有效期
        }));
      } else {
        localStorage.removeItem('hp-lite-user');
      }

      router.push("/client");
    } else {
      MessagePlugin.error(res.msg || '用户名或密码错误');
    }
  }).catch(() => {
    loading.value = false;
  });
};
</script>

<style scoped>
.login {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 24px;
  background: var(--hp-bg);
  overflow: hidden;
}

.login__bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.login__blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(90px);
}

.login__blob--1 {
  width: 520px;
  height: 520px;
  background: rgba(75, 111, 246, .35);
  top: -220px;
  left: -160px;
}

.login__blob--2 {
  width: 480px;
  height: 480px;
  background: rgba(47, 139, 251, .3);
  bottom: -200px;
  right: -140px;
}

.login__blob--3 {
  width: 360px;
  height: 360px;
  background: rgba(34, 211, 238, .22);
  top: 30%;
  right: 30%;
}

.login__card {
  position: relative;
  display: flex;
  width: 940px;
  max-width: 100%;
  min-height: 600px;
  background: #fff;
  border-radius: 24px;
  overflow: hidden;
  box-shadow: var(--hp-shadow-lg);
}

/* 左侧品牌 */
.login__brand {
  flex: 1;
  background: var(--hp-grad);
  color: #fff;
  padding: 44px 48px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  position: relative;
  overflow: hidden;
}

.login__brand::after {
  content: '';
  position: absolute;
  width: 420px;
  height: 420px;
  border-radius: 50%;
  background: rgba(255, 255, 255, .1);
  right: -180px;
  bottom: -200px;
}

.login__logo {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
}

.login__logo img {
  width: 44px;
  height: 44px;
}

.login__logo span {
  font-size: 23px;
  font-weight: 700;
}

.login__brand-body {
  position: relative;
}

.login__brand-title {
  margin: 0 0 12px;
  font-size: 36px;
  font-weight: 600;
  color: #fff;
}

.login__brand-desc {
  margin: 0 0 34px;
  font-size: 15px;
  color: rgba(255, 255, 255, .85);
}

.login__features {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.login__features li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: rgba(255, 255, 255, .92);
}

.login__features li :deep(svg) {
  flex: none;
  font-size: 18px;
  color: rgba(255, 255, 255, .9);
}

.login__brand-foot {
  position: relative;
  font-size: 13px;
}

.login__brand-foot a {
  color: rgba(255, 255, 255, .85);
}

.login__brand-foot a:hover {
  color: #fff;
  text-decoration: underline;
}

/* 右侧表单 */
.login__form {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 48px;
}

.login__form-inner {
  width: 100%;
}

.login__title {
  margin: 0 0 8px;
  font-size: 27px;
  font-weight: 700;
  color: var(--hp-text);
}

.login__subtitle {
  margin: 0 0 30px;
  font-size: 14px;
  color: var(--hp-text-2);
}

.login__remember {
  /* flex 消除行内基线的 3px 留白，让间距可精确控制 */
  display: flex;
  align-items: center;
  /* 上 14px（紧贴所属字段，表归属）/ 下 24px（与提交按钮拉开） */
  margin: 14px 0 24px;
  /* 左边缘与标题 / 输入框 / 按钮严格对齐，不做内缩 */
  padding-left: 0;
}

.login__remember :deep(.t-checkbox__label) {
  color: var(--hp-text-2);
  transition: color .2s;
}

.login__remember :deep(.t-checkbox:hover .t-checkbox__label) {
  color: var(--hp-text);
}

/* TDesign 默认把输入框前缀图标设为 placeholder 灰（40% 黑），几乎看不见，这里提亮并用品牌色 */
.login__form :deep(.t-input__prefix-icon) {
  color: var(--td-brand-color);
  font-size: 20px;
}

.login__form :deep(.t-input__prefix-icon .t-icon) {
  color: inherit;
}

@media (max-width: 900px) {
  .login__card {
    flex-direction: column;
    min-height: auto;
  }

  .login__brand {
    padding: 32px;
  }

  .login__brand-title {
    font-size: 28px;
  }

  .login__features {
    display: none;
  }

  .login__brand-foot {
    display: none;
  }

  .login__form {
    padding: 32px;
  }
}
</style>
