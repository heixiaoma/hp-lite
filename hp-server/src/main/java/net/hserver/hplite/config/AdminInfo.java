package net.hserver.hplite.config;

import cn.hserver.core.ioc.annotation.ConfigurationProperties;
import lombok.Data;

@ConfigurationProperties(prefix = "admin")
@Data
public class AdminInfo {

    private String username;

    private String password;

}
