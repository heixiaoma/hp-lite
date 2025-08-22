package net.hserver.hplite.init;

import io.netty.channel.Channel;
import io.netty.channel.ChannelInitializer;
import io.netty.handler.flow.FlowControlHandler;
import io.netty.incubator.codec.quic.QuicStreamChannel;
import net.hserver.hplite.codec.HpMessageDecoder;
import net.hserver.hplite.codec.HpMessageEncoder;
import net.hserver.hplite.handler.quic.QuicStreamHandler;

public class QuicChannelInitializer extends ChannelInitializer<QuicStreamChannel> {

    private final Channel w;

    public QuicChannelInitializer(Channel w) {
        this.w = w;
    }

    @Override
    protected void initChannel(QuicStreamChannel quicStreamChannel) throws Exception {
        quicStreamChannel.config().setAutoRead(false);
        quicStreamChannel.pipeline()
                .addLast(new FlowControlHandler())
                .addLast(new HpMessageDecoder(), new HpMessageEncoder())
                .addLast(new QuicStreamHandler(w));
    }
}
