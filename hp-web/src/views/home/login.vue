<template>
  <div class="login-page">
    <div class="login-container">
      <div class="login-box">
        <h1 class="logo">HP-Lite 内网穿透</h1>
        <a-spin :spinning="loading">
          <a-form :model="form" @finish="handleSubmit">
            <a-form-item label="账号" name="email" :rules="[{ required: true, message: '账号必填' }]">
              <a-input v-model:value="form.email" placeholder="请输入账号"/>
            </a-form-item>
            <a-form-item label="密码" name="password" :rules="[{ required: true, message: '密码必填' }]">
              <a-input-password v-model:value="form.password" placeholder="请输入密码"/>
            </a-form-item>
            <a-form-item style="text-align: center;">
              <a-button style="width: 50%" type="primary" html-type="submit">登录</a-button>
            </a-form-item>
          </a-form>
        </a-spin>
      </div>
    </div>
  </div>
</template>

<script setup>
import {login} from "../../api/client/user";
import {onMounted, reactive, ref} from "vue";
import {notification} from "ant-design-vue";
import userInfo from "../../data/userInfo";
import {useRouter} from "vue-router";

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

<style scoped>
.login-page {
  height: 100vh;
  background-color: #f0f2f5;
}

.logo {
  margin-bottom: 24px;
  font-size: 32px;
  color: #1890ff;
  text-align: center;
}


.login-container {
  display: flex;
  justify-content: center; /* 在主轴上水平居中 */
  align-items: center; /* 在交叉轴上垂直居中 */
  height: 100vh;
}

.login-box {
  width: 350px;
  padding: 20px;
  background-color: #fff;
  border-radius: 5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

:deep(.ant-form-item-label) {
  width: 60px !important;
}

</style>
