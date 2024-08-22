package net.hserver.hplite.handler.cmd;

import cn.hserver.core.queue.HServerQueue;
import cn.hserver.core.server.util.ExceptionUtil;
import cn.hutool.core.util.StrUtil;
import cn.hutool.json.JSONUtil;
import com.google.common.cache.Cache;
import com.google.common.cache.CacheBuilder;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.timeout.IdleState;
import io.netty.handler.timeout.IdleStateEvent;
import io.netty.util.AttributeKey;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.domian.bean.MemoryInfo;
import net.hserver.hplite.domian.bean.OnlineInfo;
import net.hserver.hplite.message.CmdMessageData;
import net.hserver.hplite.message.LocalInnerWear;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;

/**
 * @author hxm
 */
@Slf4j
public class CmdServerHandler extends SimpleChannelInboundHandler<CmdMessageData.CmdMessage> {

    private final static Cache<String, String> CACHE = CacheBuilder.newBuilder().expireAfterAccess(4, TimeUnit.SECONDS).build();

    /**
     * 将key绑定在通道上，方便后期移除时直接查询执行删除map
     */
    private final static AttributeKey<String> USER_KEY = AttributeKey.valueOf("USER_KEY");
    /**
     * 有序跳表
     */
    public final static Map<String, OnlineInfo> ONLINE = new ConcurrentHashMap<>();

