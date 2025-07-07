import http from '../../data/http'

export function getWaf(query) {
    return http({
        url: '/client/waf/list',
        method: 'get',
        params: query
    })
}

export function removeWaf(query) {
    return http({
        url: '/client/waf/removeUser',
        method: 'get',
        params: query
    })
}

export function saveWaf(data) {
    return http({
        url: '/client/waf/saveUser',
        method: 'post',
        data
    })
}
