package net.hserver.hplite.utils;

import cn.hutool.core.lang.Pair;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioDatagramChannel;
import lombok.extern.slf4j.Slf4j;
import net.hserver.hplite.message.DataStatistics;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.message.UserStatistics;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

/**
 * 统计
 */
@Slf4j
public class StatisticsUtil {
    private static final Map<Integer, UserStatistics> data = new ConcurrentHashMap<>();

    /**
     * 初始化对象
     *
     * @param userConnectInfo
     */
    public static void init(UserConnectInfo userConnectInfo) {
        if (!data.containsKey(userConnectInfo.getConfigId())) {
            UserStatistics userStatistics = new UserStatistics();
            userStatistics.setConfigId(userConnectInfo.getConfigId());
            data.put(userConnectInfo.getConfigId(), userStatistics);
        }

    }

    /**
     * 添加数据
     */
    public static void addData(UserConnectInfo userConnectInfo, long download, long upload) {
        UserStatistics userStatistics = data.get(userConnectInfo.getConfigId());
        if (userStatistics != null) {
            if (download > 0) {
                userStatistics.getDownload().add(download);
            }
            if (upload > 0) {
                userStatistics.getUpload().add(upload);
            }
        }
    }

    public static void addUvPv(UserConnectInfo userConnectInfo, ChannelHandlerContext ctx) {
        UserStatistics userStatistics = data.get(userConnectInfo.getConfigId());
        if (userStatistics != null) {
            userStatistics.getPv().add(1);
        }
    }

    /**
     * 获取有效的数据统计
     *
     * @return
     */
    public static List<DataStatistics> getStatistics() {
        return data.values().stream().filter(k -> k.getDownload().longValue() > 0 || k.getUpload().longValue() > 0).map(k -> {
            DataStatistics dataStatistics = new DataStatistics(k);
            k.rest();
            return dataStatistics;
        }).collect(Collectors.toList());
    }
}
