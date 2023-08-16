package net.hserver.hplite.handler;

import cn.hserver.core.ioc.IocUtil;
import io.netty.buffer.Unpooled;
import io.netty.channel.*;
import io.netty.channel.group.ChannelGroup;
import io.netty.channel.group.DefaultChannelGroup;
import io.netty.channel.socket.DatagramPacket;
import io.netty.channel.socket.SocketChannel;
import io.netty.handler.codec.bytes.ByteArrayDecoder;
import io.netty.handler.codec.bytes.ByteArrayEncoder;
import io.netty.util.Attribute;
import io.netty.util.concurrent.GlobalEventExecutor;
import net.hserver.hplite.config.WebConfig;
import net.hserver.hplite.domian.bean.ConnectInfo;
import net.hserver.hplite.handler.common.HpCommonHandler;
import net.hserver.hplite.handler.proxy.RemoteUdpServerHandler;
import net.hserver.hplite.message.HpMessageData;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.service.HttpService;
import net.hserver.hplite.utils.NetUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.InetSocketAddress;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

/**
 * @author hxm
 */
@ChannelHandler.Sharable
public class HpServerHandler extends HpCommonHandler {
    private static final Logger log = LoggerFactory.getLogger(HpServerHandler.class);
    private final TunnelServer remoteConnectionServer = new TunnelServer();
    public static final List<ConnectInfo> CURRENT_STATUS = new ArrayList<>();

    public final ChannelGroup channels = new DefaultChannelGroup(GlobalEventExecutor.INSTANCE);

    private final ChannelGroup udp_channels = new DefaultChannelGroup(GlobalEventExecutor.INSTANCE);

    public HpServerHandler() {
    }


    @Override
    public void channelWritabilityChanged(ChannelHandlerContext ctx) throws Exception {
        channels.forEach(targetChannel -> {
            //自己不可写，通道可以读，让通道关闭读
            //自己可写，通道不可以读，让通道打开读
            if (!ctx.channel().isWritable() && targetChannel.config().isAutoRead()) {
                targetChannel.config().setAutoRead(false);
            } else if (ctx.channel().isWritable() && !targetChannel.config().isAutoRead()) {
                targetChannel.config().setAutoRead(true);
            }
        });
        udp_channels.forEach(targetChannel -> {
            //自己不可写，通道可以读，让通道关闭读
            //自己可写，通道不可以读，让通道打开读
            if (!ctx.channel().isWritable() && targetChannel.config().isAutoRead()) {
                targetChannel.config().setAutoRead(false);
            } else if (ctx.channel().isWritable() && !targetChannel.config().isAutoRead()) {
                targetChannel.config().setAutoRead(true);
            }
        });
        super.channelWritabilityChanged(ctx);
    }


    public void offline(Integer port) {
        if (port <= 0) {
            return;
        }
        List<ConnectInfo> collect = CURRENT_STATUS.stream().filter(v -> port.equals(v.getPort())).collect(Collectors.toList());
        for (ConnectInfo connectInfo : collect) {
            connectInfo.getChannel().close();
        }
        CURRENT_STATUS.removeAll(collect);
        remoteConnectionServer.close();
    }


