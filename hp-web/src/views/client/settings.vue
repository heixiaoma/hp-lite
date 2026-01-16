<template>
  <div class="settings-page">
    <a-card class="settings-card" title="个人设置" :bordered="false">
      <a-tabs v-model:activeKey="activeTabKey" type="card">
        <!-- 邮箱设置标签页 -->
        <a-tab-pane key="email" tab="邮箱设置">
          <div class="tab-content">
            <a-form :model="emailForm" ref="emailFormRef" :rules="emailRules" layout="vertical">
              <a-form-item label="当前邮箱">
                <a-input
                  v-model:value="currentEmail"
                  disabled
                  placeholder="未设置邮箱"
                />
              </a-form-item>

              <a-divider />

              <h3>绑定新邮箱</h3>

              <a-form-item name="newEmail" label="新邮箱地址">
                <a-input
                  v-model:value="emailForm.newEmail"
                  placeholder="请输入新的邮箱地址"
                  allow-clear
                >
                  <template #prefix>
                    <MailOutlined />
                  </template>
                </a-input>
              </a-form-item>

              <a-form-item class="verification-section">
                <div class="verification-input">
                  <a-input
                    v-model:value="emailForm.code"
                    placeholder="请输入验证码"
                    allow-clear
                    maxlength="6"
                    class="code-input"
                  >
                    <template #prefix>
                      <SecurityScanOutlined />
                    </template>
                  </a-input>
                  <a-button
                    @click="sendEmailCode"
                    :loading="sendingEmailCode"
                    :disabled="!emailForm.newEmail || !isValidEmail(emailForm.newEmail)"
                    class="send-code-btn"
                  >
                    {{ emailCodeTimer > 0 ? `重新发送(${emailCodeTimer}s)` : '发送验证码' }}
                  </a-button>
                </div>
                <div class="form-help">验证码已发送到您的新邮箱，请在30分钟内使用</div>
              </a-form-item>

              <a-form-item>
                <a-button
                  type="primary"
                  @click="bindEmail"
                  :loading="bindingEmail"
                  block
                  size="large"
                >
                  绑定邮箱
                </a-button>
              </a-form-item>
            </a-form>
          </div>
        </a-tab-pane>

        <!-- 安全设置标签页 -->
        <a-tab-pane key="security" tab="安全设置">
          <div class="tab-content">
            <a-list>
              <a-list-item>
                <template #actions>
                  <a href="#" @click.prevent="handleChangePassword">修改</a>
                </template>
                <a-list-item-meta title="登录密码">
                  <template #description>
                    定期修改密码可以有效保护您的账户安全
                  </template>
                </a-list-item-meta>
              </a-list-item>

              <a-list-item>
                <template #actions>
                  <a href="#" @click.prevent="handleChangeEmail">修改</a>
                </template>
                <a-list-item-meta title="绑定邮箱">
                  <template #description>
                    已绑定邮箱 {{ currentEmail || '未绑定' }}
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </a-list>
          </div>
        </a-tab-pane>

        <!-- 账户信息标签页 -->
        <a-tab-pane key="account" tab="账户信息">
          <div class="tab-content">
            <a-descriptions :column="1" bordered>
              <a-descriptions-item label="用户ID">
                {{ userInfoData.id || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="用户名">
                {{ userInfoData.username || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="邮箱">
                {{ currentEmail || '未绑定' }}
              </a-descriptions-item>
              <a-descriptions-item label="创建时间">
                {{ formatTime(userInfoData.createTime) }}
              </a-descriptions-item>
            </a-descriptions>
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- 修改密码模态框 -->
    <a-modal
      v-model:visible="changePasswordModalVisible"
      title="修改密码"
      @ok="submitChangePassword"
      @cancel="changePasswordModalVisible = false"
    >
      <a-form :model="passwordForm" ref="passwordFormRef" :rules="passwordRules" layout="vertical">
        <a-form-item name="oldPassword" label="当前密码">
          <a-input-password
            v-model:value="passwordForm.oldPassword"
            placeholder="请输入当前密码"
          />
        </a-form-item>

        <a-form-item name="newPassword" label="新密码">
          <a-input-password
            v-model:value="passwordForm.newPassword"
            placeholder="请输入新密码"
          />
        </a-form-item>

        <a-form-item name="confirmPassword" label="确认密码">
          <a-input-password
            v-model:value="passwordForm.confirmPassword"
            placeholder="请确认新密码"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue';
import { notification } from 'ant-design-vue';
import { MailOutlined, SecurityScanOutlined } from '@ant-design/icons-vue';
import userInfo from '../../data/userInfo';
import { sendCode, getEmail, setEmail, changePassword } from '../../api/client/email';

const emailFormRef = ref(null);
const passwordFormRef = ref(null);

const activeTabKey = ref('email');
const currentEmail = ref('');
const sendingEmailCode = ref(false);
const bindingEmail = ref(false);
const changePasswordModalVisible = ref(false);
const emailCodeTimer = ref(0);

// 用户信息
const userInfoData = reactive({
  id: '',
  username: '',
  email: '',
  createTime: ''
});

const emailForm = reactive({
  newEmail: '',
  code: ''
});

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
});

const emailRules = {
  newEmail: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    {
      pattern: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
      message: '请输入有效的邮箱地址',
      trigger: 'blur'
    }
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码应为6位数字', trigger: 'blur' }
  ]
};

const passwordRules = {
  oldPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value) => {
        if (value === passwordForm.newPassword) {
          return Promise.resolve();
        } else {
          return Promise.reject(new Error('两次输入的密码不一致'));
        }
      },
      trigger: 'blur'
    }
  ]
};

