package net.hserver.hplite.handler.quic;

import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import io.netty.channel.group.ChannelGroup;
import io.netty.channel.group.DefaultChannelGroup;
import io.netty.incubator.codec.quic.QuicChannel;
import io.netty.incubator.codec.quic.QuicPathEvent;
import io.netty.util.Attribute;
import io.netty.util.AttributeKey;
import io.netty.util.concurrent.GlobalEventExecutor;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.ConnectInfo;

import java.net.InetSocketAddress;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;


@Slf4j
public class QuicHandler extends ChannelInboundHandlerAdapter {
    public final static AttributeKey<InetSocketAddress> IP_ADDRESS = AttributeKey.valueOf("IP_ADDRESS");
    /**
     * QUIC链接信息
     */
    public static final List<ConnectInfo> CURRENT_STATUS = new CopyOnWriteArrayList<>();

    public static final ChannelGroup GROUP = new DefaultChannelGroup(GlobalEventExecutor.INSTANCE);

    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        super.channelWritabilityChanged(ctx);
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) {
        QuicChannel channel = (QuicChannel) ctx.channel();
        InetSocketAddress o = (InetSocketAddress) channel.remoteSocketAddress();
        ctx.channel().attr(IP_ADDRESS).set(o);
        GROUP.add(channel);
    }

    /**
     * quic基于UDP 远程IP随时都在变化
     *
     * @param ctx
     * @param evt
     * @throws Exception
     */
    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof QuicPathEvent) {
            QuicPathEvent event = (QuicPathEvent) evt;
            ctx.channel().attr(IP_ADDRESS).set(event.remote());
        }
    }

    public void channelInactive(ChannelHandlerContext ctx) {
        ((QuicChannel) ctx.channel()).collectStats().addListener(f -> {
            if (f.isSuccess()) {
                List<ConnectInfo> superChannelId = QuicStreamSuperHandler.getSuperChannelId(ctx.channel().id());
                if (superChannelId != null && !superChannelId.isEmpty()) {
                    for (ConnectInfo connectInfo : superChannelId) {
                        log.info("连接关闭: 域名 {},端口 {}",connectInfo.getDomain(), connectInfo.getPort());
                        connectInfo.getTunnelServer().close();
                    }
                    CURRENT_STATUS.removeAll(superChannelId);
                }
                Attribute<InetSocketAddress> attr = ctx.channel().attr(IP_ADDRESS);
                InetSocketAddress address = attr.get();
                log.info("连接关闭:{},链接信息长度: {}", address, CURRENT_STATUS.size());
            }
        });
    }


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        List<ConnectInfo> superChannelId = QuicStreamSuperHandler.getSuperChannelId(ctx.channel().id());
        Attribute<InetSocketAddress> attr = ctx.channel().attr(IP_ADDRESS);
        InetSocketAddress address = attr.get();
        if (superChannelId != null && !superChannelId.isEmpty()) {
            for (ConnectInfo connectInfo : superChannelId) {
                log.error("连接异常关闭: 远端IP:{},IP {},域名 {},端口 {}", address, connectInfo.getIp(), connectInfo.getDomain(), connectInfo.getPort());
            }
        }
        log.error(address.toString() + "=>" + cause.getMessage(), cause);
        ctx.close();
    }

    @Override
    public boolean isSharable() {
        return true;
    }
}
