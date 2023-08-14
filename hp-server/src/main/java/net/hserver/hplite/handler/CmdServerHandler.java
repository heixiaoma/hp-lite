package net.hserver.hplite.handler;

import cn.hserver.core.queue.HServerQueue;
import cn.hserver.core.server.util.ExceptionUtil;
import cn.hutool.json.JSONUtil;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.timeout.IdleState;
import io.netty.handler.timeout.IdleStateEvent;
import io.netty.util.AttributeKey;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.message.CmdMessageData;
import net.hserver.hplite.message.LocalInnerWear;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentSkipListMap;

/**
 * @author hxm
 */
@Slf4j
public class CmdServerHandler extends SimpleChannelInboundHandler<CmdMessageData.CmdMessage> {

    /**
     * 将key绑定在通道上，方便后期移除时直接查询执行删除map
     */
    private final static AttributeKey<String> USER_KEY = AttributeKey.valueOf("USER_KEY");

    /**
     * 有序跳表
     */
    public final static Map<String, ChannelHandlerContext> ONLINE = new ConcurrentSkipListMap<>();

    public static boolean hasKey(String key) {
        return ONLINE.containsKey(key);
    }


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        cause.printStackTrace();
        if (!(cause instanceof IOException)) {
            cause.printStackTrace();
            log.error("HP通道 {}......\n{}", cause.getMessage(), ExceptionUtil.getMessage(cause));
        }
        ctx.close();
    }

    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof IdleStateEvent) {
            IdleStateEvent e = (IdleStateEvent) evt;
            //如果数据堆积情况，不能关闭连接，
            if (e.state() == IdleState.READER_IDLE) {
                CmdMessageData.CmdMessage keepMessage = CmdMessageData.CmdMessage.newBuilder()
                        .setData("中心节点-心跳数据")
                        .setType(CmdMessageData.CmdMessage.CmdMessageType.TIPS).build();
                ctx.writeAndFlush(keepMessage);
            } else if (e.state() == IdleState.WRITER_IDLE) {
                log.info("中心节点-心跳数据");
                CmdMessageData.CmdMessage keepMessage = CmdMessageData.CmdMessage.newBuilder()
                        .setData("中心节点-心跳数据")
                        .setType(CmdMessageData.CmdMessage.CmdMessageType.TIPS).build();
                ctx.writeAndFlush(keepMessage);
            }
        }
    }

    @Override
    protected void channelRead0(ChannelHandlerContext channelHandlerContext, CmdMessageData.CmdMessage cmdMessage) throws Exception {
        log.debug("消息类型:{},消息版本：{}", cmdMessage.getType().name(), cmdMessage.getVersion());
        if (!checkVersion(cmdMessage, channelHandlerContext)) {
            return;
        }
        switch (cmdMessage.getType()) {
            case CONNECT:
                connect(channelHandlerContext, cmdMessage.getKey(), cmdMessage.getData());
                break;
            case DISCONNECT:
                disConnect(channelHandlerContext, cmdMessage.getKey(), cmdMessage.getData());
                break;
            case TIPS:
                log.info("用户key：{}  的心跳数据", cmdMessage.getKey());
                break;
            default:
                channelHandlerContext.close();
        }
    }


    public boolean checkVersion(CmdMessageData.CmdMessage cmdMessage, ChannelHandlerContext ctx) {
        if (cmdMessage.getVersion().trim().length() == 0) {
            CmdMessageData.CmdMessage keepMessage = CmdMessageData.CmdMessage.newBuilder()
                    .setData("当前HP版本太老了，你需要升级后才能帮助你完成更好的服务")
                    .setType(CmdMessageData.CmdMessage.CmdMessageType.TIPS).build();
            ctx.writeAndFlush(keepMessage);
            ctx.close();
            return false;
        } else if (cmdMessage.getKey().trim().length() != 32) {
            CmdMessageData.CmdMessage keepMessage = CmdMessageData.CmdMessage.newBuilder()
                    .setData("设备ID不合法")
                    .setType(CmdMessageData.CmdMessage.CmdMessageType.TIPS).build();
            ctx.writeAndFlush(keepMessage);
            ctx.close();
            return false;
        }
        return true;
    }


    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        //无效的连接将其关闭删除
        String key = ctx.channel().attr(USER_KEY).get();
        if (key == null) {
            log.error("移除链接是没有检查到通道的key值{}", ctx.channel().id().asLongText());
        } else {
            ChannelHandlerContext channelHandlerContext = ONLINE.get(key);
            if (channelHandlerContext != null) {
                channelHandlerContext.close();
                ONLINE.remove(key);
                log.error("移除了通道{}", ctx.channel().id().asLongText());
            }
        }
        super.channelInactive(ctx);
    }

    /**
     * 下发穿透数据，客服端进行穿透连接
     *
     * @param key
     * @param data
     * @return
     */
    public static boolean send(String key, List<LocalInnerWear> data) {
        ChannelHandlerContext channelHandlerContext = ONLINE.get(key);
        if (channelHandlerContext == null) {
            return false;
        }
        CmdMessageData.CmdMessage LOCAL_INNER_WEAR = CmdMessageData.CmdMessage.newBuilder()
                .setData(JSONUtil.toJsonStr(data))
                .setType(CmdMessageData.CmdMessage.CmdMessageType.LOCAL_INNER_WEAR).build();
        channelHandlerContext.writeAndFlush(LOCAL_INNER_WEAR);
        return true;
    }

    public static boolean sendCloseMsg(String key, String message) {
        ChannelHandlerContext channelHandlerContext = ONLINE.get(key);
        if (channelHandlerContext == null) {
            return false;
        }
        CmdMessageData.CmdMessage LOCAL_INNER_WEAR = CmdMessageData.CmdMessage.newBuilder()
                .setData(message)
                .setType(CmdMessageData.CmdMessage.CmdMessageType.DISCONNECT).build();
        channelHandlerContext.writeAndFlush(LOCAL_INNER_WEAR);
        channelHandlerContext.close();
        return true;
    }


    /**
     * 系统连接时
     *
     * @param ctx
     * @param key
     * @param data
     */
    public void connect(ChannelHandlerContext ctx, String key, String data) {
        //查询用户key的然后查询是否存在配置，如果是存在就对齐进行下发穿透数据，此处进行异步操作哦
        ctx.channel().attr(USER_KEY).set(key);
        ONLINE.put(key, ctx);
        //发送异步队列进行检查下发穿透数据
        HServerQueue.sendQueue("CONNECT_EVENT", key);
    }

    /**
     * 客服端主动断开连接时
     *
     * @param ctx
     * @param key
     * @param data
     */
    public void disConnect(ChannelHandlerContext ctx, String key, String data) {
        ctx.close();
    }
}