    @Override
    protected void channelRead0(ChannelHandlerContext ctx, HpMessageData.HpMessage hpMessage) throws Exception {
        if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.REGISTER) {
            //校验用户
            String key = hpMessage.getMetaData().getKey();
            UserConnectInfo userConnectInfo = checkUser(key);
            if (userConnectInfo == null) {
                return;
            }
            if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.TCP) {
                processRegisterTcp(userConnectInfo, key);
            } else if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.UDP) {
                processRegisterUdp(userConnectInfo, key);
            } else if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.TCP_UDP) {
                processRegisterUdp(userConnectInfo, key);
                processRegisterTcp(userConnectInfo, key);
            } else {
                ctx.close();
            }
        } else if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.DISCONNECTED) {
            processDisconnected(hpMessage);
        } else if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.DATA) {
            processData(hpMessage);
        } else if (hpMessage.getType() == HpMessageData.HpMessage.HpMessageType.KEEPALIVE) {
            // 心跳包
        } else {
            log.error("未知类型: " + hpMessage.getType());
            ctx.close();
        }
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        udp_channels.close();
        channels.close();
        try {
            List<ConnectInfo> collect = CURRENT_STATUS.stream().filter(v -> v != null && !v.getChannel().isActive() || v != null && v.getChannel().id().asLongText().equals(ctx.channel().id().asLongText())).collect(Collectors.toList());
            for (ConnectInfo connectInfo : collect) {
                connectInfo.getChannel().close();
            }
            CURRENT_STATUS.removeAll(collect);
        } catch (Throwable e) {
            log.error(e.getMessage(), e);
        }
        remoteConnectionServer.close();
    }

    /**
     * if HpMessage.getType() == HpMessageType.REGISTER
     */
    private void processRegisterTcp(UserConnectInfo userConnectInfo, String key) {
        HpMessageData.HpMessage.MetaData.Builder metaDataBuild = HpMessageData.HpMessage.MetaData.newBuilder();
        try {
            //随机端口
            if (userConnectInfo.getPort() <= 0) {
                userConnectInfo.setPort(NetUtil.getAvailablePort());
            }
            HpServerHandler thisHandler = this;
            remoteConnectionServer.bindTcp(userConnectInfo.getPort(), new ChannelInitializer<SocketChannel>() {
                @Override
                public void initChannel(SocketChannel ch) throws Exception {
                    ch.pipeline().addLast(
                            new FlowHandlerStatistics(userConnectInfo, ch.eventLoop()),
                            new ByteArrayDecoder(),
                            new ByteArrayEncoder(),
                            new RemoteProxyHandler(thisHandler)
                    );
                    channels.add(ch);
                }
            });
            metaDataBuild.setSuccess(true);
            CURRENT_STATUS.add(new ConnectInfo(userConnectInfo.getPort(), userConnectInfo.getDomain(), ctx.channel(), key));
            metaDataBuild.setReason("连接成功，外网TCP地址是:" + IocUtil.getBean(WebConfig.class).getHost() + ":" + userConnectInfo.getPort() + ",外网HTTP地址是：http://" + userConnectInfo.getDomain());
            HttpService.pushStatus(key, "TCP穿透成功");
        } catch (Exception e) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason(e.getMessage());
            HttpService.pushStatus(key, "TCP穿透失败[" + e.getMessage() + "]");
            e.printStackTrace();
        }

        HpMessageData.HpMessage.Builder sendBackMessageBuilder = HpMessageData.HpMessage.newBuilder();
        sendBackMessageBuilder.setType(HpMessageData.HpMessage.HpMessageType.REGISTER_RESULT);
        HpMessageData.HpMessage.MetaData metaData = metaDataBuild.build();
        sendBackMessageBuilder.setMetaData(metaData);
        ctx.writeAndFlush(sendBackMessageBuilder.build());
    }


    /**
     * if HpMessage.getType() == HpMessageType.REGISTER
     */
    private void processRegisterUdp(UserConnectInfo userConnectInfo, String key) {
        /**
         * 查询这个用户是否是合法的，不是合法的直接干掉
         */
        HpMessageData.HpMessage.MetaData.Builder metaDataBuild = HpMessageData.HpMessage.MetaData.newBuilder();
        try {
            //随机端口
            if (userConnectInfo.getPort() <= 0) {
                userConnectInfo.setPort(NetUtil.getAvailablePort());
            }
            HpServerHandler thisHandler = this;
            remoteConnectionServer.bindUdp(userConnectInfo.getPort(), new ChannelInitializer<Channel>() {
                @Override
                public void initChannel(Channel ch) throws Exception {
                    RemoteUdpServerHandler remoteUdpServerHandler = new RemoteUdpServerHandler(thisHandler, remoteConnectionServer);
                    ch.pipeline().addLast(
                            new FlowHandlerStatistics(userConnectInfo, ch.eventLoop()),
                            //添加编码器作用是进行统计，包数据
                            remoteUdpServerHandler
                    );
                    udp_channels.add(ch);
                }
            });
            metaDataBuild.setSuccess(true);
            CURRENT_STATUS.add(new ConnectInfo(userConnectInfo.getPort(), "(udp)", ctx.channel(), key));
            metaDataBuild.setReason("连接成功，外网UDP地址是:" + IocUtil.getBean(WebConfig.class).getHost() + ":" + userConnectInfo.getPort());
            HttpService.pushStatus(key, "UDP穿透成功");
        } catch (Exception e) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason(e.getMessage());
            HttpService.pushStatus(key, "UDP穿透失败[" + e.getMessage() + "]");
            e.printStackTrace();
        }
        HpMessageData.HpMessage.Builder sendBackMessageBuilder = HpMessageData.HpMessage.newBuilder();
        sendBackMessageBuilder.setType(HpMessageData.HpMessage.HpMessageType.REGISTER_RESULT);
        HpMessageData.HpMessage.MetaData metaData = metaDataBuild.build();
        sendBackMessageBuilder.setMetaData(metaData);
        ctx.writeAndFlush(sendBackMessageBuilder.build());
    }

    /**
     * 内网数据返回到公网，这里做数据交换，返回给公网用户
     * if HpMessage.getType() == HpMessageType.DATA
     */
    private void processData(HpMessageData.HpMessage hpMessage) {
        byte[] bytes = hpMessage.getData().toByteArray();
        if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.TCP) {
            channels.stream().filter(channel ->
                    channel.id().asLongText().equals(hpMessage.getMetaData().getChannelId())
            ).findFirst().ifPresent(targetChannel -> {
                ChannelConfig config = getCtx().channel().config();
                if (!targetChannel.isWritable()) {
                    //自己不可写，通道可以读，让通道关闭读
                    //自己可写，通道不可以读，让通道打开读
                    if (config.isAutoRead()) {
                        config.setAutoRead(false);
                    }
                } else {
                    if (!config.isAutoRead()) {
                        config.setAutoRead(true);
                    }
                }
                targetChannel.writeAndFlush(bytes);
            });
        }
        if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.UDP) {
            udp_channels.stream().filter(channel ->
                    channel.id().asLongText().equals(hpMessage.getMetaData().getChannelId())
            ).findFirst().ifPresent(targetChannel -> {
                final Attribute<InetSocketAddress> attr = targetChannel.attr(RemoteUdpServerHandler.SENDER);
                final InetSocketAddress inetSocketAddress = attr.get();
                if (inetSocketAddress != null) {
                    ChannelConfig config = getCtx().channel().config();
                    if (!targetChannel.isWritable()) {
                        //自己不可写，通道可以读，让通道关闭读
                        //自己可写，通道不可以读，让通道打开读
                        if (config.isAutoRead()) {
                            config.setAutoRead(false);
                        }
                    } else {
                        if (!config.isAutoRead()) {
                            config.setAutoRead(true);
                        }
                    }
                    targetChannel.writeAndFlush(new DatagramPacket(Unpooled.wrappedBuffer(bytes), inetSocketAddress));
                }
            });
        }
    }

    /**
     * if HpMessage.getType() == HpMessageType.DISCONNECTED
     *
     * @param hpMessage
     */
    private void processDisconnected(HpMessageData.HpMessage hpMessage) {
        channels.close(channel -> channel.id().asLongText().equals(hpMessage.getMetaData().getChannelId()) || !channel.isActive());
        udp_channels.close(channel -> channel.id().asLongText().equals(hpMessage.getMetaData().getChannelId()) || !channel.isActive());
    }


    /**
     * 检查服务
     *
     * @param key
     * @return
     */
    private UserConnectInfo checkUser(String key) {
        UserConnectInfo login = HttpService.login(key);
        /**
         * 查询这个用户是否是合法的，不是合法的直接干掉
         */
        HpMessageData.HpMessage.MetaData.Builder metaDataBuild = HpMessageData.HpMessage.MetaData.newBuilder();
        if (login == null) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason("非法用户，登录失败，有疑问请联系管理员");
        } else {
            //检查下端口是否已经存在，存在的需要将其关闭
            offline(login.getPort());
            return login;
        }
        HttpService.pushStatus(key, metaDataBuild.getReason());
        HpMessageData.HpMessage.Builder sendBackMessageBuilder = HpMessageData.HpMessage.newBuilder();
        sendBackMessageBuilder.setType(HpMessageData.HpMessage.HpMessageType.REGISTER_RESULT);
        HpMessageData.HpMessage.MetaData metaData = metaDataBuild.build();
        sendBackMessageBuilder.setMetaData(metaData);
        ctx.writeAndFlush(sendBackMessageBuilder.build());
        ctx.close();
        return null;
    }

}
