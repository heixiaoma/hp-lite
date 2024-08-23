package net.hserver.hplite.init;

import cn.hserver.core.interfaces.InitRunner;
import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.server.context.IoMultiplexer;
import cn.hserver.core.server.util.EventLoopUtil;
import cn.hserver.core.server.util.PropUtil;
import io.netty.bootstrap.Bootstrap;
import io.netty.channel.Channel;
import io.netty.channel.ChannelHandler;
import io.netty.channel.EventLoopGroup;
import io.netty.channel.epoll.EpollDatagramChannel;
import io.netty.channel.kqueue.KQueueDatagramChannel;
import io.netty.channel.socket.DatagramChannel;
import io.netty.channel.socket.nio.NioDatagramChannel;
import io.netty.handler.ssl.util.SelfSignedCertificate;
import io.netty.incubator.channel.uring.IOUringDatagramChannel;
import io.netty.incubator.codec.quic.*;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.config.CostConfig;
import net.hserver.hplite.config.TunnelConfig;
import net.hserver.hplite.dao.TableMapper;
import net.hserver.hplite.handler.quic.QuicHandler;

import java.net.InetSocketAddress;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;

import static io.netty.channel.unix.UnixChannelOption.SO_REUSEPORT;

@Bean
@Slf4j
public class QuicServiceInit implements InitRunner {

    @Autowired
    private TableMapper tableMapper;
    @Autowired
    private TunnelConfig tunnelConfig;
    @Override
    public void init(String[] strings) {
        tableMapper.createTableUserStatistics();
        tableMapper.createTableUserConfig();
        tableMapper.createTableUserDevice();
        new Thread(this::startHpQuic).start();
    }

    public void startHpQuic() {
        try {
            //构建ssl
            SelfSignedCertificate selfSignedCertificate = new SelfSignedCertificate();
            QuicSslContext context = QuicSslContextBuilder.forServer(
                            selfSignedCertificate.privateKey(), null, selfSignedCertificate.certificate())
                    .earlyData(true)
                    .applicationProtocols(CostConfig.HP_LITE).build();
            int thread = 50;
            //win 服务器没有端口复用功能
            //linux服务器有端口复用功能
            String os = System.getProperty("os.name").toLowerCase();
            if (os.contains("win")) {
               thread=1;
            }
            EventLoopGroup group = EventLoopUtil.getEventLoop(thread, "hp-quic");
            try {
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

                Bootstrap bs = new Bootstrap();
                bs.group(group)
                        .option(SO_REUSEPORT, true)
                        .channel(datagramChannel)
                        .handler(new QuicCodecDispatcher(){
                            @Override
                            protected void initChannel(Channel channel, int i, QuicConnectionIdGenerator quicConnectionIdGenerator) throws Exception {
                                ChannelHandler codec = new QuicServerCodecBuilder().sslContext(context)
                                        .maxIdleTimeout(20, TimeUnit.SECONDS)
                                        .initialMaxData(999999999)
                                        .initialMaxStreamDataBidirectionalLocal(1000000)
                                        .initialMaxStreamDataBidirectionalRemote(1000000)
                                        .initialMaxStreamsBidirectional(1000000)
                                        .initialMaxStreamsUnidirectional(1000000)
                                        .activeMigration(true)
                                        .handler(new QuicHandler())
                                        .streamHandler(new QuicChannelInitializer(null)).build();
                                channel.pipeline().addLast(codec);
                            }
                        });

                List<Channel> channels = new ArrayList<>();
                for (int i = 0; i < thread; i++) {
                    channels.add(bs.bind(new InetSocketAddress(tunnelConfig.getPort())).sync().channel());
                }
                for (Channel channel : channels) {
                    channel.closeFuture().sync();
                }
            } finally {
                group.shutdownGracefully();
            }
        } catch (Exception e) {
          log.error(e.getMessage(),e);
        }
    }
}
