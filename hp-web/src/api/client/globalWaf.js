import http from '../../data/http'

// 全局 Web 安全规则：系统级，对所有进站流量生效，接口只有管理员能调
export function getGlobalWaf(query) {
    return http({
        url: '/client/globalWaf/list',
        method: 'get',
        params: query
    })
}

export function removeGlobalWaf(query) {
    return http({
        url: '/client/globalWaf/remove',
        method: 'get',
        params: query
    })
}

export function saveGlobalWaf(data) {
    return http({
        url: '/client/globalWaf/save',
        method: 'post',
        data
    })
}
