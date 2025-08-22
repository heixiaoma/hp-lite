import http from '../../data/http'

export function getDeviceList(query) {
    return http({
        url: '/client/device/list',
        method: 'get',
        params: query
    })
}


export function addDevice(data) {
    return http({
        url: '/client/device/add',
        method: 'post',
        data
    })
}

export function removeDevice(query) {
    return http({
        url: '/client/device/remove',
        method: 'get',
        params: query
    })
}

export function stopDevice(query) {
    return http({
        url: '/client/device/stop',
        method: 'get',
        params: query
    })
}

export function updateDevice(data) {
    return http({
        url: '/client/device/update',
        method: 'post',
        data
    })
}
