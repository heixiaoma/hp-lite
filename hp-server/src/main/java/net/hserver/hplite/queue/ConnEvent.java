package net.hserver.hplite.queue;


import cn.hserver.core.ioc.annotation.queue.QueueHandler;
import cn.hserver.core.ioc.annotation.queue.QueueListener;
import cn.hserver.core.server.util.PropUtil;
import io.netty.channel.Channel;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelId;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.socket.SocketChannel;
import io.netty.incubator.codec.quic.QuicChannel;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.ConnectInfo;
import net.hserver.hplite.handler.FlowHandlerStatistics;
import net.hserver.hplite.handler.RemoteProxyHandler;
import net.hserver.hplite.handler.RemoteUdpServerHandler;
import net.hserver.hplite.handler.TunnelServer;
import net.hserver.hplite.handler.quic.QuicHandler;
import net.hserver.hplite.handler.quic.QuicStreamHandler;
import net.hserver.hplite.handler.quic.QuicStreamSuperHandler;
import net.hserver.hplite.message.HpMessageData;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.service.HttpService;
import net.hserver.hplite.utils.NetUtil;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@QueueListener(queueName = QueueEvent.CONN_EVENT)
public class ConnEvent {

    @QueueHandler
    public void connEvent(HpMessageData.HpMessage hpMessage, String id) {
        String key = hpMessage.getMetaData().getKey();
        QuicStreamHandler handler = getHandler(key);
        if (handler == null) {
            List<Channel> collect = QuicStreamHandler.DATA_CON_CH.stream().filter(x -> x.id().asLongText().equals(id)).collect(Collectors.toList());
            for (Channel channel : collect) {
                channel.close();
                log.warn("不存在KEY：{}，关闭:{}", key, id);
            }
            log.error("不存在KEY：{}", key);
            return;
        }
        try {
            //校验用户
            UserConnectInfo userConnectInfo = checkUser(key, handler);
            if (userConnectInfo == null) {
                return;
            }
            if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.TCP) {
                processRegisterTcp(userConnectInfo, key, handler);
            } else if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.UDP) {
                processRegisterUdp(userConnectInfo, key, handler);
            } else if (hpMessage.getMetaData().getType() == HpMessageData.HpMessage.MessageType.TCP_UDP) {
                processRegisterUdp(userConnectInfo, key, handler);
                processRegisterTcp(userConnectInfo, key, handler);
            } else {
                handler.channelHandlerContext.close();
            }
        } catch (Exception e) {
            handler.channelHandlerContext.close();
            log.error(e.getMessage(), e);
        } finally {
            QuicStreamHandler.DATA_CON.remove(key);
        }
    }


    private QuicStreamHandler getHandler(String key) {
        return QuicStreamHandler.DATA_CON.get(key);
    }

    private UserConnectInfo checkUser(String key, QuicStreamHandler handler) {
        UserConnectInfo login = null;
        String error = null;
        try {
            login = HttpService.login(key);
        } catch (Exception e) {
            error = e.getMessage();
        }
        /**
         * 查询这个用户是否是合法的，不是合法的直接干掉
         */
        HpMessageData.HpMessage.MetaData.Builder metaDataBuild = HpMessageData.HpMessage.MetaData.newBuilder();
        if (error != null) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason("检查失败：" + error);
        } else if (login == null) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason("非法用户，登录失败，有疑问请联系管理员");
        }  else {
            //检查下端口是否已经存在，存在的需要将其关闭
            offline(login.getPort(), key, handler);
            return login;
        }
        HttpService.pushStatus(key, metaDataBuild.getReason());
        HpMessageData.HpMessage.Builder sendBackMessageBuilder = HpMessageData.HpMessage.newBuilder();
        sendBackMessageBuilder.setType(HpMessageData.HpMessage.HpMessageType.REGISTER_RESULT);
        HpMessageData.HpMessage.MetaData metaData = metaDataBuild.build();
        sendBackMessageBuilder.setMetaData(metaData);
        handler.sendMessage(sendBackMessageBuilder.build(),ChannelFutureListener.CLOSE);
        return null;
    }


    private void processRegisterTcp(UserConnectInfo userConnectInfo, String key, QuicStreamHandler handler) {
        TunnelServer remoteConnectionServer = null;
        HpMessageData.HpMessage.MetaData.Builder metaDataBuild = HpMessageData.HpMessage.MetaData.newBuilder();
        try {
            //随机端口
            if (userConnectInfo.getPort() <= 0) {
                userConnectInfo.setPort(NetUtil.getAvailablePort());
            }
            remoteConnectionServer = new TunnelServer();
            ConnectInfo connectInfo = new ConnectInfo(handler.getSuperChannel(),remoteConnectionServer, userConnectInfo, handler.getSuperChannelId(), key);
            remoteConnectionServer.bindTcp(userConnectInfo.getPort(), new ChannelInitializer<SocketChannel>() {
                @Override
                public void initChannel(SocketChannel ch) throws Exception {
                    ch.pipeline().addLast(
                            new FlowHandlerStatistics(userConnectInfo),
                            new RemoteProxyHandler(handler,userConnectInfo)
                    );
                }
            });
            metaDataBuild.setSuccess(true);
            handler.addConnectInfo(connectInfo);
            metaDataBuild.setReason("连接成功，外网TCP端口是:" + userConnectInfo.getPort() + ",外网HTTP地址是：http(s)://" + userConnectInfo.getDomain());
            HttpService.pushStatus(key, "TCP映射成功");
        } catch (Exception e) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason(e.getMessage());
            HttpService.pushStatus(key, "TCP映射失败[" + e.getMessage() + "]");
            if (remoteConnectionServer != null) {
                remoteConnectionServer.close();
            }
            log.error(e.getMessage(),e);
        }
        HpMessageData.HpMessage.Builder sendBackMessageBuilder = HpMessageData.HpMessage.newBuilder();
        sendBackMessageBuilder.setType(HpMessageData.HpMessage.HpMessageType.REGISTER_RESULT);
        HpMessageData.HpMessage.MetaData metaData = metaDataBuild.build();
        sendBackMessageBuilder.setMetaData(metaData);
        try {
            if (metaData.getSuccess()) {
                handler.sendMessage(sendBackMessageBuilder.build(),null);
            } else {
                handler.sendMessage(sendBackMessageBuilder.build(),ChannelFutureListener.CLOSE);
            }
        } catch (Exception e) {
            if (remoteConnectionServer != null) {
                remoteConnectionServer.close();
            }
            throw e;
        }
    }

    /**
     * if HpMessage.getType() == HpMessageType.REGISTER
     */
    private void processRegisterUdp(UserConnectInfo userConnectInfo, String key, QuicStreamHandler handler) {
        /**
         * 查询这个用户是否是合法的，不是合法的直接干掉
         */
        TunnelServer remoteConnectionServer = null;
        HpMessageData.HpMessage.MetaData.Builder metaDataBuild = HpMessageData.HpMessage.MetaData.newBuilder();
        try {
            //随机端口
            if (userConnectInfo.getPort() <= 0) {
                userConnectInfo.setPort(NetUtil.getAvailablePort());
            }
            remoteConnectionServer = new TunnelServer();
            ConnectInfo connectInfo = new ConnectInfo(handler.getSuperChannel(),remoteConnectionServer, userConnectInfo, "(udp)", handler.getSuperChannelId(), key);
            remoteConnectionServer.bindUdp(userConnectInfo.getPort(), new ChannelInitializer<Channel>() {
                @Override
                public void initChannel(Channel ch) throws Exception {
                    ch.config().setAutoRead(false);
                    RemoteUdpServerHandler remoteUdpServerHandler = new RemoteUdpServerHandler(handler,userConnectInfo);
                    ch.pipeline().addLast(
                            new FlowHandlerStatistics(userConnectInfo),
                            //添加编码器作用是进行统计，包数据
                            remoteUdpServerHandler
                    );
                }
            });
            metaDataBuild.setSuccess(true);
            handler.addConnectInfo(connectInfo);
            metaDataBuild.setReason("连接成功，外网UDP端口是:" +userConnectInfo.getPort());
            HttpService.pushStatus(key, "UDP映射成功");
        } catch (Exception e) {
            metaDataBuild.setSuccess(false);
            metaDataBuild.setReason(e.getMessage());
            HttpService.pushStatus(key, "UDP映射失败[" + e.getMessage() + "]");
            if (remoteConnectionServer != null) {
                remoteConnectionServer.close();
            }
            log.error(e.getMessage(),e);
        }
        HpMessageData.HpMessage.Builder sendBackMessageBuilder = HpMessageData.HpMessage.newBuilder();
        sendBackMessageBuilder.setType(HpMessageData.HpMessage.HpMessageType.REGISTER_RESULT);
        HpMessageData.HpMessage.MetaData metaData = metaDataBuild.build();
        sendBackMessageBuilder.setMetaData(metaData);
        try {
            if (metaData.getSuccess()) {
                handler.sendMessage(sendBackMessageBuilder.build(),null);
            } else {
                handler.sendMessage(sendBackMessageBuilder.build(),ChannelFutureListener.CLOSE);
            }
        } catch (Exception e) {
            if (remoteConnectionServer != null) {
                remoteConnectionServer.close();
            }
            throw e;
        }
    }


    public void offline(Integer port, String key, QuicStreamHandler handler) {

        try {
            List<ConnectInfo> collect = new ArrayList<>();
            if (port <= 0) {
                collect.addAll(QuicStreamSuperHandler.getByKey(key));
            } else {
                collect.addAll(QuicStreamSuperHandler.getByPort(port));
            }
            for (ConnectInfo connectInfo : collect) {
                connectInfo.getTunnelServer().close();
                ChannelId channelId = connectInfo.getChannelId();
                QuicChannel quicChannel = handler.getQuicChannel(channelId);
                if (quicChannel != null) {
                    quicChannel.close();
                }
            }
            QuicHandler.CURRENT_STATUS.removeAll(collect);
        } catch (Exception e) {
            log.error(e.getMessage(), e);
        }
    }
}
