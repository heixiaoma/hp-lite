package net.hserver.hplite.handler.common;

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
public abstract class HpAbsHandler extends SimpleChannelInboundHandler<byte[]> {
    private static final Logger log = LoggerFactory.getLogger(HpAbsHandler.class);

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        if (!(cause instanceof IOException)){
            log.error("外网TCP ......",cause);
        }
        ctx.close();
    }

    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof IdleStateEvent) {
            IdleStateEvent e = (IdleStateEvent) evt;
            //不可写说明在传输数据，不要傻逼的给关了
            if (e.state() == IdleState.READER_IDLE&&ctx.channel().isWritable()) {
                System.out.println("Read idle loss connection....");
                ctx.close();
            } else if (e.state() == IdleState.WRITER_IDLE) {
                HpMessageData.HpMessage.Builder messageBuild = HpMessageData.HpMessage.newBuilder();
                messageBuild.setType(HpMessageData.HpMessage.HpMessageType.KEEPALIVE);
                ctx.writeAndFlush(messageBuild.build());
            }
        }
    }
}
