import http from '../../data/http'

export function getReverse(query) {
    return http({
        url: '/client/reverse/list',
        method: 'get',
        params: query
    })
}

export function removeReverse(query) {
    return http({
        url: '/client/reverse/remove',
        method: 'get',
        params: query
    })
}

export function saveReverse(data) {
    return http({
        url: '/client/reverse/save',
        method: 'post',
        data
    })
}
