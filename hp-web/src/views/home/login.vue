<template>
  <div class="login-page">
    <!-- 背景装饰元素 -->
    <div class="login-bg">
      <div class="bg-shape bg-shape-1"></div>
      <div class="bg-shape bg-shape-2"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-card">
      <!-- 左侧品牌区 -->
      <div class="login-card-left">
        <div class="login-logo">
          <img src="/logo-back.png" alt="HP-Lite Logo" class="logo-image">
          <span class="logo-text">HP-Lite</span>
        </div>

        <div class="login-info">
          <h2 class="info-title">内网穿透</h2>
          <p class="info-desc">无需公网IP，轻松实现内网服务外网访问</p>

          <div class="login-features">
            <div class="feature-item">
              <i class="feature-icon"><CheckCircleOutlined /></i>
              <span class="feature-text">多平台支持</span>
            </div>
            <div class="feature-item">
              <i class="feature-icon"><CheckCircleOutlined /></i>
              <span class="feature-text">自定义域名</span>
            </div>
            <div class="feature-item">
              <i class="feature-icon"><CheckCircleOutlined /></i>
              <span class="feature-text">HTTPS支持</span>
            </div>
            <div class="feature-item">
              <i class="feature-icon"><CheckCircleOutlined /></i>
              <span class="feature-text">双通道多协议</span>
            </div>
            <div class="feature-item">
              <i class="feature-icon"><CheckCircleOutlined /></i>
              <span class="feature-text">云端管理</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧表单区 -->
      <div class="login-card-right">
        <div class="login-form-container">
          <h2 class="form-title">欢迎回来</h2>
          <p class="form-subtitle">请登录您的账户继续使用</p>

          <a-form ref="loginFormRef" size="large" :model="form" :rules="rules">
            <a-form-item name="email">
              <a-input
                  v-model:value="form.email"
                  allow-clear
                  placeholder="用户名或邮箱"
                  class="form-input"
              >
                <template #prefix>
                  <UserOutlined class="input-icon" />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item name="password">
              <a-input-password
                  v-model:value="form.password"
                  allow-clear
                  placeholder="密码"
                  class="form-input"
                  @keyup.enter="handleSubmit"
              >
                <template #prefix>
                  <LockOutlined class="input-icon" />
                </template>
              </a-input-password>
            </a-form-item>

            <a-form-item class="remember-me">
              <a-checkbox v-model:checked="rememberMe">记住我</a-checkbox>
            </a-form-item>

            <a-form-item>
              <a-button
                  type="primary"
                  @click="handleSubmit"
                  size="large"
                  :loading="loading"
                  class="login-button"
                  block
              >
                登录
              </a-button>
            </a-form-item>
          </a-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import { notification } from 'ant-design-vue';
import { login } from "../../api/client/user";
import userInfo from "../../data/userInfo";
import { CheckCircleOutlined, LockOutlined, UserOutlined } from '@ant-design/icons-vue';

const router = useRouter();
const loginFormRef  = ref(null);
const loading = ref(false);
const rememberMe = ref(false);

const form = reactive({
  email: '',
  password: '',
});

