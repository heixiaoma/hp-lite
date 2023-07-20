package net.hserver.hplite.protocol;

import cn.hserver.core.interfaces.ProtocolDispatcherAdapter;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Order;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelPipeline;
import io.netty.handler.timeout.IdleStateHandler;
import net.hserver.hplite.codec.CmdMessageDecoder;
import net.hserver.hplite.codec.CmdMessageEncoder;
import net.hserver.hplite.handler.CmdServerHandler;

import java.net.InetSocketAddress;

@Order(6)
@Bean
public class CmdProtocolDispatcher implements ProtocolDispatcherAdapter {

    @Override
    public boolean dispatcher(ChannelHandlerContext ctx, ChannelPipeline channelPipeline, byte[] bytes) {
        InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().localAddress();
        if (socketAddress.getPort() == 6666) {
            channelPipeline.addLast(new IdleStateHandler(60, 30, 0));
            channelPipeline.addLast(new CmdMessageEncoder());
            channelPipeline.addLast(new CmdMessageDecoder());
            channelPipeline.addLast(new CmdServerHandler());
            return true;
        }
        return false;
    }
}
