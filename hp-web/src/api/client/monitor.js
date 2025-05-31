import http from '../../data/http'

export function monitorList(query) {
    return http({
        url: '/client/monitor/list',
        method: 'get',
        params: query
    })
}
export function monitorDetail(query) {
    return http({
        url: '/client/monitor/detail',
        method: 'get',
        params: query
    })
}
