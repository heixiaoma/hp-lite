import axios from 'axios'
import {notification} from 'ant-design-vue';
import userInfo from './userInfo.js'


// create an axios instance
const service = axios.create({
    // baseURL: "http://127.0.0.1:9090", // url = base url + request url
    // baseURL: "http://38.147.185.236:9090", // url = base url + request url
    // baseURL: "http://xxx.com", // url = base url + request url
    baseURL: "", // url = base url + request url
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
            notification.open({
                message: "请求异常",
                description: res.msg || 'Error'
            })
            // 50008: Illegal token; 50012: Other clients logged in; 50014: Token expired;
            if (res.code === -2 || res.code === -3 || res.code === -4 || res.code === -5) {
                // to re-login
                notification.open({
                    message: "重新登录",
                    description: "登录过期，重新登录试试吧",
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
        notification.open({
            message: "请求失败",
            description: error.message,
        })

        return Promise.reject(error)
    }
)

export default service
