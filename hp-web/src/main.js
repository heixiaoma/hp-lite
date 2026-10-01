import {createApp} from 'vue'
import App from './App.vue'
import {router} from "./router/index.js";
import TDesign from 'tdesign-vue-next';
import 'tdesign-vue-next/es/style/index.css';
import './styles/global.css';

const app = createApp(App);
app.use(TDesign)
app.use(router)
app.mount('#app')
