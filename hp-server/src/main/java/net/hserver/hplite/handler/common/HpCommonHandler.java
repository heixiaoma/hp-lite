package net.hserver.hplite.handler.common;

import cn.hserver.core.server.util.ExceptionUtil;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.timeout.IdleState;
import io.netty.handler.timeout.IdleStateEvent;
import net.hserver.hplite.message.HpMessageData;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;

/**
 * @author hxm
 */
public abstract class HpCommonHandler extends SimpleChannelInboundHandler<HpMessageData.HpMessage> {
    private static final Logger log = LoggerFactory.getLogger(HpCommonHandler.class);

    protected ChannelHandlerContext ctx;

    public ChannelHandlerContext getCtx() {
        return ctx;
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        this.ctx = ctx;
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        if (!(cause instanceof IOException)){
            cause.printStackTrace();
            log.error("HP通道 {}......\n{}",cause.getMessage(), ExceptionUtil.getMessage(cause));
        }
        ctx.close();
    }

    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof IdleStateEvent) {
            IdleStateEvent e = (IdleStateEvent) evt;
            //如果数据堆积情况，不能关闭连接，
            if (e.state() == IdleState.READER_IDLE&&ctx.channel().isWritable()) {
                System.out.println("Read idle loss connection.");
                ctx.close();
            } else if (e.state() == IdleState.WRITER_IDLE) {
                HpMessageData.HpMessage keepMessage = HpMessageData.HpMessage.newBuilder().setType(HpMessageData.HpMessage.HpMessageType.KEEPALIVE).build();
                ctx.writeAndFlush(keepMessage);
            }
        }
    }
}
