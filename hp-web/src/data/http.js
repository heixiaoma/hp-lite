import axios from 'axios'
import {NotifyPlugin} from 'tdesign-vue-next';
import userInfo from './userInfo.js'


/* 后端地址。
   指定了就用指定的（环境变量 VITE_API_BASE_URL，方便本地调试时连线上后端）；
   没指定就是空串，此时按浏览器地址栏推导 —— 前后端同机部署不用改代码也不用重新打包。
   导出是为了让 WebSocket 能按同一套地址推导 ws:// / wss:// */
const CONFIGURED_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').trim().replace(/\/+$/, '')

// 只取协议 + 域名 + 端口：接口都是 /client/... 这种绝对路径，带上 path 会拼错
export const baseURL = CONFIGURED_BASE_URL || window.location.origin

// create an axios instance
const service = axios.create({
    baseURL, // url = base url + request url
    // withCredentials: true, // send cookies when cross-domain requests
    timeout: 500000 // request timeout
})

// request interceptor
service.interceptors.request.use(
    config => {
        // do something before request is sent

        if (userInfo.getUserInfo()) {
            // let each request carry token
            // ['X-Token'] is a custom headers key
            // please modify it according to the actual situation
            config.headers['token'] = userInfo.getUserInfo().token
        }
        return config
    },
    error => {
        // do something with request error
        console.log(error) // for debug
        return Promise.reject(error)
    }
)

// response interceptor
service.interceptors.response.use(
    /**
     * If you want to get http information such as headers or status
     * Please return  response => response
     */

    /**
     * Determine the request status by custom code
     * Here is just an example
     * You can also judge the status by HTTP Status Code
     */
    response => {
        const res = response.data
        // if the custom code is not 20000, it is judged as an error.
        if (res.code !== 200) {
            NotifyPlugin.error({
                title: "请求异常",
                content: res.msg || 'Error'
            })
            // 50008: Illegal token; 50012: Other clients logged in; 50014: Token expired;
            if (res.code === -2 || res.code === -3 || res.code === -4 || res.code === -5) {

                userInfo.removeUserInfo()

                // to re-login
                NotifyPlugin.warning({
                    title: "重新登录",
                    content: "登录过期，重新登录试试吧",
                })
                location.href = "/"
            }
            return Promise.reject(new Error(res.msg || 'Error'))
        } else {
            return res
        }
    },
    error => {
        console.log('err' + error) // for debug
        NotifyPlugin.error({
            title: "请求失败",
            content: error.message,
        })

        return Promise.reject(error)
    }
)

export default service
