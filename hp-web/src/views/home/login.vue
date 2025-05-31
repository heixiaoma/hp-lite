<template>

  <div class="view-account">
    <div class="view-account-container">
      <div class="view-account-top">
        <div class="view-account-top-title">HP-Lite 内网穿透</div>
        <div class="view-account-top-desc">欢迎使用内网穿透后台管理系统</div>
      </div>
      <div class="view-account-form">
        <a-form ref="loginFormRef" size="large" :model="form" :rules="rules">
          <a-form-item name="username">
            <a-input v-model:value="form.email" allow-clear placeholder="请输入用户名">
              <template #prefix>
                <user-outlined :style="{ color: '#808695' }" />
              </template>
            </a-input>
          </a-form-item>
          <a-form-item name="password">
            <a-input-password
                v-model:value="form.password"
                allow-clear
                placeholder="请输入密码"
                @keyup.enter="handleSubmit"
            >
              <template #prefix>
                <lock-outlined :style="{ color: '#808695' }" />
              </template>
            </a-input-password>
          </a-form-item>
          <a-form-item>
            <a-button type="primary" @click="handleSubmit" size="large" :loading="loading" block>
              登录
            </a-button>
          </a-form-item>
        </a-form>
      </div>
    </div>
  </div>


<!--  <div class="login-page">-->
<!--    <div class="login-container">-->
<!--      <div class="login-box">-->
<!--        <h1 class="logo"></h1>-->
<!--        <a-spin :spinning="loading">-->
<!--          <a-form :model="form" @finish="handleSubmit">-->
<!--            <a-form-item label="账号" name="email" :rules="[{ required: true, message: '账号必填' }]">-->
<!--              <a-input v-model:value="form.email" placeholder="请输入账号"/>-->
<!--            </a-form-item>-->
<!--            <a-form-item label="密码" name="password" :rules="[{ required: true, message: '密码必填' }]">-->
<!--              <a-input-password v-model:value="form.password" placeholder="请输入密码"/>-->
<!--            </a-form-item>-->
<!--            <a-form-item style="text-align: center;">-->
<!--              <a-button style="width: 50%" type="primary" html-type="submit">登录</a-button>-->
<!--            </a-form-item>-->
<!--          </a-form>-->
<!--        </a-spin>-->
<!--      </div>-->
<!--    </div>-->
<!--  </div>-->
</template>

<script setup>
import {login} from "../../api/client/user";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";
import userInfo from "../../data/userInfo";
import {useRouter} from "vue-router";
import { LockOutlined, UserOutlined } from '@ant-design/icons-vue';
const router = useRouter()

const loading = ref(false)

const form = reactive({
  email: '',
  password: '',
});


onMounted(() => {
  let data = userInfo.getUserInfo()
  if (data && data.expTime > Date.parse(new Date())) {
    router.push("/client")
  }
})


const handleSubmit = () => {
  console.log("---表单数据---", form)
  loading.value = true
  login(form).then(res => {
    loading.value = false
    if (res.code === 200) {
      notification.open({
        message: res.msg,
      })
      //存储数据
      userInfo.setUserInfo(res.data)
      router.push("/client")
    }
  }).catch(e => {
    loading.value = false
  })
}

</script>

<style scoped lang="less">
.view-account {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: auto;

  &-container {
    flex: 1;
    padding: 32px 0;
    width: 380px;
    margin: 0 auto;
  }

  &-top {
    padding: 32px 0;
    text-align: center;

    &-title {
      font-size: 44px;
      color: #28313b;
      margin-bottom: 20px;
    }

    &-desc {
      font-size: 14px;
      color: #808695;
    }
  }

  .default-color {
    color: #515a6e;
  }
}

@media (min-width: 768px) {
  .view-account {
    background-image: url('../assets/login/login.svg');
    background-repeat: no-repeat;
    background-position: 50%;
    background-size: 100%;
  }
}
</style>
