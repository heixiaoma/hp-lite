package net.hserver.hplite.handler;

import com.google.protobuf.ByteString;
import io.netty.buffer.ByteBufUtil;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.channel.socket.DatagramPacket;
import io.netty.incubator.codec.quic.QuicStreamChannel;
import io.netty.util.Attribute;
import io.netty.util.AttributeKey;
import net.hserver.hplite.config.CostConfig;
import net.hserver.hplite.handler.quic.QuicStreamHandler;
import net.hserver.hplite.message.HpMessageData;
import net.hserver.hplite.message.UserConnectInfo;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;


public class RemoteUdpServerHandler extends
        SimpleChannelInboundHandler<DatagramPacket> {
    private static final Logger log = LoggerFactory.getLogger(RemoteUdpServerHandler.class);

    private QuicStreamChannel n;

    private ChannelHandlerContext w;

    private final QuicStreamHandler proxyHandler;
    private final UserConnectInfo userConnectInfo;
    public static final AttributeKey<InetSocketAddress> SENDER = AttributeKey.valueOf("Sender");

    public RemoteUdpServerHandler(QuicStreamHandler proxyHandler, UserConnectInfo userConnectInfo) {
        this.proxyHandler = proxyHandler;
        this.userConnectInfo = userConnectInfo;
    }



    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause)
            throws Exception {
        log.error("UDP", cause);
        ctx.channel().close();
        if (n!=null){
            n.shutdownInput();
            n.shutdownOutput();
            n.shutdown();
            n.close();
        }
    }


    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        Channel channel = ctx.channel();
        if (channel.isWritable()){
            Boolean andSet = channel.attr(CostConfig.flow).getAndSet(null);
            if (andSet!=null&&andSet){
                if (n.isOpen()) {
                    n.read();
                }
            }
        }
        super.channelWritabilityChanged(ctx);
    }


    /**
     * 外网回来的数据通道被激活通知内网服务器建立客服端
     *
     * @param ctx
     * @throws Exception
     */
    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        if (!ctx.channel().isOpen()){
            return;
        }
        this.w=ctx;
        String id = ctx.channel().id().asLongText();
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.CONNECTED);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.UDP).setChannelId(id).build());
        //建立内网双向通道
        n = proxyHandler.createStreamChannel(ctx.channel());
        if (n != null) {
            n.writeAndFlush(messageBuilder.build()).addListener(future -> {
               if (future.isSuccess()){
                    safeRead(false);
               }else {
                   ctx.channel().close();
               }
            });
        } else {
            ctx.channel().close();
        }
        super.channelActive(ctx);
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        if (n != null) {
            String id = ctx.channel().id().asLongText();
            HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
            messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setChannelId(id).build());
            messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DISCONNECTED);
            HpMessageData.HpMessage build = messageBuilder.build();
            n.writeAndFlush(build).addListener(future -> {
                n.shutdownOutput();
                n.close();
            });
        }
        ctx.channel().close();
        super.channelInactive(ctx);
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
        if (!ctx.channel().isOpen()){
            return;
        }
        String id = ctx.channel().id().asLongText();
        if (n == null || !n.isOpen()) {
            log.error("获取不到映射关系:" + id);
            n.close();
            ctx.channel().close();
            return;
        }
        if (!n.isActive() || !n.isOpen()||!ctx.channel().isActive()) {
            ctx.channel().close();
            return;
        }
        safeRead(true);
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DATA);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.UDP).setChannelId(id).build());
        messageBuilder.setData(ByteString.copyFrom(ByteBufUtil.getBytes(msg.content())));
        n.writeAndFlush(messageBuilder.build()).addListener((ChannelFutureListener)future -> {
            if (future.isSuccess()) {
                safeRead(false);
            } else {
                future.channel().close();
            }
        });
        final Attribute<InetSocketAddress> attr = ctx.channel().attr(SENDER);
        attr.set(msg.sender());
    }

    @Override
    public void channelReadComplete(ChannelHandlerContext ctx) throws Exception {
        ctx.flush(); // 确保所有待写入的数据都被刷新到远程节点
        super.channelReadComplete(ctx);
    }


    public void safeRead(boolean checkRead){
        if (n.isOpen()&&w.channel().isOpen()) {
            Boolean aBoolean = n.attr(CostConfig.flow).get();
            if (aBoolean!=null){
                return;
            }
            if (n.isWritable()) {
                if (!checkRead) {
                    w.channel().read();
                }
            } else {
                n.attr(CostConfig.flow).set(true);
            }
        }
    }

}
