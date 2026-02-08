import http from '../../data/http'

export function getSafe(query) {
    return http({
        url: '/client/safe/list',
        method: 'get',
        params: query
    })
}

export function removeSafe(query) {
    return http({
        url: '/client/safe/remove',
        method: 'get',
        params: query
    })
}

export function saveSafe(data) {
    return http({
        url: '/client/safe/save',
        method: 'post',
        data
    })
}
export function querySafe(query) {
    return http({
        url: '/client/safe/query',
        method: 'get',
        params: query
    })
}