const rules = reactive({
  email: [
    { required: true, message: '请输入用户名或邮箱', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
  ],
});

onMounted(() => {
  const savedUser = localStorage.getItem('hp-lite-user');
  if (savedUser) {
    const { email, password, expTime } = JSON.parse(savedUser);
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

const handleSubmit = () => {
  loginFormRef.value.validate().then(() => {
    loading.value = true;
    login(form).then(res => {
      loading.value = false;
      if (res.code === 200) {
        notification.success({
          message: '登录成功',
          description: '欢迎回来，正在为您跳转...',
        });
        // 存储用户信息
        userInfo.setUserInfo(res.data);
        // 记住登录状态
        if (rememberMe.value) {
          localStorage.setItem('hp-lite-user', JSON.stringify({
            email: form.email,
            password: form.password,
            expTime: Date.now() + 7 * 24 * 60 * 60 * 1000 // 7天有效期
          }));
        } else {
          localStorage.removeItem('hp-lite-user');
        }

        router.push("/client");
      } else {
        notification.error({
          message: '登录失败',
          description: res.msg || '用户名或密码错误',
        });
      }
    }).catch(error => {
      loading.value = false;
    });
  }).catch(error => {
    console.log('表单验证失败:', error);
  });
};
</script>

<style scoped lang="less">
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f0f2f5;
  position: relative;
  overflow: hidden;

  .login-bg {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    z-index: 0;

    .bg-shape {
      position: absolute;
      width: 600px;
      height: 600px;
      border-radius: 50%;
      filter: blur(100px);
      opacity: 0.3;
    }

    .bg-shape-1 {
      background-color: #4b6ff6;
      top: -300px;
      left: -300px;
    }

    .bg-shape-2 {
      background-color: #1890ff;
      bottom: -300px;
      right: -300px;
    }
  }
}

.login-card {
  display: flex;
  width: 900px;
  height: 600px;
  background-color: white;
  border-radius: 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  position: relative;
  z-index: 1;

  @media (max-width: 992px) {
    width: 90%;
    height: auto;
    flex-direction: column;
  }

  .login-card-left {
    flex: 1;
    background: linear-gradient(135deg, #4b6ff6 0%, #1890ff 100%);
    color: white;
    padding: 100px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;

    @media (max-width: 992px) {
      padding: 40px;
    }

    .login-logo {
      display: flex;
      align-items: center;

      .logo-image {
        width: 50px;
        height: 50px;
        margin-right: 15px;
      }

      .logo-text {
        color: #ffffff;
        font-size: 24px;
        font-weight: 600;
      }
    }

    .login-info {
      .info-title {
        color: #ffffff;
        font-size: 32px;
        font-weight: 500;
        margin-bottom: 20px;
      }

      .info-desc {
        font-size: 16px;
        opacity: 0.8;
        margin-bottom: 40px;
      }

      .login-features {
        .feature-item {
          display: flex;
          align-items: center;
          margin-bottom: 15px;

          .feature-icon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 24px;
            height: 24px;
            background-color: rgba(255, 255, 255, 0.2);
            border-radius: 50%;
            margin-right: 10px;
          }

          .feature-text {
            opacity: 0.9;
          }
        }
      }
    }
  }

  .login-card-right {
    flex: 1;
    padding: 60px;
    display: flex;
    align-items: center;

    @media (max-width: 992px) {
      padding: 40px;
    }

    .login-form-container {
      width: 100%;

      .form-title {
        font-size: 28px;
        font-weight: 600;
        color: #28313b;
        margin-bottom: 10px;
      }

      .form-subtitle {
        font-size: 16px;
        color: #808695;
        margin-bottom: 40px;
      }

      .form-input {
        height: 50px;
        border-radius: 8px;
        border: 1px solid #d9d9d9;

        &:focus {
          border-color: #4b6ff6;
          box-shadow: 0 0 0 2px rgba(75, 111, 246, 0.2);
        }
      }

      .input-icon {
        color: #808695;
        font-size: 18px;
      }

      .remember-me {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 30px;

        .forgot-password {
          color: #4b6ff6;
          text-decoration: none;

          &:hover {
            text-decoration: underline;
          }
        }
      }

      .login-button {
        height: 50px;
        border-radius: 8px;
        background: linear-gradient(135deg, #4b6ff6 0%, #1890ff 100%);
        font-size: 16px;
        font-weight: 500;

        &:hover, &:focus {
          background-color: #3a5ee6;
          border-color: #3a5ee6;
        }
      }

      .login-register {
        text-align: center;
        margin-top: 20px;
        color: #808695;

        .register-link {
          color: #4b6ff6;
          text-decoration: none;
          font-weight: 500;

          &:hover {
            text-decoration: underline;
          }
        }
      }
    }
  }
}

/* 动画效果 */
.login-card {
  animation: fadeIn 0.6s ease-out;

  .login-card-left {
    animation: slideInLeft 0.6s ease-out;
  }

  .login-card-right {
    animation: slideInRight 0.6s ease-out;
  }
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes slideInLeft {
  from { transform: translateX(-100px); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}

@keyframes slideInRight {
  from { transform: translateX(100px); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}

/* 表单元素动画 */
.form-input {
  transition: all 0.3s ease;
}

.login-button {
  transition: all 0.3s ease;
  transform: translateY(0);

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(75, 111, 246, 0.3);
  }

  &:active {
    transform: translateY(0);
  }
}
</style>