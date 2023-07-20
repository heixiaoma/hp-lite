package net.hserver.hplite.protocol;

import cn.hserver.core.interfaces.ProtocolDispatcherAdapter;
import cn.hserver.core.ioc.annotation.Bean;
import cn.hserver.core.ioc.annotation.Order;
import cn.hserver.core.server.util.PropUtil;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelPipeline;
import io.netty.handler.timeout.IdleStateHandler;
import net.hserver.hplite.codec.HpMessageDecoder;
import net.hserver.hplite.codec.HpMessageEncoder;
import net.hserver.hplite.handler.HpServerHandler;

import java.net.InetSocketAddress;

@Order(6)
@Bean
public class HpProtocolDispatcher implements ProtocolDispatcherAdapter {

    //判断HP头cd
    @Override
    public boolean dispatcher(ChannelHandlerContext ctx, ChannelPipeline channelPipeline, byte[] bytes) {
        InetSocketAddress socketAddress = (InetSocketAddress) ctx.channel().localAddress();
        if (socketAddress.getPort() == 9090) {
            channelPipeline.addLast(new IdleStateHandler(60, 30, 0));
            channelPipeline.addLast(new HpMessageDecoder());
            channelPipeline.addLast(new HpMessageEncoder());
            channelPipeline.addLast(new HpServerHandler());
            return true;
        }
        return false;
    }
}
