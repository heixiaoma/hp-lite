package net.hserver.hplite.handler.quic;

import cn.hserver.core.queue.HServerQueue;
import cn.hserver.core.server.util.ReleaseUtil;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.Unpooled;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.group.ChannelGroup;
import io.netty.channel.group.DefaultChannelGroup;
import io.netty.channel.socket.DatagramPacket;
import io.netty.util.Attribute;
import io.netty.util.AttributeKey;
import io.netty.util.concurrent.GlobalEventExecutor;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.config.CostConfig;
import net.hserver.hplite.handler.RemoteUdpServerHandler;
import net.hserver.hplite.message.HpMessageData;
import net.hserver.hplite.queue.QueueEvent;

import java.net.InetSocketAddress;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;


@Slf4j
public class QuicStreamHandler extends QuicStreamSuperHandler {
    public static final AttributeKey<String> KEY = AttributeKey.valueOf("KEY");
    public final static Map<String, QuicStreamHandler> DATA_CON = new ConcurrentHashMap<>();
    public static final ChannelGroup DATA_CON_CH = new DefaultChannelGroup(GlobalEventExecutor.INSTANCE);

    public QuicStreamHandler(Channel w) {
        super(w);
    }


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        log.error(cause.getMessage(), cause);
        String s = getCtxStream().attr(KEY).get();
        DATA_CON.remove(s);
        getCtxStream().close();
        if (w!=null){
            w.close();
        }
    }

    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        Channel channel = ctx.channel();
        if (channel.isWritable()){
            Boolean andSet = channel.attr(CostConfig.flow).getAndSet(null);
            if (andSet!=null&&andSet){
                if (w.isOpen()) {
                    w.read();
                }
            }
        }
        super.channelWritabilityChanged(ctx);
    }
    @Override
    protected void channelRead0(ChannelHandlerContext ctx, HpMessageData.HpMessage hpMessage) throws Exception {
        if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.REGISTER) {
            String key = hpMessage.getMetaData().getKey();
            QuicStreamHandler quicStreamHandler = DATA_CON.get(key);
            if (quicStreamHandler != null) {
                quicStreamHandler.channelHandlerContext.close();
                log.warn("存在KEY:{},IP:{},大小:{}", key, getIpAddress(), DATA_CON.size());
                DATA_CON.remove(key);
            }
            getCtxStream().attr(KEY).set(key);
            String cid = getCtxStream().id().asLongText();
            DATA_CON_CH.add(getCtxStream());
            DATA_CON.put(key, this);
            //异步处理,上面做了双层保险
            HServerQueue.sendQueue(QueueEvent.CONN_EVENT, hpMessage, cid);
            safeRead(false);
        } else if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.DISCONNECTED) {
            processDisconnected(hpMessage);
        } else if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.DATA) {
            processData(hpMessage);
        } else if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.KEEPALIVE) {
           safeRead(false);
        } else {
            log.error("未知类型: " + hpMessage.getType());
            getCtxStream().close();
        }
    }


    private void processDisconnected(HpMessageData.HpMessage hpMessage) {
        if (w != null) {
            w.close();
            getCtxStream().close();
        }
    }


    private void processData(HpMessageData.HpMessage hpMessage) {
        if (w == null || !w.isOpen() || !w.isActive()) {
            log.error("找不到关联：" + hpMessage.getMetaData().getChannelId());
            if (w!=null){
                w.close();
            }
            getCtxStream().close();
            return;
        }
        if (!w.isActive() || !w.isOpen()) {
            getCtxStream().close();
            return;
        }
        safeRead(true);
        //检查读写
        byte[] bytes = hpMessage.getData().toByteArray();
        ByteBuf byteBuf = Unpooled.wrappedBuffer(bytes);
        if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.TCP) {
                 w.writeAndFlush(byteBuf).addListener((ChannelFutureListener)future -> {
                     ReleaseUtil.release(byteBuf);
                    if (future.isSuccess()) {
                       safeRead(false);
                    } else {
                        future.channel().close();
                    }
                });
        }
        if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.UDP) {
            final Attribute<InetSocketAddress> attr = w.attr(RemoteUdpServerHandler.SENDER);
            final InetSocketAddress inetSocketAddress = attr.get();
            if (inetSocketAddress != null) {
                try {
                    w.writeAndFlush(new DatagramPacket(byteBuf, inetSocketAddress)).addListener((ChannelFutureListener)future -> {
                        ReleaseUtil.release(byteBuf);
                        if (future.isSuccess()) {
                           safeRead(false);
                        } else {
                            future.channel().close();
                        }
                    });
                }catch (Exception e){
                    log.error(e.getMessage(), e);
                }
            }
        }
    }

    public void safeRead(boolean checkRead){
        if (w==null){
            getCtxStream().read();
            return;
        }
        if (w.isOpen()&&getCtxStream().isOpen()) {
            Boolean aBoolean = w.attr(CostConfig.flow).get();
            if (aBoolean!=null){
                return;
            }
            if (w.isWritable()) {
                if (!checkRead) {
                    getCtxStream().read();
                }
            } else {
                w.attr(CostConfig.flow).set(true);
            }
        }
    }
}
