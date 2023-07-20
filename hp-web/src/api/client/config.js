import http from '../../data/http'

export function getConfigList(query) {
    return http({
        url: '/client/config/getConfigList',
        method: 'get',
        params: query
    })
}


export function removeConfig(query) {
    return http({
        url: '/client/config/removeConfig',
        method: 'get',
        params: query
    })
}

export function refConfig(query) {
    return http({
        url: '/client/config/refConfig',
        method: 'get',
        params: query
    })
}
export function addConfig(data) {
    return http({
        url: '/client/config/addConfig',
        method: 'post',
        data
    })
}


export function getDeviceKey(query) {
    return http({
        url: '/client/config/getDeviceKey',
        method: 'get',
        params: query
    })
}
