package net.hserver.hplite.handler;

import io.netty.buffer.ByteBuf;
import io.netty.channel.ChannelDuplexHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.ChannelPromise;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.utils.StatisticsUtil;


/**
 * 套餐校验，针对流量，网速，限速实现
 */
@Slf4j
public class FlowHandlerStatistics extends ChannelDuplexHandler {

    private final UserConnectInfo userConnectInfo;

    //有限流方式
    public FlowHandlerStatistics(UserConnectInfo userConnectInfo) {
        StatisticsUtil.init(userConnectInfo);
        this.userConnectInfo = userConnectInfo;
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        StatisticsUtil.addUvPv(userConnectInfo, ctx);
        super.channelActive(ctx);
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        super.channelInactive(ctx);
    }

    @Override
    public void channelRead(ChannelHandlerContext ctx, Object msg) throws Exception {
        if (msg instanceof ByteBuf) {
            ByteBuf bufData = (ByteBuf) msg;
            StatisticsUtil.addData(userConnectInfo, 0, bufData.readableBytes());
        }
        super.channelRead(ctx, msg);
    }

    @Override
    public void write(ChannelHandlerContext ctx, Object msg, ChannelPromise promise) throws Exception {
        if (msg instanceof ByteBuf) {
            ByteBuf bufData = (ByteBuf) msg;
            StatisticsUtil.addData(userConnectInfo, bufData.readableBytes(), 0);
        }
        super.write(ctx, msg, promise);
    }
}
