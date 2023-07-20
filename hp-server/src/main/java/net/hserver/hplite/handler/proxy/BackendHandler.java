package net.hserver.hplite.handler.proxy;

import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelInboundHandlerAdapter;
import io.netty.util.ReferenceCountUtil;
import io.netty.util.ReferenceCounted;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class BackendHandler extends ChannelInboundHandlerAdapter {
    private static final Logger log = LoggerFactory.getLogger(BackendHandler.class);

    private final Channel inboundChannel;

    public BackendHandler(Channel inboundChannel) {
        this.inboundChannel = inboundChannel;
    }


    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        inboundChannel.config().setAutoRead(ctx.channel().isWritable());
        super.channelWritabilityChanged(ctx);
    }

    @Override
    public void channelRead(final ChannelHandlerContext ctx, Object msg) {
        inboundChannel.writeAndFlush(msg).addListener((ChannelFutureListener) future -> {
            if (!future.isSuccess()) {
                future.channel().close();
                inboundChannel.close();
                if (msg == ReferenceCounted.class) {
                    ReferenceCounted msg1 = (ReferenceCounted) msg;
                    if (msg1.refCnt() > 0) {
                        if (((ReferenceCounted) msg).refCnt() > 0) {
                            ReferenceCountUtil.release(msg);
                        }
                    }
                }
            }
        });
    }

        @Override
        public void channelInactive (ChannelHandlerContext ctx){
            FrontendHandler.closeOnFlush(inboundChannel);
        }

        @Override
        public void exceptionCaught (ChannelHandlerContext ctx, Throwable cause){
            log.error("WEB代理反向写", cause);
            FrontendHandler.closeOnFlush(ctx.channel());
        }
}

