import http from '../../data/http'

export function sendCode(data) {
    return http({
        url: '/email/sendCode',
        method: 'post',
        data
    })
}

export function verifyEmail(data) {
    return http({
        url: '/email/verifyEmail',
        method: 'post',
        data
    })
}

export function resetPassword(data) {
    return http({
        url: '/email/resetPassword',
        method: 'post',
        data
    })
}

export function setEmail(data) {
    return http({
        url: '/user/setEmail',
        method: 'post',
        data
    })
}

export function getEmail() {
    return http({
        url: '/user/getEmail',
        method: 'get'
    })
}

export function changePassword(data) {
    return http({
        url: '/user/changePassword',
        method: 'post',
        data
    })
}
