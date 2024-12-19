import http from '../../data/http'

export function getDomain(query) {
    return http({
        url: '/client/domain/list',
        method: 'get',
        params: query
    })
}

export function removeDomain(query) {
    return http({
        url: '/client/domain/remove',
        method: 'get',
        params: query
    })
}

export function addDomain(data) {
    return http({
        url: '/client/domain/add',
        method: 'post',
        data
    })
}

export function genSSL(query) {
    return http({
        url: '/client/domain/gen',
        method: 'get',
        params: query
    })
}

export function queryDomain(query) {
    return http({
        url: '/client/domain/query',
        method: 'get',
        params: query
    })
}
