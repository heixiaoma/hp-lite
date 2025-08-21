import http from '../../data/http'


export function getGithubToken(query) {
    return http({
        url: '/client/giscus/token',
        method: 'get',
        params: query
    })
}
