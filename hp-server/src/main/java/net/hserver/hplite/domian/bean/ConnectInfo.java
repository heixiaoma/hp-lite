package net.hserver.hplite.domian.bean;

import cn.hutool.core.util.StrUtil;
import io.netty.channel.Channel;
import io.netty.channel.ChannelId;
import io.netty.handler.ssl.SslContext;
import io.netty.incubator.codec.quic.QuicChannel;
import lombok.Data;
import net.hserver.hplite.handler.TunnelServer;
import net.hserver.hplite.message.UserConnectInfo;
import net.hserver.hplite.utils.DateUtil;
import net.hserver.hplite.utils.SSLUtil;

import java.util.Date;


@Data
public class ConnectInfo {

    /**
     * proxyVersion不为空就开启了真实IP解析，只能取值V1和V2
     */
    private String proxyVersion;
    private String proxyIp;
    private Integer proxyPort;
    /**
     * 自定义域名前缀
     */
    private String domain;

    /**
     * 链接通道
     */
    private ChannelId channelId;

    /**
     * 来源端口
     */
    private Integer port;

    /**
     * 来源IP
     */
    private String ip;

    private String date;

    /**
     * configKey
     */
    private String key;

    private SslContext sslContext;

    private QuicChannel quicChannel;

    private TunnelServer tunnelServer;

    //UDP
    public ConnectInfo(QuicChannel quicChannel,TunnelServer tunnelServer, UserConnectInfo connectInfo, String domain, ChannelId channelId, String key) {
        this.quicChannel=quicChannel;
        this.port = connectInfo.getPort();
        this.domain = domain;
        this.channelId = channelId;
        this.date = DateUtil.dateToStamp(new Date());
        this.key = key;
        this.tunnelServer = tunnelServer;

        this.proxyVersion = connectInfo.getProxyVersion();
        this.proxyIp = connectInfo.getProxyIp();
        this.proxyPort = connectInfo.getProxyPort();
    }

    //TCP
    public ConnectInfo(QuicChannel quicChannel, TunnelServer tunnelServer, UserConnectInfo userConnectInfo, ChannelId channelId, String key) {
        this.quicChannel=quicChannel;
        this.port = userConnectInfo.getPort();
        this.domain = userConnectInfo.getDomain();
        this.channelId = channelId;
        this.date = DateUtil.dateToStamp(new Date());
        this.key = key;
        this.tunnelServer = tunnelServer;
        this.proxyVersion = userConnectInfo.getProxyVersion();
        this.proxyIp = userConnectInfo.getProxyIp();
        this.proxyPort = userConnectInfo.getProxyPort();
        if (StrUtil.isNotEmpty(userConnectInfo.getCertificateKey()) && StrUtil.isNotEmpty(userConnectInfo.getCertificateContent())) {
            this.sslContext = SSLUtil.buildSSLContext(userConnectInfo.getCertificateKey(), userConnectInfo.getCertificateContent(), null);
        }
    }
}
