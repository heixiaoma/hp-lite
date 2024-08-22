package net.hserver.hplite.message;

import lombok.Data;
import lombok.NoArgsConstructor;
import net.hserver.hplite.domian.entity.UserConfigEntity;

@Data
@NoArgsConstructor
public class UserConnectInfo {
    private Integer id;

    /**
     * proxyVersion不为空就开启了真实IP解析，只能取值V1和V2
     */
    private String proxyVersion;
    private String proxyIp;
    private Integer proxyPort;

    private Integer configId;
    //自定义的域名
    private String domain;

    //开通的端口号 -1随机
    private Integer port;


    /**
     * SSL证书Key
     */
    private String certificateKey;

    /**
     * 证书内容
     */
    private String certificateContent;

}
