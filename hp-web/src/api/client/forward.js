import http from '../../data/http'

export function getForward(query) {
    return http({
        url: '/client/forward/list',
        method: 'get',
        params: query
    })
}

export function removeForward(query) {
    return http({
        url: '/client/forward/remove',
        method: 'get',
        params: query
    })
}

export function saveForward(data) {
    return http({
        url: '/client/forward/save',
        method: 'post',
        data
    })
}
