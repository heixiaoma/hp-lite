package net.hserver.hplite.handler;

import cn.hutool.core.lang.Pair;
import io.netty.channel.ChannelHandlerContext;
import io.netty.handler.traffic.GlobalTrafficShapingHandler;
import io.netty.handler.traffic.TrafficCounter;
import io.netty.util.concurrent.EventExecutor;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.utils.StatisticsUtil;

/**
 * 套餐校验，针对流量，网速，限速实现
 */
@Slf4j
public class FlowHandlerStatistics extends GlobalTrafficShapingHandler {


    private final UserConnectInfo userConnectInfo;


    //有限流方式
    public FlowHandlerStatistics(UserConnectInfo userConnectInfo, EventExecutor executor) {
        super(executor);
        StatisticsUtil.init(userConnectInfo);
        this.userConnectInfo = userConnectInfo;
    }

    @Override
    public void channelActive(ChannelHandlerContext ctx) throws Exception {
        StatisticsUtil.addUvPv(userConnectInfo,ctx);
        super.channelActive(ctx);
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        StatisticsUtil.removeUvPv(userConnectInfo,ctx);
        super.channelInactive(ctx);
    }


    @Override
    protected void doAccounting(TrafficCounter counter) {
        Pair<Integer, Integer> uvPv = StatisticsUtil.getUvPv(userConnectInfo);
        StatisticsUtil.addData(userConnectInfo, counter.lastWrittenBytes(), counter.lastReadBytes(),uvPv.getKey(),uvPv.getValue() );
    }


}