const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

const sendEmailCode = () => {
  if (!emailForm.newEmail || !isValidEmail(emailForm.newEmail)) {
    notification.error({
      message: '邮箱格式错误',
      description: '请输入有效的邮箱地址'
    });
    return;
  }

  sendingEmailCode.value = true;
  sendCode({
    email: emailForm.newEmail,
    type: 'verify_email'
  }).then(res => {
    sendingEmailCode.value = false;
    if (res.code === 200) {
      notification.success({
        message: '验证码已发送',
        description: res.msg || '请检查您的邮箱'
      });
      emailCodeTimer.value = 60; // 60秒
      startEmailCodeTimer();
    } else {
      notification.error({
        message: '发送失败',
        description: res.msg || '无法发送验证码'
      });
    }
  }).catch(() => {
    sendingEmailCode.value = false;
    notification.error({
      message: '发送失败',
      description: '请稍后重试'
    });
  });
};

const bindEmail = () => {
  emailFormRef.value.validate().then(() => {
    bindingEmail.value = true;
    setEmail({
      email: emailForm.newEmail,
      code: emailForm.code
    }).then(res => {
      bindingEmail.value = false;
      if (res.code === 200) {
        notification.success({
          message: '邮箱绑定成功',
          description: res.msg || '邮箱绑定成功'
        });
        currentEmail.value = emailForm.newEmail;
        emailForm.newEmail = '';
        emailForm.code = '';
        // 更新用户信息
        const userData = userInfo.getUserInfo();
        userData.email = currentEmail.value;
        userInfo.setUserInfo(userData);
      } else {
        notification.error({
          message: '绑定失败',
          description: res.msg || '邮箱绑定失败'
        });
      }
    }).catch(() => {
      bindingEmail.value = false;
      notification.error({
        message: '绑定失败',
        description: '请稍后重试'
      });
    });
  });
};

const startEmailCodeTimer = () => {
  const timer = setInterval(() => {
    emailCodeTimer.value--;
    if (emailCodeTimer.value <= 0) {
      clearInterval(timer);
    }
  }, 1000);
};

const handleChangePassword = () => {
  changePasswordModalVisible.value = true;
};

const handleChangeEmail = () => {
  activeTabKey.value = 'email';
};

const submitChangePassword = () => {
  passwordFormRef.value.validate().then(() => {
    changePassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword
    }).then(res => {
      if (res.code === 200) {
        notification.success({
          message: '密码修改成功',
          description: '您的密码已成功修改，请使用新密码登录'
        });
        changePasswordModalVisible.value = false;
        // 清空表单
        passwordForm.oldPassword = '';
        passwordForm.newPassword = '';
        passwordForm.confirmPassword = '';
      } else {
        notification.error({
          message: '密码修改失败',
          description: res.msg || '修改密码失败，请重试'
        });
      }
    }).catch(() => {
      notification.error({
        message: '密码修改失败',
        description: '请稍后重试'
      });
    });
  });
};

const formatTime = (time) => {
  if (!time) return '-';
  return new Date(time).toLocaleString();
};

onMounted(() => {
  // 从localStorage获取用户信息
  const userData = userInfo.getUserInfo();
  if (userData) {
    userInfoData.id = userData.id || '';
    userInfoData.username = userData.email || '';
    userInfoData.createTime = userData.createTime || '';
  }

  // 获取当前邮箱
  getEmail().then(res => {
    if (res.code === 200) {
      currentEmail.value = res.data.email || '';
      userInfoData.email = res.data.email || '';
    }
  }).catch(() => {
    // 从userInfo中获取
    const data = userInfo.getUserInfo();
    currentEmail.value = data?.email || '';
    userInfoData.email = data?.email || '';
  });
});
</script>

<style scoped lang="less">
.settings-page {
  padding: 24px;

  .settings-card {
    max-width: 1000px;
  }

  .tab-content {
    padding: 20px 0;
  }

  .verification-section {
    :deep(.ant-form-item-label) {
      display: none;
    }

    .verification-input {
      display: flex;
      gap: 8px;

      .code-input {
        flex: 1;
      }

      .send-code-btn {
        white-space: nowrap;
      }
    }

    .form-help {
      font-size: 12px;
      color: #999;
      margin-top: 8px;
    }
  }
}
</style>
