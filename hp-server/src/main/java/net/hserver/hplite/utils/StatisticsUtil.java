package net.hserver.hplite.utils;

import cn.hutool.core.lang.Pair;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.socket.SocketChannel;
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
public class StatisticsUtil {

    //pv uv
    private final static Map<Integer, Map<String, Integer>> visitorCountMap = new ConcurrentHashMap<>();

    private static final Map<Integer, UserStatistics> data = new ConcurrentHashMap<>();

    /**
     * 初始化对象
     *
     * @param userConnectInfo
     */
    public static void init(UserConnectInfo userConnectInfo) {
        if (!data.containsKey(userConnectInfo.getConfigId())) {
            UserStatistics userStatistics = new UserStatistics();
            userStatistics.setUserId(userConnectInfo.getId());
            userStatistics.setConfigId(userConnectInfo.getConfigId());
            data.put(userConnectInfo.getConfigId(), userStatistics);
        }

    }


    private static String getIp(ChannelHandlerContext ctx) {
        return ((SocketChannel) ctx.channel()).remoteAddress().getAddress().getHostAddress();
    }

    /**
     * 添加数据
     */
    public static void addData(UserConnectInfo userConnectInfo, long download, long upload, int uv, int pv) {
        UserStatistics userStatistics = data.get(userConnectInfo.getConfigId());
        if (userStatistics != null) {
            userStatistics.getDownload().add(download);
            userStatistics.getUpload().add(upload);
            userStatistics.setUv(uv);
            userStatistics.setPv(pv);

        }
    }

    public static Pair<Integer, Integer> getUvPv(UserConnectInfo userConnectInfo) {
        Map<String, Integer> stringIntegerMap = visitorCountMap.get(userConnectInfo.getConfigId());
        int uv = stringIntegerMap.size();
        int pv = stringIntegerMap.values().stream().mapToInt(Integer::intValue).sum();
        return Pair.of(uv, pv);
    }


    public static void addUvPv(UserConnectInfo userConnectInfo, ChannelHandlerContext ctx) {
        Map<String, Integer> uvPv = visitorCountMap.get(userConnectInfo.getConfigId());
        String ip = getIp(ctx);
        if (uvPv == null) {
            uvPv = new ConcurrentHashMap<>();
            uvPv.put(ip, 1);
            visitorCountMap.put(userConnectInfo.getConfigId(), uvPv);
        } else {
            Integer integer = uvPv.get(ip);
            if (integer == null) {
                integer = 0;
            }
            uvPv.put(ip, integer + 1);
        }
    }

    public static void removeUvPv(UserConnectInfo userConnectInfo, ChannelHandlerContext ctx) {
        Map<String, Integer> uvPv = visitorCountMap.get(userConnectInfo.getConfigId());
        if (uvPv != null) {
            String ip = getIp(ctx);
            uvPv.remove(ip);
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
