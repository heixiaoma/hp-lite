package net.hserver.hplite.handler;

import cn.hserver.core.server.context.ConstConfig;
import cn.hserver.core.server.context.IoMultiplexer;
import cn.hserver.core.server.util.EventLoopUtil;
import io.netty.bootstrap.Bootstrap;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.ChannelOption;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.epoll.EpollDatagramChannel;
import io.netty.channel.kqueue.KQueueDatagramChannel;
import io.netty.channel.socket.DatagramChannel;
import io.netty.channel.socket.nio.NioDatagramChannel;
import io.netty.incubator.channel.uring.IOUringDatagramChannel;
import lombok.extern.slf4j.Slf4j;


/**
 * @author hxm
 */

@Slf4j
public class TunnelServer {
    private Boolean hasTcp;
    private Channel tcpChannel;
    private Channel udpChannel;
    private final static EventLoopGroup BossGroup = EventLoopUtil.getEventLoop(32, "Boss-TunnelServer");
    private final static EventLoopGroup WorkerGroup = EventLoopUtil.getEventLoop(128, "Worker-TunnelServer");

    public synchronized void bindTcp(int port, ChannelInitializer<?> channelInitializer) throws Exception {
        hasTcp=true;
        ServerBootstrap b = new ServerBootstrap();
        b.group(BossGroup, WorkerGroup)
                .channel(EventLoopUtil.getEventLoopTypeClass())
                .childHandler(channelInitializer)
                .childOption(ChannelOption.AUTO_READ, false)
                .childOption(ChannelOption.TCP_NODELAY, true)
                .childOption(ChannelOption.SO_KEEPALIVE, true);
        b.option(ChannelOption.SO_BACKLOG, ConstConfig.backLog);
        tcpChannel = b.bind(port).sync().channel();
        log.info(port + ":完成");
    }


    public synchronized void bindUdp(int port, ChannelInitializer<?> channelInitializer) throws Exception {
        hasTcp=false;
        IoMultiplexer eventLoopType = EventLoopUtil.getEventLoopType();
        Class<? extends DatagramChannel> datagramChannel;
        switch (eventLoopType) {
            case KQUEUE:
                datagramChannel = KQueueDatagramChannel.class;
                break;
            case EPOLL:
                datagramChannel = EpollDatagramChannel.class;
                break;
            case IO_URING:
                datagramChannel = IOUringDatagramChannel.class;
                break;
            default:
                datagramChannel = NioDatagramChannel.class;
                break;
        }

        Bootstrap b = new Bootstrap();
        b.group(WorkerGroup)
                .channel(datagramChannel)
                .option(ChannelOption.SO_BROADCAST, true)
                .option(ChannelOption.SO_REUSEADDR, true)
                .handler(channelInitializer);
        udpChannel = b.bind(port).sync().channel();
    }

    public Boolean getHasTcp() {
        return hasTcp;
    }

    public synchronized void close() {
        if (udpChannel != null) {
            udpChannel.close();
        }
        if (tcpChannel != null) {
            tcpChannel.close();
        }
//        try {
////            if (udpWorkerGroup != null) {
////                udpWorkerGroup.shutdownGracefully();
////            }
////            if (tcpBossGroup != null) {
////                tcpBossGroup.shutdownGracefully();
////            }
////            if (tcpWorkerGroup != null) {
////                tcpWorkerGroup.shutdownGracefully();
////            }
//        } catch (Exception e) {
//            log.error(e.getMessage(), e);
//        }
    }

}
