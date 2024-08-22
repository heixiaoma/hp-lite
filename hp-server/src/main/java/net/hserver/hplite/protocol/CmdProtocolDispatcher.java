package net.hserver.hplite.protocol;

import cn.hserver.core.interfaces.ProtocolDispatcherAdapter;
import cn.hserver.core.ioc.annotation.Autowired;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Order;
import cn.hserver.core.ioc.annotation.Value;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelPipeline;
import io.netty.handler.timeout.IdleStateHandler;
import net.hserver.hplite.codec.CmdMessageDecoder;
import net.hserver.hplite.codec.CmdMessageEncoder;
import net.hserver.hplite.handler.cmd.CmdServerHandler;

import java.net.InetSocketAddress;

@Order(6)
@Bean
public class CmdProtocolDispatcher implements ProtocolDispatcherAdapter {

    @Value("cmd.port")
    private Integer port;

    @Override
    public boolean dispatcher(ChannelHandlerContext ctx, ChannelPipeline channelPipeline, byte[] bytes) {
        InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().localAddress();
        if (socketAddress.getPort() == port) {
            channelPipeline.addLast(new IdleStateHandler(600, 200, 0));
            channelPipeline.addLast(new CmdMessageEncoder());
            channelPipeline.addLast(new CmdMessageDecoder());
            channelPipeline.addLast(new CmdServerHandler());
            return true;
        }
        return false;
    }
}
