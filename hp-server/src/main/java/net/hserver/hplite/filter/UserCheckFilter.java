package net.hserver.hplite.filter;

import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Order;
import cn.hserver.core.server.util.JsonResult;
import cn.hserver.plugin.web.context.Webkit;
import cn.hserver.plugin.web.interfaces.FilterAdapter;
import cn.hserver.plugin.web.interfaces.HttpRequest;
import net.hserver.hplite.domian.bean.Token;
import net.hserver.hplite.utils.TokenUtil;

/**
 * 用户端权限校验->/client/**
 */
@Bean
@Order(2)
public class UserCheckFilter implements FilterAdapter {

    private final static String uri1 = "/client/";

    @Override
    public void doFilter(Webkit webkit) throws Exception {
        HttpRequest httpRequest = webkit.httpRequest;
        if (httpRequest.getNettyUri().startsWith(uri1)) {
            String token = httpRequest.getHeader("token");
            Token tokenInfo = TokenUtil.getToken(token);
            if (token == null || token.trim().length() == 0 || tokenInfo == null || tokenInfo.getRole() != Token.Role.ADMIN) {
                webkit.httpResponse.sendJson(JsonResult.error(-2, "超级用户权限校验失败"));
            }
        }
    }
}
