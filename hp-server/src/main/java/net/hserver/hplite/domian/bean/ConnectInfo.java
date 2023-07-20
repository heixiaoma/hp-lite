package net.hserver.hplite.domian.bean;

import io.netty.channel.Channel;
import lombok.Data;


@Data
public class ConnectInfo {
    /**
     * 用户自定义域名
     */
    private String customDomain;

    /**
     * 链接通道
     */
    private Channel channel;

    /**
     * 来源端口
     */
    private Integer port;

    /**
     * 来源IP
     */
    private String ip;

    /**
     * configKey
     */
    private String key;

    public ConnectInfo(Integer port, String customDomain, Channel channel, String key) {
        this.port = port;
        this.customDomain = customDomain;
        this.channel = channel;
        this.ip = channel.remoteAddress().toString();
        this.key=key;
    }

}
