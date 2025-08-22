import http from '../../data/http'

export function getUser(query) {
    return http({
        url: '/client/user/list',
        method: 'get',
        params: query
    })
}

export function removeUser(query) {
    return http({
        url: '/client/user/removeUser',
        method: 'get',
        params: query
    })
}

export function saveUser(data) {
    return http({
        url: '/client/user/saveUser',
        method: 'post',
        data
    })
}
