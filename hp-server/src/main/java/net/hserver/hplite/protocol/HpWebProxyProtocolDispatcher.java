package net.hserver.hplite.protocol;

import cn.hserver.core.interfaces.ProtocolDispatcherAdapter;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Order;
import cn.hserver.core.ioc.annotation.Value;
import cn.hserver.core.server.util.protocol.HostUtil;
import cn.hserver.plugin.web.protocol.DispatchHttp;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelPipeline;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.ssl.OptionalSslHandler;
import net.hserver.hplite.domian.bean.ConnectInfo;
import net.hserver.hplite.handler.proxy.FrontendHandler;
import net.hserver.hplite.handler.proxy.RouterHandler;
import net.hserver.hplite.handler.quic.QuicStreamHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;
import java.nio.ByteBuffer;

/**
 * 优先级要调整到自己的管理后台http协议的之上，都是http协议，所以这里需要判断是否是80
 */
@Order(0)
@Bean
public class HpWebProxyProtocolDispatcher implements ProtocolDispatcherAdapter {
    private static final Logger log = LoggerFactory.getLogger(HpWebProxyProtocolDispatcher.class);

    //判断HP头
    @Override
    public boolean dispatcher(ChannelHandlerContext ctx, ChannelPipeline channelPipeline, byte[] headers) {
        InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().localAddress();
        if (socketAddress.getPort() == 80 || socketAddress.getPort() == 443) {
            try {
                String host = HostUtil.getHost(ByteBuffer.wrap(headers));
                if (host != null) {
                    ConnectInfo connectInfo = QuicStreamHandler.getByDomain(host);
                    if (connectInfo == null) {
                        addErrorHandler(channelPipeline);
                    } else {
                        addProxyHandler(socketAddress.getPort() == 80, channelPipeline, host, connectInfo);
                    }
                    return true;
                }
            } catch (Exception e) {
                log.error(e.getMessage(), e);
                return false;
            }
        }
        return false;
    }

    /**
     * 未知来源的访问直接响应错误的映射
     *
     * @param pipeline
     */
    public void addErrorHandler(ChannelPipeline pipeline) {
        pipeline.addLast(new HttpServerCodec());
        pipeline.addLast(new RouterHandler());
    }

    /**
     * 存在反向代理
     *
     * @param host
     * @param pipeline
     */
    public void addProxyHandler(boolean hasHttp, ChannelPipeline pipeline, String host, ConnectInfo connectInfo) {
        pipeline.channel().config().setAutoRead(false);
        if (connectInfo.getSslContext() != null) {
            pipeline.addLast(new OptionalSslHandler(connectInfo.getSslContext()));
        }
        pipeline.addLast(new FrontendHandler(connectInfo, host));
    }


}
