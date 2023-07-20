import {createApp} from 'vue'
import App from './App.vue'
import {router} from "./router/index.js";
import Antd from 'ant-design-vue';
import 'ant-design-vue/dist/antd.css';

const elementApp = createApp(App);
elementApp.use(Antd)
elementApp.use(router)
elementApp.mount('#app')
