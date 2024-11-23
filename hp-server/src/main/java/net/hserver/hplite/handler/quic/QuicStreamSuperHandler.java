package net.hserver.hplite.handler.quic;

import cn.hutool.json.JSONUtil;
import io.netty.channel.*;
import io.netty.channel.socket.ChannelInputShutdownReadComplete;
import io.netty.incubator.codec.quic.QuicChannel;
import io.netty.incubator.codec.quic.QuicStreamChannel;
import io.netty.incubator.codec.quic.QuicStreamType;
import io.netty.util.concurrent.Future;
import io.netty.util.concurrent.GenericFutureListener;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.ConnectInfo;
import net.hserver.hplite.handler.TunnelServer;
import net.hserver.hplite.init.QuicChannelInitializer;
import net.hserver.hplite.message.HpMessageData;

import java.net.InetSocketAddress;
import java.util.List;
import java.util.stream.Collectors;

import static net.hserver.hplite.handler.quic.QuicHandler.CURRENT_STATUS;


@Slf4j
public abstract class QuicStreamSuperHandler extends SimpleChannelInboundHandler<HpMessageData.HpMessage> {

    public ChannelHandlerContext channelHandlerContext;

    protected final Channel w;

    protected QuicStreamSuperHandler(Channel w) {
        this.w = w;
    }

    /**
     * 更具域名前缀匹配
     *
     * @param domain
     * @return
     */
    public static ConnectInfo getByDomain(String domain) {
        return CURRENT_STATUS.stream().filter(v -> domain != null && v != null && v.getDomain() != null && domain.equals(v.getDomain())).findFirst().orElse(null);
    }

    public static List<ConnectInfo> getByDomains(String domain) {
        return CURRENT_STATUS.stream().filter(v -> domain != null && v != null && v.getDomain() != null && domain.equals(v.getDomain())).collect(Collectors.toList());
    }

    public static List<ConnectInfo> getByKey(String key) {
        return CURRENT_STATUS.stream().filter(v -> key.equals(v.getKey())).collect(Collectors.toList());
    }

    public static List<ConnectInfo> getByPort(Integer port) {
        return CURRENT_STATUS.stream().filter(v -> port.equals(v.getPort())).collect(Collectors.toList());
    }


    /**
     * 链接端口关闭
     *
     * @param id
     * @return
     */
    public static List<ConnectInfo> getSuperChannelId(ChannelId id) {
        return CURRENT_STATUS.stream().filter(v -> v != null && v.getChannelId() != null && id.asLongText().equals(v.getChannelId().asLongText())).collect(Collectors.toList());
    }

    public void addConnectInfo(ConnectInfo connectInfo,boolean hasTcp) {
        String key = connectInfo.getKey();
        List<ConnectInfo> byDomain = getByKey(key);
        for (ConnectInfo info : byDomain) {
            TunnelServer tunnelServer = info.getTunnelServer();
            if (hasTcp==tunnelServer.getHasTcp()) {
                tunnelServer.close();
                QuicChannel quicChannel = getQuicChannel(info.getChannelId());
                if (quicChannel != null) {
                    quicChannel.close();
                }
                CURRENT_STATUS.remove(info);
            }
        }
        CURRENT_STATUS.add(connectInfo);
    }

    @Override
    public void channelReadComplete(ChannelHandlerContext ctx) throws Exception {
        ctx.flush(); // 确保所有待写入的数据都被刷新到远程节点
        super.channelReadComplete(ctx);
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        this.channelHandlerContext = ctx;
        if (w != null) {
            if (w.isActive()) {
                ctx.read();
            } else {
                w.close();
                ctx.channel().close();
            }
        } else {
            ctx.read();
        }
        super.channelActive(ctx);
    }


    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt == ChannelInputShutdownReadComplete.INSTANCE) {
            ctx.channel().close();
            if (w != null) {
                getCtxStream().shutdownOutput();
                w.close();
            }
        }
        super.userEventTriggered(ctx, evt);
    }


    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        ctx.channel().close();
        if (w != null) {
            getCtxStream().shutdownOutput();
            w.close();
        }
        super.channelInactive(ctx);
    }

    /**
     * 更具连接创建新的流传递
     *
     * @return
     */
    public QuicStreamChannel createStreamChannel(Channel w) {
        if (channelHandlerContext != null) {
            QuicChannel channel = (QuicChannel) QuicHandler.GROUP.find(channelHandlerContext.channel().parent().id());
            if (channel == null) {
                log.error(channelHandlerContext.channel().parent().id() + "-找不到ID");
                return null;
            }
            return createStreamChannel(channel, w);
        } else {
            return null;
        }
    }

    public QuicStreamChannel createStreamChannel(QuicChannel channel, Channel w) {
        if (channel == null || !channel.isActive() || !channel.isOpen() || !channel.isWritable()) {
            log.error("创建流失败：为空，未激活，未打开,不可读写");
            if (channel != null) {
                log.error("激活:" + channel.isActive() + ",打开:" + channel.isOpen() + ",读写:" + channel.isWritable());
            }
            return null;
        }
        try {
            Future<QuicStreamChannel> stream = channel.createStream(QuicStreamType.BIDIRECTIONAL, new QuicChannelInitializer(w));
            return stream.sync().getNow();
        } catch (Exception e) {
            log.error("创建流失败：" + e.getMessage(), e);
            return null;
        }
    }

    public QuicChannel getQuicChannel(ChannelId channelId) {
        return (QuicChannel) QuicHandler.GROUP.find(channelId);
    }

    /**
     * 发送数据
     *
     * @param hpMessage
     */
    public void sendMessage(HpMessageData.HpMessage hpMessage, GenericFutureListener<? extends Future<? super Void>> var1) {
        QuicStreamChannel streamChannel = createStreamChannel(null);
        if (streamChannel != null) {
            ChannelFuture channelFuture = streamChannel.writeAndFlush(hpMessage);
            if (var1 != null) {
                channelFuture.addListener(var1);
            }
        }
    }

    /**
     * 获取远程的IP地址
     *
     * @return
     */
    public InetSocketAddress getIpAddress() {
        QuicChannel channel = (QuicChannel) QuicHandler.GROUP.find(channelHandlerContext.channel().parent().id());
        return channel.attr(QuicHandler.IP_ADDRESS).get();
    }

    public QuicChannel getSuperChannel() {
        return (QuicChannel) channelHandlerContext.channel().parent();
    }

    public ChannelId getSuperChannelId() {
        return channelHandlerContext.channel().parent().id();
    }

    public QuicStreamChannel getCtxStream() {
        return (QuicStreamChannel) channelHandlerContext.channel();
    }


}
