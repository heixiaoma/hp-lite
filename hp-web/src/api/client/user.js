import http from '../../data/http'

export function login(data) {
    return http({
        url: '/user/login',
        method: 'post',
        data
    })
}
