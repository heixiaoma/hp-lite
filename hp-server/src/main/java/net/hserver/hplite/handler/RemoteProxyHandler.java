package net.hserver.hplite.handler;

import com.google.protobuf.ByteString;
import io.netty.channel.ChannelHandlerContext;
import net.hserver.hplite.handler.common.HpAbsHandler;
import net.hserver.hplite.message.HpMessageData;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;


/**
 * @author hxm
 */
public class RemoteProxyHandler extends HpAbsHandler {
    private static final Logger log = LoggerFactory.getLogger(RemoteProxyHandler.class);

    private final HpServerHandler proxyHandler;


    public RemoteProxyHandler(HpServerHandler proxyHandler) {
        this.proxyHandler = proxyHandler;
    }

    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        proxyHandler.getCtx().channel().config().setAutoRead(ctx.channel().isWritable());
        super.channelWritabilityChanged(ctx);
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        proxyHandler.getCtx().channel().config().setAutoRead(ctx.channel().isWritable());
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.CONNECTED);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.TCP).setChannelId(ctx.channel().id().asLongText()).build());
        proxyHandler.getCtx().writeAndFlush(messageBuilder.build());
    }


    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DISCONNECTED);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setChannelId(ctx.channel().id().asLongText()).build());
        HpMessageData.HpMessage build = messageBuilder.build();
        proxyHandler.getCtx().writeAndFlush(build);
    }


    /**
     * 用户穿透完成在外部创建的TCP服务，当有数据时进来，此时包装数据对象返回给内网客服端
     *
     * @param ctx
     * @param msg
     * @throws Exception
     */
    @Override
    public void channelRead0(ChannelHandlerContext ctx, byte[] msg) throws Exception {
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DATA);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.TCP).setChannelId(ctx.channel().id().asLongText()).build());
        messageBuilder.setData(ByteString.copyFrom(msg));
        proxyHandler.getCtx().writeAndFlush(messageBuilder.build());
    }
}
