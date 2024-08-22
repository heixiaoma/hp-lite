package net.hserver.hplite.handler;

import cn.hutool.core.util.StrUtil;
import com.google.protobuf.ByteString;
import io.netty.buffer.ByteBuf;
import io.netty.buffer.ByteBufUtil;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.haproxy.HAProxyCommand;
import io.netty.handler.codec.haproxy.HAProxyMessage;
import io.netty.handler.codec.haproxy.HAProxyProtocolVersion;
import io.netty.handler.codec.haproxy.HAProxyProxiedProtocol;
import io.netty.incubator.codec.quic.QuicStreamChannel;
import net.hserver.hplite.config.CostConfig;
import net.hserver.hplite.handler.quic.QuicStreamHandler;
import net.hserver.hplite.message.HpMessageData;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.utils.HAProxyMessageUtil;
import net.hserver.hplite.utils.NetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;


/**
 * @author hxm
 */
public class RemoteProxyHandler extends SimpleChannelInboundHandler<Object> {
    private static final Logger log = LoggerFactory.getLogger(RemoteProxyHandler.class);
    private final QuicStreamHandler proxyHandler;
    private final UserConnectInfo userConnectInfo;

    private QuicStreamChannel n;
    private ChannelHandlerContext w;

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().remoteAddress();
        log.error("外网TCP_IP:" + ctx.channel().remoteAddress() + "内网:" + ctx.channel().localAddress(), cause);
        ctx.channel().close();
    }

    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        Channel channel = ctx.channel();
        if (channel.isWritable()) {
            Boolean andSet = channel.attr(CostConfig.flow).getAndSet(null);
            if (andSet != null && andSet) {
                if (n.isOpen()) {
                    n.read();
                }
            }
        }
        super.channelWritabilityChanged(ctx);
    }

    public RemoteProxyHandler(QuicStreamHandler proxyHandler, UserConnectInfo userConnectInfo) {
        this.proxyHandler = proxyHandler;
        this.userConnectInfo = userConnectInfo;
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        if (!ctx.channel().isOpen()){
            return;
        }
        this.w = ctx;
        String id = ctx.channel().id().asLongText();
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.CONNECTED);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.TCP).setChannelId(id).build());
        //建立内网双向通道
        n = proxyHandler.createStreamChannel(ctx.channel());
        if (n != null) {
            n.writeAndFlush(messageBuilder.build()).addListener(future -> {
                if (future.isSuccess()) {
                    InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().remoteAddress();
                    //如果是内网地址说明是走代理发送过来的，他已经发送过这个协议了，所以我们要检查是否走http代理过来，部署http代理过的，就是直接tcp访问的，我们要补IP协议
                    if (StrUtil.isNotEmpty(userConnectInfo.getProxyVersion()) && !NetUtil.isPrivateAddress(socketAddress)) {
                        //发送原始IP数据包
                        HAProxyMessage message = new HAProxyMessage(
                                HAProxyProtocolVersion.valueOf(userConnectInfo.getProxyVersion()), HAProxyCommand.PROXY, HAProxyProxiedProtocol.TCP4,
                                socketAddress.getHostString(), userConnectInfo.getProxyIp(), socketAddress.getPort(), userConnectInfo.getProxyPort());
                        byte[] bytes = HAProxyMessageUtil.encodeBytes(message);
                        HpMessageData.HpMessage.Builder messageBuilder2 = HpMessageData.HpMessage.newBuilder();
                        messageBuilder2.setType(HpMessageData.HpMessage.HpMessageType.DATA);
                        messageBuilder2.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.TCP).setChannelId(id).build());
                        messageBuilder2.setData(ByteString.copyFrom(bytes));
                        n.writeAndFlush(messageBuilder2.build()).addListener(future1 -> {
                            if (future1.isSuccess()) {
                                safeRead(false);
                            } else {
                                ctx.channel().close();
                            }
                        });
                    } else {
                        safeRead(false);
                    }
                } else {
                    ctx.channel().close();
                    log.error(future.cause().getMessage(), future.cause());
                }
            });
        } else {
            ctx.channel().close();
            log.error("获取失败");
        }
        super.channelActive(ctx);
    }

    /**
     * 用户映射完成在外部创建的TCP服务，当有数据时进来，此时包装数据对象返回给内网客服端
     *
     * @param ctx
     * @param msg
     * @throws Exception
     */
    @Override
    public void channelRead0(ChannelHandlerContext ctx, Object msg) throws Exception {
        if (!ctx.channel().isOpen()){
            return;
        }
        String id = ctx.channel().id().asLongText();
        if (n == null || !n.isOpen()) {
            log.error("内网不活跃");
            n.close();
            ctx.channel().close();
            return;
        }
        if (!n.isActive() || !n.isOpen() || !ctx.channel().isActive()) {
            ctx.channel().close();
            return;
        }
        safeRead(true);
        HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
        messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DATA);
        messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setType(HpMessageData.HpMessage.MessageType.TCP).setChannelId(id).build());
        messageBuilder.setData(ByteString.copyFrom(ByteBufUtil.getBytes((ByteBuf) msg)));
        n.writeAndFlush(messageBuilder.build()).addListener((ChannelFutureListener) future -> {
            if (future.isSuccess()) {
                safeRead(false);
            } else {
                future.channel().close();
            }
        });
    }

    @Override
    public void channelReadComplete(ChannelHandlerContext ctx) throws Exception {
        ctx.flush(); // 确保所有待写入的数据都被刷新到远程节点
        super.channelReadComplete(ctx);
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        String id = ctx.channel().id().asLongText();
        if (n != null) {
            HpMessageData.HpMessage.Builder messageBuilder = HpMessageData.HpMessage.newBuilder();
            messageBuilder.setType(HpMessageData.HpMessage.HpMessageType.DISCONNECTED);
            messageBuilder.setMetaData(HpMessageData.HpMessage.MetaData.newBuilder().setChannelId(id).build());
            HpMessageData.HpMessage build = messageBuilder.build();
            n.writeAndFlush(build).addListener(future -> {
                n.shutdownOutput();
                n.close();
            });
        }
        ctx.channel().close();
        super.channelInactive(ctx);
    }


    public void safeRead(boolean checkRead) {
        if (n.isOpen() && w.channel().isOpen()) {
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
