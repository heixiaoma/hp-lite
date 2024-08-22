package net.hserver.hplite.handler.proxy;

import cn.hserver.core.server.util.EventLoopUtil;
import cn.hserver.core.server.util.ReleaseUtil;
import cn.hutool.core.util.StrUtil;
import io.netty.bootstrap.Bootstrap;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.handler.codec.haproxy.HAProxyCommand;
import io.netty.handler.codec.haproxy.HAProxyMessage;
import io.netty.handler.codec.haproxy.HAProxyProtocolVersion;
import io.netty.handler.codec.haproxy.HAProxyProxiedProtocol;
import net.hserver.hplite.config.CostConfig;
import net.hserver.hplite.domian.bean.ConnectInfo;
import net.hserver.hplite.utils.HAProxyMessageUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import javax.net.ssl.SSLHandshakeException;
import java.net.InetSocketAddress;

public class FrontendHandler extends ChannelInboundHandlerAdapter {
    private static final Logger log = LoggerFactory.getLogger(FrontendHandler.class);
    private static final EventLoopGroup eventLoop = EventLoopUtil.getEventLoop(20, "HP-WEB-PROXY");

    private Channel outboundChannel;
    private Channel inboundChannel;
    private final ConnectInfo connectInfo;
    private final String host;

    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        Channel channel = ctx.channel();
        if (channel.isWritable()){
            Boolean andSet = channel.attr(CostConfig.flow).getAndSet(null);
            if (andSet!=null&&andSet){
                if (outboundChannel.isOpen()) {
                    outboundChannel.read();
                }
            }
        }
        super.channelWritabilityChanged(ctx);
    }

    public void safeRead(){
        if (inboundChannel.isOpen() && outboundChannel.isOpen()) {
            Boolean aBoolean = outboundChannel.attr(CostConfig.flow).get();
            if (aBoolean!=null){
                return;
            }
            if (outboundChannel.isWritable()) {
                inboundChannel.read();
            } else {
                outboundChannel.attr(CostConfig.flow).set(true);
            }
        }
    }

    public FrontendHandler(ConnectInfo connectInfo, String host) {
        this.connectInfo = connectInfo;
        this.host = host;
    }

    static void closeOnFlush(Channel ch) {
        if (ch.isActive()) {
            ch.writeAndFlush(Unpooled.EMPTY_BUFFER).addListener(ChannelFutureListener.CLOSE);
        }
    }

    public void write(ChannelHandlerContext ctx, Object msg) {
        outboundChannel.writeAndFlush(msg).addListener((ChannelFutureListener) future -> {
            if (future == null) {
                ReleaseUtil.release(msg);
                return;
            }
            if (!future.isSuccess()) {
                log.error(future.cause().getMessage(), future.cause());
                future.channel().close();
                ReleaseUtil.release(msg);
            } else {
                safeRead();
            }
        });
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        this.inboundChannel = ctx.channel();
        if (!inboundChannel.isOpen()){
            return;
        }
        Bootstrap b = new Bootstrap();
        b.group(eventLoop)
                .option(ChannelOption.AUTO_READ, false)
                .channel(EventLoopUtil.getEventLoopTypeClassClient())
                .handler(new ChannelInitializer<Channel>() {
                    @Override
                    protected void initChannel(Channel ch) throws Exception {
                        ch.pipeline().addLast(new BackendHandler(inboundChannel));
                    }
                });
        ChannelFuture sync = b.connect("127.0.0.1", connectInfo.getPort()).sync();
        if (sync.isSuccess()) {
            outboundChannel = sync.channel();
            if (StrUtil.isNotEmpty(connectInfo.getProxyVersion())) {
                //发送原始IP数据包
                InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().remoteAddress();
                HAProxyMessage message = new HAProxyMessage(
                        HAProxyProtocolVersion.valueOf(connectInfo.getProxyVersion()), HAProxyCommand.PROXY, HAProxyProxiedProtocol.TCP4,
                        socketAddress.getHostString(), connectInfo.getProxyIp(), socketAddress.getPort(), connectInfo.getProxyPort());
                ByteBuf byteBuf = HAProxyMessageUtil.encodeByteBuf(message);
                ChannelFuture sync1 = outboundChannel.writeAndFlush(byteBuf).sync();
                if (sync1.isSuccess()) {
                    safeRead();
                } else {
                    inboundChannel.close();
                }
                ReleaseUtil.release(byteBuf);
            } else {
                safeRead();
            }
        } else {
            inboundChannel.close();
        }

    }

    @Override
    public void channelReadComplete(ChannelHandlerContext ctx) throws Exception {
        ctx.flush(); // 确保所有待写入的数据都被刷新到远程节点
        super.channelReadComplete(ctx);
    }

    @Override
    public void channelRead(final ChannelHandlerContext ctx, Object msg) {
        if (outboundChannel == null) {
            ctx.close();
            ReleaseUtil.release(msg);
            return;
        }
        if (outboundChannel.isActive()) {
            write(ctx, msg);
        } else {
            outboundChannel.close();
            ctx.channel().close();
            ReleaseUtil.release(msg);
        }
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        if (outboundChannel != null) {
            closeOnFlush(outboundChannel);
        }
        ctx.fireChannelInactive();
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        if (cause instanceof SSLHandshakeException) {
            closeOnFlush(ctx.channel());
            return;
        }
        InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().remoteAddress();
        log.error("WEB通道 ...[" + host +"--"+ socketAddress + "]..." + cause.getMessage(), cause);
        closeOnFlush(ctx.channel());
    }
}

