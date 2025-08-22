package net.hserver.hplite.handler.proxy;

import cn.hserver.core.server.util.ReleaseUtil;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import net.hserver.hplite.config.CostConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class BackendHandler extends ChannelInboundHandlerAdapter {
    private static final Logger log = LoggerFactory.getLogger(BackendHandler.class);

    private final Channel inboundChannel;
    private  Channel outboundChannel;

    public BackendHandler(Channel inboundChannel) {
        this.inboundChannel = inboundChannel;
    }


    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        Channel channel = ctx.channel();
        if (channel.isWritable()){
            Boolean andSet = channel.attr(CostConfig.flow).getAndSet(null);
            if (andSet!=null&&andSet){
                if (inboundChannel.isOpen()) {
                    inboundChannel.read();
                }
            }
        }
        super.channelWritabilityChanged(ctx);
    }

    public void safeRead(){
        if (inboundChannel.isOpen() && outboundChannel.isOpen()) {
            Boolean aBoolean = inboundChannel.attr(CostConfig.flow).get();
            if (aBoolean!=null){
                return;
            }
            if (inboundChannel.isWritable()) {
                outboundChannel.read();
            } else {
                inboundChannel.attr(CostConfig.flow).set(true);
            }
        }
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        this.outboundChannel=ctx.channel();
        if (!inboundChannel.isActive()||!inboundChannel.isOpen()){
            FrontendHandler.closeOnFlush(inboundChannel);
        }else {
            safeRead();
        }
        super.channelActive(ctx);
    }

    @Override
    public void channelReadComplete(ChannelHandlerContext ctx) throws Exception {
        ctx.flush(); // 确保所有待写入的数据都被刷新到远程节点
        super.channelReadComplete(ctx);
    }

    @Override
    public void channelRead(final ChannelHandlerContext ctx, Object msg) {
        inboundChannel.writeAndFlush(msg).addListener((ChannelFutureListener) future -> {
            if (!future.isSuccess()) {
                future.channel().close();
                inboundChannel.close();
                ReleaseUtil.release(msg);
            }else {
                safeRead();
            }
        });
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) {
        FrontendHandler.closeOnFlush(inboundChannel);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        log.error("WEB代理反向写", cause);
        FrontendHandler.closeOnFlush(ctx.channel());
    }
}