    public static OnlineInfo getOnlineKey(String key) {
        return ONLINE.get(key);
    }


    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) throws Exception {
        cause.printStackTrace();
        if (!(cause instanceof IOException)) {
            cause.printStackTrace();
            log.error("CMD通道 {}......\n{}", cause.getMessage(), ExceptionUtil.getMessage(cause));
        }
        ctx.close();
    }

    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof IdleStateEvent) {
            IdleStateEvent e = (IdleStateEvent) evt;
            //如果数据堆积情况，不能关闭连接，
            if (e.state() == IdleState.READER_IDLE) {
                log.info("中心节点-心跳数据");
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
        InetSocketAddress socketAddress = (InetSocketAddress)channelHandlerContext.channel().remoteAddress();
        log.info("消息类型:{},消息版本：{},设备key:{},ip:{}", cmdMessage.getType().name(), cmdMessage.getVersion(), cmdMessage.getKey(),socketAddress);
        switch (cmdMessage.getType()) {
            case CONNECT:
                if (!checkVersion(cmdMessage, channelHandlerContext)) {
                    return;
                }
                //防从重
                String ifPresent = CACHE.getIfPresent(cmdMessage.getKey());
                if (ifPresent != null) {
                    channelHandlerContext.channel().close();
                    return;
                } else {
                    CACHE.put(cmdMessage.getKey(), channelHandlerContext.channel().remoteAddress().toString());
                }
                connect(channelHandlerContext, cmdMessage.getKey(), cmdMessage.getData());
                break;
            case DISCONNECT:
                disConnect(channelHandlerContext, cmdMessage.getKey(), cmdMessage.getData());
                break;
            case TIPS:
                tipsMsg(cmdMessage);
                break;
            default:
                channelHandlerContext.close();
        }
    }

    private void tipsMsg(CmdMessageData.CmdMessage cmdMessage) {
        log.info("用户key：{}  的心跳数据", cmdMessage.getKey());
        if (StrUtil.isNotEmpty(cmdMessage.getData())) {
            MemoryInfo memoryInfo = JSONUtil.toBean(cmdMessage.getData(), MemoryInfo.class);
            if (memoryInfo != null) {
                OnlineInfo onlineInfo = ONLINE.get(cmdMessage.getKey());
                onlineInfo.setMemoryInfo(memoryInfo);
            }
        }

    }


    public boolean checkVersion(CmdMessageData.CmdMessage cmdMessage, ChannelHandlerContext ctx) {
        if (!cmdMessage.getVersion().equals("hp-pro:5.0")) {
            CmdMessageData.CmdMessage keepMessage = CmdMessageData.CmdMessage.newBuilder()
                    .setData("当前HP版本太老了，你需要升级后才能帮助你完成更好的服务")
                    .setType(CmdMessageData.CmdMessage.CmdMessageType.TIPS).build();
            ctx.writeAndFlush(keepMessage);
        }
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
        if (key != null) {
            OnlineInfo onlineInfo = ONLINE.get(key);
            if (onlineInfo != null) {
                ChannelHandlerContext channelHandlerContext = onlineInfo.getCtx();
                if (channelHandlerContext != null) {
                    channelHandlerContext.close();
                    ONLINE.remove(key);
                    log.error("移除了{},通道{}", key, ctx.channel().id().asLongText());
                }
            }
        }
        super.channelInactive(ctx);
    }

    /**
     * 下发映射数据，客服端进行映射连接
     *
     * @param key
     * @param data
     * @return
     */
    public static boolean send(String key, List<LocalInnerWear> data) {
        if (StrUtil.isEmpty(key)) {
            return false;
        }
        OnlineInfo onlineInfo = ONLINE.get(key);
        if (onlineInfo == null) {
            return false;
        }
        ChannelHandlerContext channelHandlerContext = onlineInfo.getCtx();
        if (channelHandlerContext == null) {
            return false;
        }
        CmdMessageData.CmdMessage LOCAL_INNER_WEAR = CmdMessageData.CmdMessage.newBuilder()
                .setData(JSONUtil.toJsonStr(data))
                .setType(CmdMessageData.CmdMessage.CmdMessageType.LOCAL_INNER_WEAR).build();
        channelHandlerContext.writeAndFlush(LOCAL_INNER_WEAR);
        return true;
    }

    public static boolean sendTipsMsg(String key, String message) {
        if (StrUtil.isEmpty(key) || StrUtil.isEmpty(message)) {
            return false;
        }
        OnlineInfo onlineInfo = ONLINE.get(key);
        if (onlineInfo == null) {
            return false;
        }
        ChannelHandlerContext channelHandlerContext = onlineInfo.getCtx();
        if (channelHandlerContext == null) {
            return false;
        }
        CmdMessageData.CmdMessage keepMessage = CmdMessageData.CmdMessage.newBuilder()
                .setData(message)
                .setType(CmdMessageData.CmdMessage.CmdMessageType.TIPS).build();
        channelHandlerContext.writeAndFlush(keepMessage);
        return true;
    }

    public static boolean sendCloseMsg(String key, String message) {
        OnlineInfo onlineInfo = ONLINE.get(key);
        if (onlineInfo == null) {
            return false;
        }
        ChannelHandlerContext channelHandlerContext = onlineInfo.getCtx();
        if (channelHandlerContext == null) {
            return false;
        }
        return closeMsg(channelHandlerContext, message);
    }

    private static boolean closeMsg(ChannelHandlerContext ctx, String msg) {
        CmdMessageData.CmdMessage LOCAL_INNER_WEAR = CmdMessageData.CmdMessage.newBuilder()
                .setData(msg)
                .setType(CmdMessageData.CmdMessage.CmdMessageType.DISCONNECT).build();
        ctx.writeAndFlush(LOCAL_INNER_WEAR);
        ctx.close();
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
        //查询用户key的然后查询是否存在配置，如果是存在就对齐进行下发映射数据，此处进行异步操作哦
        //检查用户key是否已经存在，已经存在的直接关闭调，不能进行链接
        if (ONLINE.containsKey(key)) {
            OnlineInfo onlineInfo = ONLINE.get(key);
            onlineInfo.setCtx(ctx);
            String deviceIp = ((InetSocketAddress) ctx.channel().remoteAddress()).getHostString();
            ctx.channel().attr(USER_KEY).set(key);
            log.warn("设备KEY:{} 已经在线,设备IP:{}",key,deviceIp);
            sendTipsMsg(key, "设备在重新登陆，设备KEY已经在线，请检查是否有其他设备登陆，或者更换设备ID。");
            return;
        }
        ctx.channel().attr(USER_KEY).set(key);
        MemoryInfo memoryInfo = new MemoryInfo();
        try {
            if (StrUtil.isNotEmpty(data)) {
                memoryInfo = JSONUtil.toBean(data, MemoryInfo.class);
            }
        } catch (Exception ignored) {
        }
        ONLINE.put(key, new OnlineInfo(ctx, memoryInfo));
        String deviceIp = ((InetSocketAddress) ctx.channel().remoteAddress()).getHostString();
        //发送异步队列进行检查下发映射数据
        HServerQueue.sendQueue("CONNECT_EVENT", key);
        //发送连接消息
        sendTipsMsg(key, "连接服务成功");
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
