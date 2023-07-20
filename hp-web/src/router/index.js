import {createRouter, createWebHashHistory} from "vue-router";

// 通用UI
const home = () => import('../views/home/index.vue');
const home_login = () => import('../views/home/login.vue');

// 客户端的UI
const client_monitor = () => import('../views/client/monitor.vue');
const client_config = () => import('../views/client/config.vue');
const client_device = () => import('../views/client/device.vue');
const client_teach = () => import('../views/client/teach.vue');
const manage = () => import('../views/client/manage.vue');

const routes = [
    {path: '/', component: home},

    {path: '/home/login', component: home_login},
    //前端UI
    {
        path: '/client', component: manage,
        children: [
            {path: '', redirect: '/client/monitor'},
            {path: 'monitor', component: client_monitor},
            {path: 'config', component: client_config, name: "config"},
            {path: 'device', component: client_device, name: "device"},
            {path: 'teach', component: client_teach, name: "teach"},
        ],
    },
]

// 3. 创建路由实例并传递 `routes` 配置
// 你可以在这里输入更多的配置，但我们在这里
// 暂时保持简单
export const router = createRouter({
    // 4. 内部提供了 history 模式的实现。为了简单起见，我们在这里使用 hash 模式。
    history: createWebHashHistory(),
    routes, // `routes: routes` 的缩写
})
