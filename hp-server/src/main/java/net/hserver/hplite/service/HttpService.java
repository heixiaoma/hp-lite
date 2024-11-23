package net.hserver.hplite.service;

import cn.hserver.core.ioc.IocUtil;
import net.hserver.hplite.domian.entity.UserConfigEntity;
import net.hserver.hplite.message.UserConnectInfo;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class HttpService {
    private static final Logger log = LoggerFactory.getLogger(HttpService.class);

    public static UserConnectInfo login(String configKey) {
        try {
            UserConfigService userConfigService = IocUtil.getBean(UserConfigService.class);
            UserConfigEntity config = userConfigService.getConfig(configKey);
            UserConnectInfo userConnectInfo = new UserConnectInfo();
            userConnectInfo.setProxyPort(config.getLocalPort());
            userConnectInfo.setProxyIp(config.getLocalIp());
            userConnectInfo.setCertificateContent(config.getCertificateContent());
            userConnectInfo.setCertificateKey(config.getCertificateKey());
            userConnectInfo.setConfigId(config.getId());
            userConnectInfo.setIp(config.getServerIp());
            userConnectInfo.setDomain(config.getDomain());
            userConnectInfo.setPort(config.getPort());
            userConnectInfo.setProxyVersion(config.getProxyVersion().name());
            return userConnectInfo;
        } catch (Exception e) {
            log.error(e.getMessage(), e);
        }
        return null;
    }


    public static void pushStatus(String configKey, String msg) {
        try {
            log.info("推送消息：{},{}", configKey, msg);
            UserConfigService userConfigService = IocUtil.getBean(UserConfigService.class);
            userConfigService.pushStatus(configKey,msg);

        } catch (Exception e) {
            log.error(e.getMessage(), e);
        }
    }


}
