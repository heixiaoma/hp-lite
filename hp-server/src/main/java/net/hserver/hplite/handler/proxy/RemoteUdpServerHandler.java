package net.hserver.hplite.handler.proxy;

import com.google.protobuf.ByteString;
import io.netty.buffer.ByteBufUtil;
import io.netty.channel.ChannelConfig;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.channel.socket.DatagramPacket;
import io.netty.util.Attribute;
import io.netty.util.AttributeKey;
import net.hserver.hplite.handler.TunnelServer;
import net.hserver.hplite.handler.common.HpCommonHandler;
import net.hserver.hplite.message.HpMessageData;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;

public class RemoteUdpServerHandler extends
        SimpleChannelInboundHandler<DatagramPacket> {
    private static final Logger log = LoggerFactory.getLogger(RemoteUdpServerHandler.class);

    private final TunnelServer tunnelServer;

    private final HpCommonHandler proxyHandler;

    public static final AttributeKey<InetSocketAddress> SENDER = AttributeKey.valueOf("Sender");

    public RemoteUdpServerHandler(HpCommonHandler proxyHandler, TunnelServer tunnelServer) {
        this.tunnelServer = tunnelServer;
        this.proxyHandler = proxyHandler;
    }

    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        ChannelConfig config = proxyHandler.getCtx().channel().config();
        //自己不可写，通道可以读，让通道关闭读
        //自己可写，通道不可以读，让通道打开读
        if (!ctx.channel().isWritable() && config.isAutoRead()) {
            config.setAutoRead(false);
        } else if (ctx.channel().isWritable() && !config.isAutoRead()) {
            config.setAutoRead(true);
        }
        super.channelWritabilityChanged(ctx);
    }


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause)
            throws Exception {
        log.error("UDP", cause);
        ctx.close();
    }

    /**
     * 外网回来的数据通道被激活通知内网服务器建立客服端
     *
     * @param ctx
     * @throws Exception
     */
    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        proxyHandler.getCtx().channel().config().setAutoRead(ctx.channel().isWritable());
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.CONNECTED);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.UDP).setChannelId(ctx.channel().id().asLongText()).build());
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
     * 外网的UDP数据返回到内网去 由内网客服端去转发
     *
     * @param ctx
     * @param msg
     * @throws Exception
     */
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, DatagramPacket msg) throws Exception {
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DATA);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.UDP).setChannelId(ctx.channel().id().asLongText()).build());
        messageBuilder.setData(ByteString.copyFrom(ByteBufUtil.getBytes(msg.content())));
        proxyHandler.getCtx().writeAndFlush(messageBuilder.build());
        final Attribute<InetSocketAddress> attr = ctx.channel().attr(SENDER);
        attr.set(msg.sender());
    }
}
