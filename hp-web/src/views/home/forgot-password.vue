<template>
  <div class="forgot-password-page">
    <div class="forgot-password-bg">
      <div class="bg-shape bg-shape-1"></div>
      <div class="bg-shape bg-shape-2"></div>
    </div>

    <div class="forgot-password-card">
      <div class="forgot-password-header">
        <a href="#" @click.prevent="goBack" class="back-btn">
          <i>&larr;</i> 返回登录
        </a>
      </div>

      <div class="forgot-password-container">
        <h2 class="form-title">重置密码</h2>
        <p class="form-subtitle">通过邮箱验证重置您的密码</p>

        <a-steps :current="currentStep" :status="stepStatus">
          <a-step title="邮箱验证" description="输入邮箱并验证" />
          <a-step title="设置新密码" description="设置新的登录密码" />
          <a-step title="完成" description="密码重置成功" />
        </a-steps>

        <!-- 第一步：邮箱验证 -->
        <div v-if="currentStep === 0" class="step-content">
          <a-form ref="emailFormRef" :model="emailForm" :rules="emailRules" layout="vertical">
            <a-form-item name="email" label="邮箱地址">
              <a-input
                v-model:value="emailForm.email"
                placeholder="请输入您的邮箱地址"
                allow-clear
              >
                <template #prefix>
                  <MailOutlined class="input-icon" />
                </template>
              </a-input>
            </a-form-item>

            <a-form-item>
              <a-button
                type="primary"
                @click="sendCode"
                :loading="sendingCode"
                block
                size="large"
              >
                发送验证码
              </a-button>
            </a-form-item>
          </a-form>

          <div v-if="codeSent" class="verify-code-section">
            <a-form ref="codeFormRef" :model="codeForm" :rules="codeRules" layout="vertical">
              <a-form-item name="code" label="验证码">
                <a-input
                  v-model:value="codeForm.code"
                  placeholder="请输入6位验证码"
                  allow-clear
                  maxlength="6"
                >
                  <template #prefix>
                    <SecurityScanOutlined class="input-icon" />
                  </template>
                </a-input>
              </a-form-item>

              <a-form-item>
                <a-button
                  type="primary"
                  @click="verifyCode"
                  :loading="verifyingCode"
                  block
                  size="large"
                >
                  验证
                </a-button>
              </a-form-item>

              <div class="code-timer">
                <span v-if="codeTimer > 0">验证码有效期：{{ codeTimer }} 秒</span>
                <a v-else @click="sendCode">重新发送</a>
              </div>
            </a-form>
          </div>
        </div>

        <!-- 第二步：设置新密码 -->
        <div v-if="currentStep === 1" class="step-content">
          <a-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" layout="vertical">
            <a-form-item name="password" label="新密码">
              <a-input-password
                v-model:value="passwordForm.password"
                placeholder="请输入新密码，至少6位"
              >
                <template #prefix>
                  <LockOutlined class="input-icon" />
                </template>
              </a-input-password>
            </a-form-item>

            <a-form-item name="confirmPassword" label="确认密码">
              <a-input-password
                v-model:value="passwordForm.confirmPassword"
                placeholder="请确认新密码"
              >
                <template #prefix>
                  <LockOutlined class="input-icon" />
                </template>
              </a-input-password>
            </a-form-item>

            <a-form-item>
              <a-button
                type="primary"
                @click="resetPassword"
                :loading="resettingPassword"
                block
                size="large"
              >
                重置密码
              </a-button>
            </a-form-item>
          </a-form>
        </div>

        <!-- 第三步：完成 -->
        <div v-if="currentStep === 2" class="step-content success">
          <div class="success-icon">
            <CheckCircleOutlined />
          </div>
          <h3>密码重置成功</h3>
          <p>您的密码已成功重置，请使用新密码登录</p>
          <a-button
            type="primary"
            @click="goToLogin"
            block
            size="large"
          >
            返回登录
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { notification } from 'ant-design-vue';
import {
  MailOutlined,
  LockOutlined,
  SecurityScanOutlined,
  CheckCircleOutlined
} from '@ant-design/icons-vue';
import { sendCode as sendCodeApi, verifyEmail as verifyCodeApi, resetPassword as resetPasswordApi } from '../../api/client/email';

const router = useRouter();
const emailFormRef = ref(null);
const codeFormRef = ref(null);
const passwordFormRef = ref(null);

const currentStep = ref(0);
const stepStatus = ref('process');
const codeSent = ref(false);
const sendingCode = ref(false);
const verifyingCode = ref(false);
const resettingPassword = ref(false);
const codeTimer = ref(0);

const emailForm = reactive({
  email: ''
});

const codeForm = reactive({
  code: ''
});

const passwordForm = reactive({
  password: '',
  confirmPassword: ''
});

const emailRules = {
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    {
      pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
      message: '请输入有效的邮箱地址',
      trigger: 'blur'
    }
  ]
};

