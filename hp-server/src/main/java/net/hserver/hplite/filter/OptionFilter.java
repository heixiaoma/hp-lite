package net.hserver.hplite.filter;

import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Order;
import cn.hserver.plugin.web.context.Webkit;
import cn.hserver.plugin.web.interfaces.FilterAdapter;
import io.netty.handler.codec.http.HttpMethod;
import io.netty.handler.codec.http.HttpResponseStatus;

/**
 * Option,跨域等操作
 */
@Bean
@Order(1)
public class OptionFilter implements FilterAdapter {
    @Override
    public void doFilter(Webkit webkit) throws Exception {
        webkit.httpResponse.setHeader("Access-Control-Allow-Origin", "*");
        webkit.httpResponse.setHeader("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE");
        webkit.httpResponse.setHeader("Access-Control-Allow-Credentials", "true");
        webkit.httpResponse.setHeader("Access-Control-Allow-Headers", "*");
        if (webkit.httpRequest.getRequestType() == HttpMethod.OPTIONS) {
            webkit.httpResponse.sendStatusCode(HttpResponseStatus.NO_CONTENT);
            webkit.httpResponse.sendText("");
        }

    }
}