const codeRules = {
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码应为6位数字', trigger: 'blur' }
  ]
};

const passwordRules = {
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value) => {
        if (value === passwordForm.password) {
          return Promise.resolve();
        } else {
          return Promise.reject(new Error('两次输入的密码不一致'));
        }
      },
      trigger: 'blur'
    }
  ]
};

const sendCode = () => {
  emailFormRef.value.validate().then(() => {
    sendingCode.value = true;
    sendCodeApi({
      email: emailForm.email,
      type: 'reset_password'
    }).then(res => {
      sendingCode.value = false;
      if (res.code === 200) {
        codeSent.value = true;
        codeTimer.value = 60; // 60秒
        notification.success({
          message: '验证码已发送',
          description: res.msg || '请检查您的邮箱'
        });
        startCodeTimer();
      } else {
        notification.error({
          message: '发送失败',
          description: res.msg || '无法发送验证码'
        });
      }
    }).catch(() => {
      sendingCode.value = false;
      notification.error({
        message: '发送失败',
        description: '请稍后重试'
      });
    });
  });
};

const verifyCode = () => {
  codeFormRef.value.validate().then(() => {
    verifyingCode.value = true;
    verifyCodeApi({
      email: emailForm.email,
      code: codeForm.code,
      type: 'reset_password'
    }).then(res => {
      verifyingCode.value = false;
      if (res.code === 200) {
        notification.success({
          message: '验证成功',
          description: res.msg || '请设置新密码'
        });
        currentStep.value = 1;
      } else {
        notification.error({
          message: '验证失败',
          description: res.msg || '验证码无效或已过期'
        });
      }
    }).catch(() => {
      verifyingCode.value = false;
      notification.error({
        message: '验证失败',
        description: '请稍后重试'
      });
    });
  });
};

const resetPassword = () => {
  passwordFormRef.value.validate().then(() => {
    resettingPassword.value = true;
    resetPasswordApi({
      email: emailForm.email,
      code: codeForm.code,
      password: passwordForm.password
    }).then(res => {
      resettingPassword.value = false;
      if (res.code === 200) {
        notification.success({
          message: '密码重置成功',
          description: res.msg || '请使用新密码登录'
        });
        currentStep.value = 2;
      } else {
        notification.error({
          message: '重置失败',
          description: res.msg || '无法重置密码'
        });
      }
    }).catch(() => {
      resettingPassword.value = false;
      notification.error({
        message: '重置失败',
        description: '请稍后重试'
      });
    });
  });
};

const startCodeTimer = () => {
  const timer = setInterval(() => {
    codeTimer.value--;
    if (codeTimer.value <= 0) {
      clearInterval(timer);
    }
  }, 1000);
};

const goBack = () => {
  router.push('/');
};

const goToLogin = () => {
  router.push('/');
};

onMounted(() => {
  const savedEmail = localStorage.getItem('hp-lite-forgot-email');
  if (savedEmail) {
    emailForm.email = savedEmail;
  }
});
</script>

<style scoped lang="less">
.forgot-password-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f0f2f5;
  position: relative;
  overflow: hidden;

  .forgot-password-bg {
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

.forgot-password-card {
  width: 500px;
  background-color: white;
  border-radius: 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  position: relative;
  z-index: 1;

  @media (max-width: 600px) {
    width: 90%;
  }

  .forgot-password-header {
    padding: 20px;
    border-bottom: 1px solid #f0f0f0;

    .back-btn {
      color: #4b6ff6;
      text-decoration: none;
      font-size: 14px;
      transition: all 0.3s ease;

      &:hover {
        color: #1890ff;
      }
    }
  }
}

.forgot-password-container {
  padding: 40px;

  @media (max-width: 600px) {
    padding: 20px;
  }

  .form-title {
    font-size: 28px;
    font-weight: 600;
    color: #28313b;
    margin-bottom: 10px;
  }

  .form-subtitle {
    font-size: 14px;
    color: #808695;
    margin-bottom: 30px;
  }

  :deep(.ant-steps) {
    margin-bottom: 30px;
  }

  .step-content {
    animation: fadeIn 0.3s ease-out;

    .input-icon {
      color: #808695;
      font-size: 16px;
    }

    .verify-code-section {
      margin-top: 30px;
      padding-top: 20px;
      border-top: 1px solid #f0f0f0;
    }

    .code-timer {
      text-align: center;
      font-size: 12px;
      color: #999;
      margin-top: 10px;

      a {
        color: #4b6ff6;
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
      }
    }

    &.success {
      text-align: center;

      .success-icon {
        font-size: 80px;
        color: #52c41a;
        margin-bottom: 20px;

        :deep(svg) {
          width: 1em;
          height: 1em;
        }
      }

      h3 {
        font-size: 20px;
        color: #28313b;
        margin-bottom: 10px;
      }

      p {
        color: #808695;
        margin-bottom: 30px;
      }
    }
  }
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
